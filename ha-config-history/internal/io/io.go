package io

import (
	"encoding/json"
	"fmt"
	"ha-config-history/internal/types"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SanitizePath validates that a path doesn't contain directory traversal attempts
func SanitizePath(path string) error {
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("invalid path: contains directory traversal")
	}
	return nil
}

func ReadMultipleConfigsFromSingleFile(rootPath string, config *types.ConfigBackupOptions) ([]*types.ConfigBackup, error) {
	currentTime := time.Now().UTC()
	filePath := rootPath + "/" + config.Path

	configBackups := []*types.ConfigBackup{}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(data, &rootNode); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	if len(rootNode.Content) == 0 || rootNode.Content[0].Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected a YAML sequence at root")
	}

	contentNode := rootNode.Content[0]

	for _, yamlNode := range contentNode.Content {
		configBackup, err := types.NewYamlConfigBackup(config.Path, filePath, yamlNode, config, currentTime)
		if err != nil {
			return nil, fmt.Errorf("failed to create config backup for %s: %w", filePath, err)
		}
		configBackups = append(configBackups, configBackup)
	}

	return configBackups, nil
}

func ReadSingleConfigFromSingleFile(rootPath string, config *types.ConfigBackupOptions) (*types.ConfigBackup, error) {
	return ReadSingleConfigFromSingleFilename(rootPath, config.Path, config)
}

func ReadSingleConfigFromSingleFilename(rootPath, filename string, config *types.ConfigBackupOptions) (*types.ConfigBackup, error) {
	currentTime := time.Now().UTC()
	filePath := rootPath + "/" + filename

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	configBackup, err := types.NewBlobConfigBackup(filename, filePath, data, config, currentTime)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return configBackup, nil
}

func ReadMultipleConfigsFromDirectory(rootPath string, config *types.ConfigBackupOptions) ([]*types.ConfigBackup, error) {
	directoryPath := rootPath + "/" + config.Path

	configBackups := []*types.ConfigBackup{}
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", directoryPath, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		included, err := isFileIncluded(config, file)
		if err != nil {
			return nil, err
		}

		excluded, err := isFileExcluded(config, file)
		if err != nil {
			return nil, err
		}

		if !included || excluded {
			continue
		}

		configBackup, err := ReadSingleConfigFromSingleFilename(directoryPath, file.Name(), config)
		if err != nil {
			return nil, fmt.Errorf("failed to read config from file %s: %w", file.Name(), err)
		}
		configBackups = append(configBackups, configBackup)
	}

	return configBackups, nil
}

func isFileIncluded(config *types.ConfigBackupOptions, file os.DirEntry) (bool, error) {
	included := true
	if len(config.IncludeFilePatterns) > 0 {
		matched := false
		for _, pattern := range config.IncludeFilePatterns {
			match, err := filepath.Match(pattern, file.Name())
			if err != nil {
				return false, fmt.Errorf("invalid include pattern %s: %w", pattern, err)
			}
			if match {
				matched = true
				break
			}
		}

		if !matched {
			included = false
		}
	}
	return included, nil
}

func isFileExcluded(config *types.ConfigBackupOptions, file os.DirEntry) (bool, error) {
	excluded := false
	if len(config.ExcludeFilePatterns) > 0 {
		for _, pattern := range config.ExcludeFilePatterns {
			match, err := filepath.Match(pattern, file.Name())
			if err != nil {
				return false, fmt.Errorf("invalid exclude pattern %s: %w", pattern, err)
			}
			if match {
				excluded = true
				break
			}
		}
	}
	return excluded, nil
}

