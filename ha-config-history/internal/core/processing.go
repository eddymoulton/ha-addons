package core

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"time"
)

func (s *Server) ProcessAllConfigOptions() {
	// Process configs from new grouped structure
	for _, group := range s.AppSettings.ConfigGroups {
		for _, options := range group.Configs {
			s.processConfigOptions(options)
		}
	}

}

func (s *Server) processConfigOptions(options *types.ConfigBackupOptions) {
	if options.BackupType == "multiple" {
		current, err := io.ReadMultipleConfigsFromSingleFile(s.AppSettings.HomeAssistantConfigDir, options)
		if err != nil {
			slog.Error("Error reading single file for multiple configs", "error", err)
			return
		}

		slog.Info("Processing backups for multiple configs",
			"found_active_configs", len(current),
			"known_backups", len(s.State.CachedConfigMetadata),
		)

		for _, configBackup := range current {
			s.queue <- BackupJob{
				Options: options,
				Backup:  configBackup,
			}
		}

		for _, configBackup := range current {
			err = s.watchDirectoryForFile(configBackup.FilePath, options)
			if err != nil {
				slog.Error("Error watching file for changes", "error", err)
			}
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

		s.queue <- BackupJob{
			Options: options,
			Backup:  configBackup,
		}

		err = s.watchDirectoryForFile(configBackup.FilePath, options)
		if err != nil {
			slog.Error("Error watching file for changes", "error", err)
		}
	}

	if options.BackupType == "directory" {
		current, err := io.ReadMultipleConfigsFromDirectory(s.AppSettings.HomeAssistantConfigDir, options)
		if err != nil {
			slog.Error("Error reading configs from directory", "error", err)
			return
		}

		slog.Info("Processing backups for directory configs",
			"found_active_configs", len(current),
			"known_backups", len(s.State.CachedConfigMetadata),
		)

		for _, configBackup := range current {
			s.queue <- BackupJob{
				Options: options,
				Backup:  configBackup,
			}
		}

		for _, configBackup := range current {
			err = s.watchDirectoryForFile(configBackup.FilePath, options)
			if err != nil {
				slog.Error("Error watching file for changes", "error", err)
			}
		}
	}
}

func (s *Server) startQueueProcessor() {
	go func() {
		for job := range s.queue {
			start := time.Now()
			s.handleUpdateToFile(job.Options, job.Backup)
			slog.Debug("Processed backup job",
				"id", job.Backup.ID,
				"path", job.Backup.Path,
				"duration", time.Since(start),
			)
		}
		slog.Info("Queue processor stopped")
	}()
}

func (s *Server) handleUpdateToFile(backupOptions *types.ConfigBackupOptions, activeConfigBackup *types.ConfigBackup) {
	s.State.Mu.RLock()
	for _, metadata := range s.State.CachedConfigMetadata {
		slog.Debug("Cached config metadata",
			"friendlyName", metadata.FriendlyName,
			"id", metadata.ID,
			"backups", metadata.BackupCount,
			"size", metadata.BackupsSize,
		)
	}

	metadata, exists := s.State.CachedConfigMetadata[activeConfigBackup.ConfigIdentifier]
	needsBackup := !exists || activeConfigBackup.Hash != metadata.LastHash
	s.State.Mu.RUnlock()

	if needsBackup {
		slog.Info("Config changed, saving backup",
			"friendlyName", activeConfigBackup.FriendlyName,
			"id", activeConfigBackup.ID,
		)

		backupDir, err := io.GetBackupDirectory(s.AppSettings.BackupDir, activeConfigBackup)
		if err != nil {
			slog.Error("Error getting config backup directory",
				"id", activeConfigBackup.ID,
				"error", err,
			)
		}

		err = io.SaveConfigBackup(activeConfigBackup, backupDir)
		if err != nil {
			slog.Error("Error saving config backup",
				"id", activeConfigBackup.ID,
				"error", err,
			)
		}

		updatedMetadata, err := io.CleanupAndUpdateMetadata(activeConfigBackup, backupOptions, backupDir, s.AppSettings.DefaultMaxBackups, s.AppSettings.DefaultMaxBackupAgeDays)
		if err != nil {
			slog.Error("Error updating config metadata",
				"id", activeConfigBackup.ID,
				"error", err,
			)
		}

		if updatedMetadata != nil {
			s.State.Mu.Lock()
			s.State.CachedConfigMetadata[activeConfigBackup.ConfigIdentifier] = updatedMetadata
			s.State.Mu.Unlock()
		}
	}
}
