package core

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

// backupJob represents a backup job to be processed in the queue
type backupJob struct {
	GroupSlug types.GroupSlug
	Options   *types.ConfigBackupOptions
	Backup    *types.ConfigBackup
}

func NewBackupJob(
	groupSlug types.GroupSlug,
	options *types.ConfigBackupOptions,
	backup *types.ConfigBackup) backupJob {
	return backupJob{
		GroupSlug: groupSlug,
		Options:   options,
		Backup:    backup,
	}
}

type Server struct {
	State          *State
	AppSettings    *types.AppSettings
	ConfigPath     string
	queue          chan backupJob
	processingFile bool
	fileWatcher    *fsnotify.Watcher
}

func (s *Server) validateConfig() {
	if !io.DirectoryExists(s.AppSettings.HomeAssistantConfigDir) {
		slog.Error("Home Assistant configuration directory does not exist",
			"dir", s.AppSettings.HomeAssistantConfigDir)
	}

	// Check configs from new grouped structure
	for _, group := range s.AppSettings.ConfigGroups {
		uniquePaths := make(map[string]struct{})
		for _, options := range group.Configs {
			if _, exists := uniquePaths[options.Path]; exists {
				slog.Warn("Duplicate config path found in group",
					"group", group.Name,
					"path", options.Path)
			} else {
				uniquePaths[options.Path] = struct{}{}
			}
		}
	}
}

func NewServer(config *types.AppSettings, configPath string) *Server {
	summaries, err := io.LoadAllBackupConfigSummaries(config.BackupDir)
	if err != nil {
		slog.Error("Error loading metadata", "error", err)
	}
	if summaries == nil {
		summaries = map[types.GroupSlug]types.BackupConfigSummaryMap{}
	}

	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		State: &State{
			CachedBackupSummaries: summaries,
			FileLookup:            make(WatchedFileLookup),
		},
		AppSettings: config,
		ConfigPath:  configPath,
		queue:       make(chan backupJob),
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

func (s *Server) WaitForInactive() {
	time.Sleep(1 * time.Second)

	for len(s.queue) != 0 || s.processingFile {
		time.Sleep(100 * time.Millisecond)
	}
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

type State struct {
	Mu                    sync.RWMutex
	CachedBackupSummaries map[types.GroupSlug]types.BackupConfigSummaryMap
	CronJob               *cron.Cron
	FileLookup            WatchedFileLookup
}

type GroupedConfigBackupOptions struct {
	GroupSlug types.GroupSlug
	Options   *types.ConfigBackupOptions
}

type WatchedFileLookup map[string][]GroupedConfigBackupOptions

func (w WatchedFileLookup) AddOrUpdate(filePath string, groupSlug types.GroupSlug, options *types.ConfigBackupOptions) {
	entries, exists := w[filePath]
	if !exists {
		w[filePath] = []GroupedConfigBackupOptions{
			{
				GroupSlug: groupSlug,
				Options:   options,
			},
		}
		return
	}

	for i, entry := range entries {
		if entry.GroupSlug == groupSlug {
			entries[i].Options = options
			w[filePath] = entries
			return
		}
	}

	w[filePath] = append(entries, GroupedConfigBackupOptions{
		GroupSlug: groupSlug,
		Options:   options,
	})
}