func LoadAllMetadata(backupFolder string) (map[types.ConfigIdentifier]*types.ConfigMetadata, error) {
	// BackupsFolder structure:
	//  - group1
	//    - config1
	//      - 20231010T120000.backup
	//      - 20231011T120000.yaml
	//      - metadata.json

	metadataMap := map[types.ConfigIdentifier]*types.ConfigMetadata{}

	groups, err := os.ReadDir(backupFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup folder %s: %w", backupFolder, err)
	}

	for _, group := range groups {
		if group.IsDir() {
			groupPath := filepath.Join(backupFolder, group.Name())
			configs, err := os.ReadDir(groupPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read group folder %s: %w", groupPath, err)
			}

			for _, config := range configs {
				if config.IsDir() {
					metadataPath := filepath.Join(groupPath, config.Name(), "metadata.json")
					metadataBlob, err := os.ReadFile(metadataPath)
					if err != nil {
						return nil, fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
					}

					var metadata types.ConfigMetadata
					if err := json.Unmarshal(metadataBlob, &metadata); err != nil {
						return nil, fmt.Errorf("failed to parse metadata JSON in %s: %w", metadataPath, err)
					}

					metadataMap[types.ConfigIdentifier{Group: group.Name(), ID: config.Name()}] = &metadata
				}
			}
		}
	}

	return metadataMap, nil
}

func GetBackupDirectory(backupFolder string, configBackup *types.ConfigBackup) (string, error) {
	configBackupFolder := filepath.Join(backupFolder, configBackup.Group, configBackup.ID)

	if _, err := os.Stat(configBackupFolder); os.IsNotExist(err) {
		err := os.MkdirAll(configBackupFolder, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create config backup directory %s: %w", configBackupFolder, err)
		}
	}

	return configBackupFolder, nil
}

func SaveConfigBackup(configBackup *types.ConfigBackup, backupDirectory string) error {
	backupPath := filepath.Join(backupDirectory, fmt.Sprintf("%s.backup", configBackup.ModifiedDate.Format("20060102T150405")))
	err := os.WriteFile(backupPath, configBackup.Blob, 0644)

	if err != nil {
		return fmt.Errorf("failed to save config backup: %w", err)
	}

	// Verify the backup was saved and can be read back
	if _, err := os.ReadFile(backupPath); err != nil {
		return fmt.Errorf("backup saved but cannot be read back from %s: %w", backupPath, err)
	}

	return nil
}

func CleanupAndUpdateMetadata(configBackup *types.ConfigBackup, backupOptions *types.ConfigBackupOptions, backupDirectory string, defaultMaxBackups *int, defaultMaxBackupAgeDays *int) (*types.ConfigMetadata, error) {
	backupsCount, backupsSize, err := dirMetrics(backupDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory metrics for %s: %w", backupDirectory, err)
	}

	// Determine effective MaxBackupAgeDays: prefer option value, fall back to default
	effectiveMaxBackupAgeDays := backupOptions.MaxBackupAgeDays
	if effectiveMaxBackupAgeDays == nil {
		effectiveMaxBackupAgeDays = defaultMaxBackupAgeDays
	}

	oldestBackupTimeAllowed := time.Unix(0, 0)
	if effectiveMaxBackupAgeDays != nil {
		oldestBackupTimeAllowed = time.Now().UTC().AddDate(0, 0, -*effectiveMaxBackupAgeDays)
	}

	// Determine effective MaxBackups: prefer option value, fall back to default
	effectiveMaxBackups := backupOptions.MaxBackups
	if effectiveMaxBackups == nil {
		effectiveMaxBackups = defaultMaxBackups
	}

	if effectiveMaxBackups != nil {
		entries, err := os.ReadDir(backupDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to read backup directory %s: %w", backupDirectory, err)
		}

		filenames := []string{}
		for _, entry := range entries {
			// TODO: V2 - remove .yaml
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".backup" || filepath.Ext(entry.Name()) == ".yaml") {
				filenames = append(filenames, entry.Name())
			}
		}

		sort.Sort(sort.Reverse(sort.StringSlice(filenames)))

		currentBackupCount := 0
		for _, filename := range filenames {
			currentBackupCount++
			if currentBackupCount > *effectiveMaxBackups {
				RemoveBackup(backupDirectory, filename, "exceeded max backups limit")
				continue
			}

			dateStr := filename[:len(filename)-5]
			backupDate, err := time.ParseInLocation("20060102T150405", dateStr, time.UTC)
			if err != nil {
				continue
			}
			if backupDate.Before(oldestBackupTimeAllowed) {
				RemoveBackup(backupDirectory, filename, "older than max backup age")
			}
		}
	}

	metadataPath := filepath.Join(backupDirectory, "metadata.json")
	metadata := types.NewConfigMetadata(configBackup, backupsCount, backupsSize, backupOptions.BackupType)
	metadataBlob, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return metadata, os.WriteFile(metadataPath, metadataBlob, 0644)
}

