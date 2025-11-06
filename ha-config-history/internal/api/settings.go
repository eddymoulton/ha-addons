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

	"github.com/gin-gonic/gin"
)

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

		configData, err := json.MarshalIndent(newSettings, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, UpdateSettingsResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to serialize settings: %v", err),
			})
			return
		}

		if err := os.WriteFile("config.json", configData, 0644); err != nil {
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
