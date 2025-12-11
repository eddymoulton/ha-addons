package api

import (
	"fmt"
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type RestoreBackupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func RestoreBackupHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		groupSlug := types.GroupSlug(c.Param("group"))
		configPath := c.Param("path")
		id := c.Param("id")
		filename := c.Param("filename")

		var configOptions *types.ConfigBackupOptions
		// Search through all config groups to find the config with matching path
		for _, configGroup := range s.AppSettings.ConfigGroups {
			if configGroup.Slug != groupSlug {
				continue
			}

			for _, config := range configGroup.Configs {
				if config.Path == configPath {
					configOptions = config
					break
				}
			}
		}

		if configOptions == nil {
			c.JSON(http.StatusNotFound, RestoreBackupResponse{
				Success: false,
				Error:   "Config not found",
			})
			return
		}

		backupContent, err := io.GetConfigBackup(s.AppSettings.BackupDir, groupSlug, configPath, id, filename)
		if err != nil {
			c.JSON(http.StatusNotFound, RestoreBackupResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to load backup: %v", err),
			})
			return
		}

		fullPath := filepath.Join(s.AppSettings.HomeAssistantConfigDir, configOptions.Path)

		if configOptions.BackupType == "single" {
			if err := io.RestoreEntireFile(fullPath, backupContent); err != nil {
				c.JSON(http.StatusInternalServerError, RestoreBackupResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to restore backup: %v", err),
				})
				return
			}
		}

		if configOptions.BackupType == "multiple" {
			if err := io.RestorePartialFile(fullPath, backupContent, *configOptions); err != nil {
				c.JSON(http.StatusInternalServerError, RestoreBackupResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to restore backup: %v", err),
				})
				return
			}
		}

		if configOptions.BackupType == "directory" {
			fullPath := filepath.Join(s.AppSettings.HomeAssistantConfigDir, configOptions.Path, id)
			if err := io.RestoreEntireFile(fullPath, backupContent); err != nil {
				c.JSON(http.StatusInternalServerError, RestoreBackupResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to restore backup: %v", err),
				})
				return
			}
		}

		slog.Info("Backup restored successfully", "group", groupSlug, "path", configPath, "id", id, "filename", filename, "fullPath", fullPath)

		c.JSON(http.StatusOK, RestoreBackupResponse{
			Success: true,
			Message: fmt.Sprintf("Successfully restored backup to %s", fullPath),
		})
	}
}
