package api

import (
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"maps"
	"net/http"
	"slices"
	"sort"

	"github.com/gin-gonic/gin"
)

func GetConfigsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		s.State.Mu.RLock()
		metadata := slices.Collect(maps.Values(s.State.CachedConfigMetadata))
		s.State.Mu.RUnlock()

		sort.Slice(metadata, func(i, j int) bool {
			return metadata[i].FriendlyName < metadata[j].FriendlyName
		})

		if metadata == nil {
			metadata = []*types.ConfigMetadata{}
		}

		c.IndentedJSON(http.StatusOK, metadata)
	}
}

func ProcessConfigsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		s.ProcessAllConfigOptions()
		c.JSON(http.StatusOK, gin.H{
			"status": "backup process completed",
		})
	}
}

func ListConfigBackupsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		group := c.Param("group")
		id := c.Param("id")

		backups, err := io.ListConfigBackups(s.AppSettings.BackupDir, group, id)
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
		group := c.Param("group")
		id := c.Param("id")
		filename := c.Param("filename")

		content, err := io.GetConfigBackup(s.AppSettings.BackupDir, group, id, filename)
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
