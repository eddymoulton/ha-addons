package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigBackupIdentifier uniquely identifies a single configuration backup
// ie. a single file that is backed up multiple times
type ConfigBackupIdentifier struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type BackupConfigSummaryMap map[ConfigBackupIdentifier]*BackupConfigSummary

type BackupConfigSummary struct {
	ConfigBackupIdentifier
	FriendlyName string `json:"friendlyName"`
	LastHash     string `json:"lastHash"`
	BackupCount  int    `json:"backupCount"`
	BackupsSize  int64  `json:"backupsSize"`
	BackupType   string `json:"backupType"`

	// TODO: V2 Remove
	Group string `json:"group,omitempty"` // For backward compatibility
}

func NewConfigBackupSummary(configBackup *ConfigBackup, backupCount int, backupsSize int64, backupType string) *BackupConfigSummary {
	return &BackupConfigSummary{
		ConfigBackupIdentifier: ConfigBackupIdentifier{
			ID:   configBackup.ID,
			Path: configBackup.Path,
		},
		FriendlyName: configBackup.FriendlyName,
		LastHash:     configBackup.Hash,
		BackupCount:  backupCount,
		BackupsSize:  backupsSize,
		BackupType:   backupType,
	}
}

type ConfigBackup struct {
	ConfigBackupIdentifier
	FriendlyName string `json:"friendlyName,omitempty"`
	Hash         string `json:"hash,omitempty"`
	ModifiedDate time.Time
	BackupType   string `json:"backupType"` // "multiple", "single", "directory"
	FilePath     string `json:"-"`
	Blob         []byte `json:"-"`
}

func NewBlobConfigBackup(filename, filepath string, blob []byte, config *ConfigBackupOptions, modifiedDate time.Time) (*ConfigBackup, error) {
	if config.BackupType == stateName[BackupTypeSingle] {
		return &ConfigBackup{
			ConfigBackupIdentifier: ConfigBackupIdentifier{
				ID:   config.Path,
				Path: config.Path,
			},
			FriendlyName: config.Path,
			Hash:         hashByteSlice(blob),
			ModifiedDate: modifiedDate,
			BackupType:   config.BackupType,
			FilePath:     filepath,
			Blob:         blob,
		}, nil
	}

	if config.BackupType == stateName[BackupTypeMultiple] {
		return nil, fmt.Errorf("blob backups do not support multiple backup type")
	}

	if config.BackupType == stateName[BackupTypeDirectory] {
		return &ConfigBackup{
			ConfigBackupIdentifier: ConfigBackupIdentifier{
				ID:   filename,
				Path: config.Path,
			},
			FriendlyName: filename,
			Hash:         hashByteSlice(blob),
			BackupType:   config.BackupType,
			ModifiedDate: modifiedDate,
			FilePath:     filepath,
			Blob:         blob,
		}, nil
	}

	return nil, fmt.Errorf("unknown backup type: %s", config.BackupType)
}

func NewYamlConfigBackup(filename, filepath string, yamlNode *yaml.Node, config *ConfigBackupOptions, modifiedDate time.Time) (*ConfigBackup, error) {
	blob, _ := yaml.Marshal(yamlNode)

	if config.BackupType == stateName[BackupTypeSingle] {
		return nil, fmt.Errorf("yaml backups do not support single backup type")
	}

	if config.BackupType == stateName[BackupTypeMultiple] {
		return &ConfigBackup{
			ConfigBackupIdentifier: ConfigBackupIdentifier{
				ID:   GetYamlNodeValue(yamlNode, *config.IdNode),
				Path: config.Path,
			},
			FriendlyName: GetYamlNodeValue(yamlNode, *config.FriendlyNameNode),
			Hash:         hashByteSlice(blob),
			BackupType:   config.BackupType,
			ModifiedDate: modifiedDate,
			FilePath:     filepath,
			Blob:         blob,
		}, nil
	}

	if config.BackupType == stateName[BackupTypeDirectory] {
		return nil, fmt.Errorf("yaml backups do not support directory backup type")
	}

	return nil, fmt.Errorf("unknown backup type: %s", config.BackupType)
}
