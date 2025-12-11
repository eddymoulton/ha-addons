package api

import (
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListConfigBackupsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		groupSlug := types.GroupSlug(c.Param("group"))
		configPath := c.Param("path")
		id := c.Param("id")

		backups, err := io.ListConfigBackups(s.AppSettings.BackupDir, groupSlug, configPath, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.IndentedJSON(http.StatusOK, backups)
	}
}

func GetConfigBackupHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		groupSlug := types.GroupSlug(c.Param("group"))
		configPath := c.Param("path")
		id := c.Param("id")
		filename := c.Param("filename")

		content, err := io.GetConfigBackup(s.AppSettings.BackupDir, groupSlug, configPath, id, filename)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Header("Content-Type", "application/x-yaml")
		c.String(http.StatusOK, string(content))
	}
}

func DeleteConfigBackupHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		groupSlug := types.GroupSlug(c.Param("group"))
		configPath := c.Param("path")
		id := c.Param("id")
		filename := c.Param("filename")

		err := io.DeleteBackup(s.AppSettings.BackupDir, groupSlug, configPath, id, filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		metadata, err := io.UpdateMetadataAfterDeletion(s.AppSettings.BackupDir, groupSlug, configPath, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if metadata != nil {
			s.State.Mu.Lock()
			s.State.CachedBackupSummaries[groupSlug][types.ConfigBackupIdentifier{Path: configPath, ID: id}] = metadata
			s.State.Mu.Unlock()
		} else {
			s.State.Mu.Lock()
			delete(s.State.CachedBackupSummaries[groupSlug], types.ConfigBackupIdentifier{Path: configPath, ID: id})
			s.State.Mu.Unlock()
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "backup deleted successfully",
		})
	}
}

func DeleteAllConfigBackupsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		groupSlug := types.GroupSlug(c.Param("group"))
		configPath := c.Param("path")
		id := c.Param("id")

		err := io.DeleteAllBackups(s.AppSettings.BackupDir, groupSlug, configPath, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		s.State.Mu.Lock()
		delete(s.State.CachedBackupSummaries[groupSlug], types.ConfigBackupIdentifier{Path: configPath, ID: id})
		s.State.Mu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"status": "all backups deleted successfully",
		})
	}
}
