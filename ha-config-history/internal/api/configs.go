package api

import (
	"ha-config-history/internal/core"
	"ha-config-history/internal/types"
	"maps"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

type ConfigResponse struct {
	Groups map[types.GroupSlug][]*types.BackupConfigSummary `json:"groups"`
}

func GetConfigsHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		s.State.Mu.RLock()
		defer s.State.Mu.RUnlock()

		groups := make(map[types.GroupSlug][]*types.BackupConfigSummary)

		for _, configGroup := range s.AppSettings.ConfigGroups {
			groupConfigs := make([]*types.BackupConfigSummary, 0)
			if groupSummaries, exists := s.State.CachedBackupSummaries[configGroup.Slug]; exists {
				groupConfigs = slices.Collect(maps.Values(groupSummaries))
			}

			groups[configGroup.Slug] = groupConfigs
		}

		c.IndentedJSON(http.StatusOK, ConfigResponse{
			Groups: groups,
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