func RemoveBackup(backupDirectory, filename, reason string) {
	backupPath := filepath.Join(backupDirectory, filename)
	err := os.Remove(backupPath)
	if err != nil {
		slog.Error("Failed to remove old backup", "file", backupPath, "error", err, "reason", reason)
	} else {
		slog.Info("Removed old backup due to max backups limit", "file", backupPath, "reason", reason)
	}
}

func dirMetrics(path string) (int, int64, error) {
	var size int64
	var count int
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() != "metadata.json" {
			size += info.Size()
			count++
		}
		return err
	})
	return count, size, err
}

func GetConfigBackup(backupFolder, group, id, filename string) ([]byte, error) {
	// Validate path components for directory traversal
	if err := SanitizePath(group); err != nil {
		return nil, fmt.Errorf("invalid group parameter: %w", err)
	}
	if err := SanitizePath(id); err != nil {
		return nil, fmt.Errorf("invalid id parameter: %w", err)
	}
	if err := SanitizePath(filename); err != nil {
		return nil, fmt.Errorf("invalid filename parameter: %w", err)
	}

	backupPath := filepath.Join(backupFolder, group, id, filename)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file not found: %s", filename)
	}

	content, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	return content, nil
}

type BackupInfo struct {
	Filename string    `json:"filename"`
	Date     time.Time `json:"date"`
	Size     int64     `json:"size"`
}

func ListConfigBackups(backupFolder, group, configID string) ([]BackupInfo, error) {
	// Validate path components for directory traversal
	if err := SanitizePath(group); err != nil {
		return nil, fmt.Errorf("invalid group parameter: %w", err)
	}
	if err := SanitizePath(configID); err != nil {
		return nil, fmt.Errorf("invalid configID parameter: %w", err)
	}

	configFolder := filepath.Join(backupFolder, group, configID)

	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		return nil, fmt.Errorf("config not found: %s", configID)
	}

	entries, err := os.ReadDir(configFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read config folder %s: %w", configFolder, err)
	}

	backups := []BackupInfo{}
	for _, entry := range entries {
		// TODO: V2 - remove .yaml
		// .yaml kept for compatibility with previous versions
		if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".backup" || filepath.Ext(entry.Name()) == ".yaml") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			dateStr := entry.Name()[:len(entry.Name())-7] // Remove .backup extension
			if filepath.Ext(entry.Name()) == ".yaml" {
				dateStr = entry.Name()[:len(entry.Name())-5] // Remove .yaml extension
			}

			date, err := time.ParseInLocation("20060102T150405", dateStr, time.UTC)
			if err != nil {
				date = info.ModTime()
			}

			backups = append(backups, BackupInfo{
				Filename: entry.Name(),
				Date:     date,
				Size:     info.Size(),
			})
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Date.After(backups[j].Date)
	})

	return backups, nil
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func RestoreEntireFile(filepath string, blob []byte) error {
	err := os.WriteFile(filepath, blob, 0644)
	if err != nil {
		return fmt.Errorf("failed to restore config to %s: %w", filepath, err)
	}
	return nil
}

