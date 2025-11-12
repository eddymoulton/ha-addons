package core

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"log/slog"
	"path/filepath"
)

func (s *Server) startFileWatcher() {
	go func() {
		for {
			select {
			case event, ok := <-s.fileWatcher.Events:
				if !ok {
					return
				}
				slog.Debug("File watcher event", "file", event.Name, "event", event.Op)

				s.State.Mu.RLock()
				options, exists := s.State.FileLookup[event.Name]
				s.State.Mu.RUnlock()

				if !exists {
					slog.Debug("No backup options found for changed file", "file", event.Name)
					continue
				}

				if options.BackupType == "single" {
					backup, err := io.ReadSingleConfigFromSingleFilename(s.AppSettings.HomeAssistantConfigDir, options.Path, options)
					if err != nil {
						slog.Error("Error reading updated config from file", "file", event.Name, "error", err)
						continue
					}

					s.queue <- BackupJob{
						Options: options,
						Backup:  backup,
					}
				}

				if options.BackupType == "directory" {
					filename := filepath.Base(event.Name)
					fullDirectory := filepath.Dir(event.Name)
					backup, err := io.ReadSingleConfigFromSingleFilename(fullDirectory, filename, options)
					if err != nil {
						slog.Error("Error reading updated config from file", "file", event.Name, "error", err)
						continue
					}

					s.queue <- BackupJob{
						Options: options,
						Backup:  backup,
					}
				}

				if options.BackupType == "multiple" {
					current, err := io.ReadMultipleConfigsFromSingleFile(s.AppSettings.HomeAssistantConfigDir, options)
					if err != nil {
						slog.Error("Error reading updated multiple configs from file", "file", event.Name, "error", err)
						continue
					}

					for _, configBackup := range current {
						s.queue <- BackupJob{
							Options: options,
							Backup:  configBackup,
						}
					}
				}

			case err, ok := <-s.fileWatcher.Errors:
				if !ok {
					return
				}
				slog.Error("File watcher error", "error", err)
			}
		}
	}()
}

func (s *Server) watchDirectoryForFile(path string, options *types.ConfigBackupOptions) error {
	directory := filepath.Dir(path)
	slog.Info("Adding directory to watcher for file", "directory", directory, "file", options.Path)

	for _, existing := range s.fileWatcher.WatchList() {
		if existing == directory {
			slog.Info("Directory already being watched", "directory", directory)
			s.State.Mu.Lock()
			s.State.FileLookup[path] = options
			s.State.Mu.Unlock()
			return nil
		}
	}

	err := s.fileWatcher.Add(directory)
	if err != nil {
		slog.Error("Error adding directory watcher", "error", err)
	}

	s.State.Mu.Lock()
	s.State.FileLookup[path] = options
	s.State.Mu.Unlock()
	return err
}
