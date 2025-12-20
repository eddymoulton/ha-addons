package core

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"time"
)

func (s *Server) ProcessAllConfigOptions() {
	for _, group := range s.AppSettings.ConfigGroups {
		for _, options := range group.Configs {
			s.processConfigOptions(group.Slug, options)
		}
	}

}

func (s *Server) processConfigOptions(groupSlug types.GroupSlug, options *types.ConfigBackupOptions) {
	if options.BackupType == "multiple" {
		current, err := io.ReadMultipleConfigsFromSingleFile(s.AppSettings.HomeAssistantConfigDir, options)
		if err != nil {
			slog.Error("Error reading single file for multiple configs", "error", err)
			return
		}

		slog.Info("Processing backups for multiple configs",
			"found_active_configs", len(current),
			"known_backups", len(s.State.CachedBackupSummaries),
		)

		for _, configBackup := range current {
			s.queueAndWatch(groupSlug, options, configBackup)
		}
	}

	if options.BackupType == "single" {
		configBackup, err := io.ReadSingleConfigFromSingleFile(s.AppSettings.HomeAssistantConfigDir, options)
		if err != nil {
			slog.Error("Error reading single config file", "error", err)
			return
		}

		slog.Info("Processing backup for single config",
			"id", configBackup.ID,
			"friendlyName", configBackup.FriendlyName,
		)

		s.queueAndWatch(groupSlug, options, configBackup)
	}

	if options.BackupType == "directory" {
		current, err := io.ReadMultipleConfigsFromDirectory(s.AppSettings.HomeAssistantConfigDir, options)
		if err != nil {
			slog.Error("Error reading configs from directory", "error", err)
			return
		}

		slog.Info("Processing backups for directory configs",
			"found_active_configs", len(current),
			"known_backups", len(s.State.CachedBackupSummaries),
		)

		for _, configBackup := range current {
			s.queueAndWatch(groupSlug, options, configBackup)
		}
	}
}

func (s *Server) queueAndWatch(groupSlug types.GroupSlug, options *types.ConfigBackupOptions, configBackup *types.ConfigBackup) {
	s.queue <- NewBackupJob(groupSlug, options, configBackup)
	err := s.watchDirectoryForFile(groupSlug, configBackup.FilePath, options)
	if err != nil {
		slog.Error("Error watching file for changes", "error", err)
	}
}

func (s *Server) startQueueProcessor() {
	go func() {
		for job := range s.queue {
			start := time.Now()
			s.processingFile = true
			s.handleUpdateToFile(job.GroupSlug, job.Options, job.Backup)
			s.processingFile = false
			slog.Debug("Processed backup job",
				"id", job.Backup.ID,
				"path", job.Backup.Path,
				"group", job.GroupSlug,
				"duration", time.Since(start),
			)
		}
		slog.Info("Queue processor stopped")
	}()
}

func (s *Server) handleUpdateToFile(
	groupSlug types.GroupSlug,
	backupOptions *types.ConfigBackupOptions,
	activeConfigBackup *types.ConfigBackup) {

	if s.needsUpdate(groupSlug, activeConfigBackup) {
		slog.Info("Config changed, saving backup",
			"friendlyName", activeConfigBackup.FriendlyName,
			"id", activeConfigBackup.ID,
		)

		err := io.SaveConfigBackup(s.AppSettings.BackupDir, groupSlug, activeConfigBackup)
		if err != nil {
			slog.Error("Error saving config backup",
				"id", activeConfigBackup.ID,
				"error", err,
			)
		}

		updatedMetadata, err := io.CleanupAndUpdateMetadata(
			groupSlug,
			activeConfigBackup,
			backupOptions,
			s.AppSettings.BackupDir,
			s.AppSettings.DefaultMaxBackups,
			s.AppSettings.DefaultMaxBackupAgeDays)

		if err != nil {
			slog.Error("Error updating config metadata",
				"id", activeConfigBackup.ID,
				"error", err,
			)
		}

		if updatedMetadata != nil {
			s.updateCachedMetadata(groupSlug, updatedMetadata)
		}
	}
}

func (s *Server) needsUpdate(groupSlug types.GroupSlug, activeConfigBackup *types.ConfigBackup) bool {
	s.State.Mu.RLock()
	defer s.State.Mu.RUnlock()

	groupMetadata, exists := s.State.CachedBackupSummaries[groupSlug]
	if !exists {
		return true
	}

	metadata, exists := groupMetadata[activeConfigBackup.ConfigBackupIdentifier]
	return !exists || activeConfigBackup.Hash != metadata.LastHash
}

func (s *Server) updateCachedMetadata(groupSlug types.GroupSlug, metadata *types.BackupConfigSummary) {
	s.State.Mu.Lock()
	defer s.State.Mu.Unlock()

	groupMetadata, exists := s.State.CachedBackupSummaries[groupSlug]
	if !exists {
		groupMetadata = types.BackupConfigSummaryMap{}
		s.State.CachedBackupSummaries[groupSlug] = groupMetadata
	}

	groupMetadata[metadata.ConfigBackupIdentifier] = metadata
}
