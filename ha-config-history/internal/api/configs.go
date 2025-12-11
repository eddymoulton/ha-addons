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

type ConfigResponse struct {
	Configs []*types.ConfigMetadata            `json:"configs"`
	Groups  map[string][]*types.ConfigMetadata `json:"groups"`
}

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

		groups := make(map[string][]*types.ConfigMetadata)

		metadataByPath := make(map[string]*types.ConfigMetadata)
		for _, meta := range metadata {
			metadataByPath[meta.Path] = meta
		}

		for _, configGroup := range s.AppSettings.ConfigGroups {
			groupConfigs := make([]*types.ConfigMetadata, 0)
			for _, config := range configGroup.Configs {
				if meta, exists := metadataByPath[config.Path]; exists {
					groupConfigs = append(groupConfigs, meta)
				}
			}
			if len(groupConfigs) > 0 {
				groups[configGroup.GroupName] = groupConfigs
			}
		}

		c.IndentedJSON(http.StatusOK, ConfigResponse{
			Configs: metadata,
			Groups:  groups,
		})
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
		path := c.Param("path")
		id := c.Param("id")

		backups, err := io.ListConfigBackups(s.AppSettings.BackupDir, path, id)
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
		path := c.Param("path")
		id := c.Param("id")
		filename := c.Param("filename")

		content, err := io.GetConfigBackup(s.AppSettings.BackupDir, path, id, filename)
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
		path := c.Param("path")
		id := c.Param("id")
		filename := c.Param("filename")

		if err := io.SanitizePath(path); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid path parameter",
			})
			return
		}
		if err := io.SanitizePath(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id parameter",
			})
			return
		}
		if err := io.SanitizePath(filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid filename parameter",
			})
			return
		}

		err := io.DeleteBackup(s.AppSettings.BackupDir, path, id, filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		metadata, err := io.UpdateMetadataAfterDeletion(s.AppSettings.BackupDir, path, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if metadata != nil {
			s.State.Mu.Lock()
			s.State.CachedConfigMetadata[types.ConfigIdentifier{Path: path, ID: id}] = metadata
			s.State.Mu.Unlock()
		} else {
			s.State.Mu.Lock()
			delete(s.State.CachedConfigMetadata, types.ConfigIdentifier{Path: path, ID: id})
			s.State.Mu.Unlock()
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "backup deleted successfully",
		})
	}
}

func DeleteAllConfigBackupsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		path := c.Param("path")
		id := c.Param("id")

		if err := io.SanitizePath(path); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid path parameter",
			})
			return
		}
		if err := io.SanitizePath(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id parameter",
			})
			return
		}

		err := io.DeleteAllBackups(s.AppSettings.BackupDir, path, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		s.State.Mu.Lock()
		delete(s.State.CachedConfigMetadata, types.ConfigIdentifier{Path: path, ID: id})
		s.State.Mu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"status": "all backups deleted successfully",
		})
	}
}
