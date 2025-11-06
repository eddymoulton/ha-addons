package api

import (
	"fmt"
	"ha-config-history/internal/core"
	"ha-config-history/internal/io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

type BackupDiffResponse struct {
	Type          string `json:"type"`
	UnifiedDiff   string `json:"unifiedDiff"`
	Content       string `json:"content"`
	OldContent    string `json:"oldContent"`
	NewContent    string `json:"newContent"`
	OldFilename   string `json:"oldFilename"`
	NewFilename   string `json:"newFilename"`
	IsFirstBackup bool   `json:"isFirstBackup"`
}

func GetBackupDiffHandler(s *core.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		group := c.Param("group")
		id := c.Param("id")
		leftFilename := c.Param("left")
		rightFilename := c.Param("right")

		leftContent, err := io.GetConfigBackup(s.AppSettings.BackupDir, group, id, leftFilename)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error loading left backup file"})
			return
		}

		rightContent, err := io.GetConfigBackup(s.AppSettings.BackupDir, group, id, rightFilename)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error loading right backup file"})
			return
		}

		edits := myers.ComputeEdits(span.URIFromPath(leftFilename), string(leftContent), string(rightContent))
		diff := fmt.Sprint(gotextdiff.ToUnified(leftFilename, rightFilename, string(leftContent), edits))

		c.JSON(http.StatusOK, BackupDiffResponse{
			Type:          "diff",
			UnifiedDiff:   diff,
			OldContent:    string(leftContent),
			NewContent:    string(rightContent),
			OldFilename:   leftFilename,
			NewFilename:   rightFilename,
			IsFirstBackup: false,
		})
	}
}