func RestorePartialFile(filepath string, blobToRestore []byte, options types.ConfigBackupOptions) error {
	var dataToRestore yaml.Node
	if err := yaml.Unmarshal(blobToRestore, &dataToRestore); err != nil {
		return fmt.Errorf("failed to parse backup YAML: %w", err)
	}
	nodeIdToRestore := types.GetYamlNodeValue(dataToRestore.Content[0], *options.IdNode)

	currentData, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read existing config file %s: %w", filepath, err)
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(currentData, &rootNode); err != nil {
		return fmt.Errorf("failed to parse existing YAML in %s: %w", filepath, err)
	}

	if len(rootNode.Content) == 0 || rootNode.Content[0].Kind != yaml.SequenceNode {
		return fmt.Errorf("expected a YAML sequence at root")
	}

	contentNode := rootNode.Content[0]
	for _, yamlNode := range contentNode.Content {
		existingNodeId := types.GetYamlNodeValue(yamlNode, *options.IdNode)
		if existingNodeId == nodeIdToRestore {
			*yamlNode = *dataToRestore.Content[0]
			break
		}
	}

	updatedBlob, err := yaml.Marshal(rootNode.Content[0])
	if err != nil {
		return fmt.Errorf("failed to serialize updated YAML: %w", err)
	}

	if err := os.WriteFile(filepath, updatedBlob, 0644); err != nil {
		return fmt.Errorf("failed to write updated config file %s: %w", filepath, err)
	}

	return nil
}

// DeleteBackup deletes a single backup file and returns an error if it fails
func DeleteBackup(backupFolder, group, id, filename string) error {
	// Validate path components for directory traversal
	if err := SanitizePath(group); err != nil {
		return fmt.Errorf("invalid group parameter: %w", err)
	}
	if err := SanitizePath(id); err != nil {
		return fmt.Errorf("invalid id parameter: %w", err)
	}
	if err := SanitizePath(filename); err != nil {
		return fmt.Errorf("invalid filename parameter: %w", err)
	}

	backupPath := filepath.Join(backupFolder, group, id, filename)

	// Check if file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", filename)
	}

	// Delete the file
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	slog.Info("Backup deleted", "file", backupPath)
	return nil
}

// UpdateMetadataAfterDeletion updates the metadata.json after a backup is deleted
// Returns nil metadata if no backups remain
func UpdateMetadataAfterDeletion(backupFolder, group, id string) (*types.ConfigMetadata, error) {
	backupDirectory := filepath.Join(backupFolder, group, id)

	// Get updated metrics
	backupsCount, backupsSize, err := dirMetrics(backupDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory metrics for %s: %w", backupDirectory, err)
	}

	// If no backups remain, delete metadata file and directory
	if backupsCount == 0 {
		metadataPath := filepath.Join(backupDirectory, "metadata.json")
		if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("Failed to remove metadata file", "path", metadataPath, "error", err)
		}

		// Try to remove the directory
		if err := os.Remove(backupDirectory); err != nil {
			slog.Warn("Failed to remove empty backup directory", "path", backupDirectory, "error", err)
		}

		return nil, nil
	}

	// Read existing metadata to preserve other fields
	metadataPath := filepath.Join(backupDirectory, "metadata.json")
	metadataBlob, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	var metadata types.ConfigMetadata
	if err := json.Unmarshal(metadataBlob, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON in %s: %w", metadataPath, err)
	}

	// Update counts
	metadata.BackupCount = backupsCount
	metadata.BackupsSize = backupsSize

	// Write updated metadata
	updatedMetadataBlob, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, updatedMetadataBlob, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return &metadata, nil
}

// DeleteAllBackups deletes all backups for a config (removes entire directory)
func DeleteAllBackups(backupFolder, group, id string) error {
	// Validate path components for directory traversal
	if err := SanitizePath(group); err != nil {
		return fmt.Errorf("invalid group parameter: %w", err)
	}
	if err := SanitizePath(id); err != nil {
		return fmt.Errorf("invalid id parameter: %w", err)
	}

	configFolder := filepath.Join(backupFolder, group, id)

	// Check if directory exists
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		return fmt.Errorf("config directory not found: %s", id)
	}

	// Delete the entire directory
	if err := os.RemoveAll(configFolder); err != nil {
		return fmt.Errorf("failed to delete config directory: %w", err)
	}

	slog.Info("All backups deleted", "group", group, "id", id)
	return nil
}
