package api

import (
	"encoding/json"
	"fmt"
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// validateGroupName checks if a group name is valid
func validateGroupName(groupName string) error {
	if strings.TrimSpace(groupName) == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	if len(groupName) > 100 {
		return fmt.Errorf("group name cannot exceed 100 characters")
	}

	// Check for potentially problematic characters
	if match, _ := regexp.MatchString(`[<>:"/\\|?*]`, groupName); match {
		return fmt.Errorf("group name contains invalid characters")
	}

	// Check for reserved names
	reservedNames := []string{"null", "undefined", "admin", "root", "system"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(groupName, reserved) {
			return fmt.Errorf("group name '%s' is reserved", groupName)
		}
	}

	return nil
}

// validateConfigPath checks if a config path is valid
func validateConfigPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("config path cannot be empty")
	}
	if len(path) > 500 {
		return fmt.Errorf("config path cannot exceed 500 characters")
	}

	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("config path cannot contain '..' sequences")
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "/homeassistant") {
		return fmt.Errorf("absolute paths must be within homeassistant directory")
	}

	return nil
}

// validateConfigGroups validates the entire config groups structure
func validateConfigGroups(configGroups []*types.ConfigBackupOptionGroup) error {
	if len(configGroups) == 0 {
		return nil // Empty groups are allowed
	}

	groupNames := make(map[string]bool)
	configPaths := make(map[string]string) // path -> group name for duplicate detection

	for i, group := range configGroups {
		if group == nil {
			return fmt.Errorf("config group at index %d is nil", i)
		}

		if err := validateGroupName(group.Name); err != nil {
			return fmt.Errorf("group at index %d: %v", i, err)
		}

		// Check for duplicate group names
		if groupNames[group.Name] {
			return fmt.Errorf("duplicate group name: '%s'", group.Name)
		}
		groupNames[group.Name] = true

		// Validate configs within group
		if len(group.Configs) == 0 {
			return fmt.Errorf("group '%s' must contain at least one config", group.Name)
		}

		for j, config := range group.Configs {
			if err := validateConfig(config, j, group.Name, configPaths); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateConfig(config *types.ConfigBackupOptions, configIndex int, groupName string, configPaths map[string]string) error {
	if config == nil {
		return fmt.Errorf("config at index %d in group '%s' is nil", configIndex, groupName)
	}

	// Validate config path
	if err := validateConfigPath(config.Path); err != nil {
		return fmt.Errorf("config '%s' in group '%s': %v", config.Path, groupName, err)
	}

	// Validate backup type
	if config.BackupType != "single" && config.BackupType != "multiple" && config.BackupType != "directory" {
		return fmt.Errorf("config '%s' in group '%s' has invalid backup type: '%s'",
			config.Path, groupName, config.BackupType)
	}

	// Validate multiple backup type fields
	if config.BackupType == "multiple" {
		if config.IdNode == nil || strings.TrimSpace(*config.IdNode) == "" {
			return fmt.Errorf("config '%s' with backup type 'multiple' must have a valid idNode", config.Path)
		}
		if config.FriendlyNameNode == nil || strings.TrimSpace(*config.FriendlyNameNode) == "" {
			return fmt.Errorf("config '%s' with backup type 'multiple' must have a valid friendlyNameNode", config.Path)
		}
	}

	// Validate max backups and age constraints
	if config.MaxBackups != nil && *config.MaxBackups < 1 {
		return fmt.Errorf("config '%s' maxBackups must be at least 1", config.Path)
	}

	if config.MaxBackupAgeDays != nil && *config.MaxBackupAgeDays < 1 {
		return fmt.Errorf("config '%s' maxBackupAgeDays must be at least 1", config.Path)
	}

	return nil
}

func GetSettingsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, s.AppSettings)
	}
}

type UpdateSettingsResponse struct {
	Success  bool     `json:"success"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

func UpdateSettingsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		var newSettings types.AppSettings
		if err := c.BindJSON(&newSettings); err != nil {
			c.JSON(http.StatusBadRequest, UpdateSettingsResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid settings format: %v", err),
			})
			return
		}

		var warnings []string

		if !io.DirectoryExists(newSettings.HomeAssistantConfigDir) {
			warnings = append(warnings, fmt.Sprintf("Home Assistant config directory does not exist: %s", newSettings.HomeAssistantConfigDir))
		}
		if !io.DirectoryExists(newSettings.BackupDir) {
			warnings = append(warnings, fmt.Sprintf("Backup directory does not exist: %s", newSettings.BackupDir))
		}

		if newSettings.CronSchedule != nil && *newSettings.CronSchedule != "" {
			if err := core.ValidateCronSchedule(*newSettings.CronSchedule); err != nil {
				c.JSON(http.StatusBadRequest, UpdateSettingsResponse{
					Success: false,
					Error:   fmt.Sprintf("Invalid cron schedule: %v", err),
				})
				return
			}
		}

		if err := validateConfigGroups(newSettings.ConfigGroups); err != nil {
			c.JSON(http.StatusBadRequest, UpdateSettingsResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid config groups: %v", err),
			})
			return
		}

		if len(newSettings.Configs) > 0 && len(newSettings.ConfigGroups) > 0 {
			warnings = append(warnings, "Both old configs format and new config groups detected. Using config groups and ignoring old configs.")
			newSettings.Configs = nil // Clear old format
		}

		configData, err := json.MarshalIndent(newSettings, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, UpdateSettingsResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to serialize settings: %v", err),
			})
			return
		}

		if err := os.WriteFile(s.ConfigPath, configData, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, UpdateSettingsResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to save settings file: %v", err),
			})
			return
		}

		cronChanged := false
		oldSchedule := ""
		newSchedule := ""
		if s.AppSettings.CronSchedule != nil {
			oldSchedule = *s.AppSettings.CronSchedule
		}
		if newSettings.CronSchedule != nil {
			newSchedule = *newSettings.CronSchedule
		}
		if oldSchedule != newSchedule {
			cronChanged = true
		}

		s.AppSettings = &newSettings

		if cronChanged {
			_ = s.RestartCronJob()
			slog.Info("Cron schedule updated", "schedule", newSchedule)
		}

		slog.Info("Settings updated successfully")

		c.JSON(http.StatusOK, UpdateSettingsResponse{
			Success:  true,
			Warnings: warnings,
		})
	}
}
