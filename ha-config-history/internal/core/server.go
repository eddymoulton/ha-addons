package core

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log"
	"log/slog"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

// BackupJob represents a backup job to be processed in the queue
type BackupJob struct {
	Options *types.ConfigBackupOptions
	Backup  *types.ConfigBackup
}

type Server struct {
	State       *State
	AppSettings *types.AppSettings
	queue       chan BackupJob
	fileWatcher *fsnotify.Watcher
}

func (s *Server) validateConfig() {
	if !io.DirectoryExists(s.AppSettings.HomeAssistantConfigDir) {
		slog.Error("Home Assistant configuration directory does not exist",
			"dir", s.AppSettings.HomeAssistantConfigDir)
	}

	uniquePaths := make(map[string]struct{})
	// Check configs from new grouped structure
	for _, group := range s.AppSettings.ConfigGroups {
		for _, options := range group.Configs {
			if _, exists := uniquePaths[options.Path]; exists {
				slog.Error("Duplicate config path found in settings",
					"path", options.Path,
					"name", options.Name)
			} else {
				uniquePaths[options.Path] = struct{}{}
			}
		}
	}
}

func NewServer(config *types.AppSettings) *Server {
	metadataMap, err := io.LoadAllMetadata(config.BackupDir)
	if err != nil {
		slog.Error("Error loading metadata", "error", err)
	}
	if metadataMap == nil {
		metadataMap = map[types.ConfigIdentifier]*types.ConfigMetadata{}
	}

	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		State: &State{
			CachedConfigMetadata: metadataMap,
			FileLookup:           make(map[string]*types.ConfigBackupOptions),
		},
		AppSettings: config,
		queue:       make(chan BackupJob),
		fileWatcher: fileWatcher,
	}
}

func (s *Server) Start() {
	s.startQueueProcessor()
	s.startFileWatcher()
	s.validateConfig()
	s.ProcessAllConfigOptions()
	_ = s.RestartCronJob()
}

type State struct {
	Mu                   sync.RWMutex
	CachedConfigMetadata map[types.ConfigIdentifier]*types.ConfigMetadata
	CronJob              *cron.Cron
	FileLookup           map[string]*types.ConfigBackupOptions
}

// Shutdown gracefully stops the server resources
func (s *Server) Shutdown() {
	slog.Info("Shutting down server...")
	if s.fileWatcher != nil {
		if err := s.fileWatcher.Close(); err != nil {
			slog.Error("Error closing file watcher", "error", err)
		}
	}
	if s.queue != nil {
		close(s.queue)
	}
	if s.State.CronJob != nil {
		s.State.CronJob.Stop()
	}
	slog.Info("Server shutdown complete")
}
