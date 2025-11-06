package types

import (
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigIdentifier struct {
	ID    string `json:"id,omitempty"`
	Group string `json:"group,omitempty"`
}

type ConfigMetadata struct {
	ConfigIdentifier
	FriendlyName string `json:"friendlyName"`
	LastHash     string `json:"lastHash,omitempty"`
	BackupCount  int    `json:"backupCount"`
	BackupsSize  int64  `json:"backupsSize"`
	BackupType   string `json:"backupType"`
}

func NewConfigMetadata(configBackup *ConfigBackup, backupCount int, backupsSize int64, backupType string) *ConfigMetadata {
	return &ConfigMetadata{
		ConfigIdentifier: ConfigIdentifier{
			ID:    configBackup.ID,
			Group: configBackup.Group,
		},
		FriendlyName: configBackup.FriendlyName,
		LastHash:     configBackup.Hash,
		BackupCount:  backupCount,
		BackupsSize:  backupsSize,
		BackupType:   backupType,
	}
}

type ConfigBackup struct {
	ConfigIdentifier
	FriendlyName string `json:"friendly_name,omitempty"`
	Hash         string `json:"hash,omitempty"`
	ModifiedDate time.Time
	BackupType   string `json:"backupType"` // "multiple", "single", "directory"
	FilePath     string `json:"-"`
	Blob         []byte `json:"-"`
}

func NewConfigBackup(filename, filepath string, yamlNode *yaml.Node, config *ConfigBackupOptions, modifiedDate time.Time) *ConfigBackup {
	blob, _ := yaml.Marshal(yamlNode)

	if config.BackupType == stateName[BackupTypeSingle] {
		return &ConfigBackup{
			ConfigIdentifier: ConfigIdentifier{
				ID:    config.Path,
				Group: config.Path,
			},
			FriendlyName: config.Path,
			Hash:         hashByteSlice(blob),
			ModifiedDate: modifiedDate,
			BackupType:   config.BackupType,
			FilePath:     filepath,
			Blob:         blob,
		}
	}

	if config.BackupType == stateName[BackupTypeMultiple] {
		return &ConfigBackup{
			ConfigIdentifier: ConfigIdentifier{
				ID:    GetYamlNodeValue(yamlNode, *config.IdNode),
				Group: config.Path,
			},
			FriendlyName: GetYamlNodeValue(yamlNode, *config.FriendlyNameNode),
			Hash:         hashByteSlice(blob),
			BackupType:   config.BackupType,
			ModifiedDate: modifiedDate,
			FilePath:     filepath,
			Blob:         blob,
		}
	}

	if config.BackupType == stateName[BackupTypeDirectory] {
		return &ConfigBackup{
			ConfigIdentifier: ConfigIdentifier{
				ID:    filename,
				Group: config.Path,
			},
			FriendlyName: filename,
			Hash:         hashByteSlice(blob),
			BackupType:   config.BackupType,
			ModifiedDate: modifiedDate,
			FilePath:     filepath,
			Blob:         blob,
		}
	}

	return &ConfigBackup{
		ConfigIdentifier: ConfigIdentifier{
			ID:    "unknown",
			Group: "unknown",
		},
		FriendlyName: "unknown",
		Hash:         hashByteSlice(blob),
		BackupType:   config.BackupType,
		ModifiedDate: modifiedDate,
		FilePath:     filepath,
		Blob:         blob,
	}
}
