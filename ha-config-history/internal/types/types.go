package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigIdentifier struct {
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
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
	ConfigIdentifier
	FriendlyName string `json:"friendly_name,omitempty"`
	Hash         string `json:"hash,omitempty"`
	ModifiedDate time.Time
	BackupType   string `json:"backupType"` // "multiple", "single", "directory"
	FilePath     string `json:"-"`
	Blob         []byte `json:"-"`
}

func NewBlobConfigBackup(filename, filepath string, blob []byte, config *ConfigBackupOptions, modifiedDate time.Time) (*ConfigBackup, error) {
	if config.BackupType == stateName[BackupTypeSingle] {
		return &ConfigBackup{
			ConfigIdentifier: ConfigIdentifier{
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
			ConfigIdentifier: ConfigIdentifier{
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
			ConfigIdentifier: ConfigIdentifier{
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
