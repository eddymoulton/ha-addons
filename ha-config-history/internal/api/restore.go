package api

import (
	"fmt"
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type RestoreBackupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func RestoreBackupHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		group := c.Param("group")
		id := c.Param("id")
		filename := c.Param("filename")

		// Validate path parameters for directory traversal
		if err := SanitizePath(group); err != nil {
			c.JSON(http.StatusBadRequest, RestoreBackupResponse{
				Success: false,
				Error:   "Invalid group parameter",
			})
			return
		}
		if err := SanitizePath(id); err != nil {
			c.JSON(http.StatusBadRequest, RestoreBackupResponse{
				Success: false,
				Error:   "Invalid id parameter",
			})
			return
		}
		if err := SanitizePath(filename); err != nil {
			c.JSON(http.StatusBadRequest, RestoreBackupResponse{
				Success: false,
				Error:   "Invalid filename parameter",
			})
			return
		}

		var configOptions *types.ConfigBackupOptions
		for i := range s.AppSettings.Configs {
			if s.AppSettings.Configs[i].Path == group {
				configOptions = s.AppSettings.Configs[i]
				break
			}
		}

		if configOptions == nil {
			c.JSON(http.StatusNotFound, RestoreBackupResponse{
				Success: false,
				Error:   "Config not found",
			})
			return
		}

		backupContent, err := io.GetConfigBackup(s.AppSettings.BackupDir, group, id, filename)
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

		slog.Info("Backup restored successfully", "group", group, "id", id, "filename", filename, "path", fullPath)

		c.JSON(http.StatusOK, RestoreBackupResponse{
			Success: true,
			Message: fmt.Sprintf("Successfully restored backup to %s", fullPath),
		})
	}
}

func SanitizePath(path string) error {
	if path == "" {
		return fmt.Errorf("invalid path: empty path")
	}

	if filepath.IsAbs(path) {
		return fmt.Errorf("invalid path: absolute paths not allowed")
	}

	if strings.Contains(path, ":") || strings.HasPrefix(path, "\\\\") {
		return fmt.Errorf("invalid path: absolute paths not allowed")
	}

	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: contains directory traversal")
	}

	cleaned := filepath.Clean(path)

	if strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return fmt.Errorf("invalid path: absolute paths not allowed")
	}

	return nil
}
