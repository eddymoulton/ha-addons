package types

import (
	"encoding/json"
	"log/slog"
	"os"
)

type AppSettings struct {
	HomeAssistantConfigDir  string                 `json:"homeAssistantConfigDir"`
	BackupDir               string                 `json:"backupDir"`
	Port                    string                 `json:"port"`
	CronSchedule            *string                `json:"cronSchedule,omitempty"`
	DefaultMaxBackups       *int                   `json:"defaultMaxBackups,omitempty"`
	DefaultMaxBackupAgeDays *int                   `json:"defaultMaxBackupAgeDays,omitempty"`
	Configs                 []*ConfigBackupOptions `json:"configs"`
}

type ConfigBackupOptions struct {
	Name                string   `json:"name"`
	Path                string   `json:"path"`
	BackupType          string   `json:"backupType"` // "multiple", "single", "directory"
	MaxBackups          *int     `json:"maxBackups,omitempty"`
	MaxBackupAgeDays    *int     `json:"maxBackupAgeDays,omitempty"`
	IdNode              *string  `json:"idNode,omitempty"`
	FriendlyNameNode    *string  `json:"friendlyNameNode,omitempty"`
	IncludeFilePatterns []string `json:"includeFilePatterns,omitempty"`
	ExcludeFilePatterns []string `json:"excludeFilePatterns,omitempty"`
}

func NewSingleConfigBackupOptions(name string, path string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Name:       name,
		Path:       path,
		BackupType: "single",
	}
}

func NewMultipleConfigBackupOptions(name string, path string, idNodeName string, friendNameNodeName string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Name:             name,
		Path:             path,
		BackupType:       "multiple",
		IdNode:           &idNodeName,
		FriendlyNameNode: &friendNameNodeName,
	}
}

func NewDirectoryConfigBackupOptions(name string, path string, includeFilePatterns, excludeFilePatterns []string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Name:                name,
		Path:                path,
		BackupType:          "directory",
		IncludeFilePatterns: includeFilePatterns,
		ExcludeFilePatterns: excludeFilePatterns,
	}
}

func LoadConfig(configPath string) *AppSettings {
	appSettings := &AppSettings{
		HomeAssistantConfigDir:  "/homeassistant",
		BackupDir:               "/data/ha-config-history/backups",
		Port:                    ":40613",
		CronSchedule:            new(string),
		DefaultMaxBackups:       nil,
		DefaultMaxBackupAgeDays: nil,
		Configs: []*ConfigBackupOptions{
			NewSingleConfigBackupOptions("Configuration", "configuration.yaml"),
			NewMultipleConfigBackupOptions("Automations", "automations.yaml", "id", "alias"),
			NewMultipleConfigBackupOptions("Scenes", "scenes.yaml", "id", "name"),
			NewDirectoryConfigBackupOptions("ESP Home", "esphome", []string{"*.yaml"}, []string{}),
			NewDirectoryConfigBackupOptions("Storage", ".storage", []string{
				"lovelace.*",
				"core.*",
				"counter.*",
				"input_boolean.*",
				"input_number.*",
				"input_select.*",
				"input_text.*",
				"input_*",
				"person",
				"energy",
				"schedule",
				"timer",
			}, []string{
				"core.analytics",
				"core.config_entries",
				"core.restore_state",
				"core.device_registry",
				"core.entity_registry",
				"core.uuid",
			}),
		},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Warn("Config file not found or unreadable, using defaults", "error", err)
		return appSettings
	}

	if err := json.Unmarshal(data, appSettings); err != nil {
		slog.Warn("Failed to parse config file, using defaults", "error", err)
		return appSettings
	}

	slog.Info("Loaded configuration",
		"homeassistantconfigdir", appSettings.HomeAssistantConfigDir,
		"backupDir", appSettings.BackupDir,
		"port", appSettings.Port,
	)

	return appSettings
}

type BackupType int

const (
	BackupTypeMultiple BackupType = iota
	BackupTypeSingle
	BackupTypeDirectory
)

// Backup type string constants
const (
	BackupTypeMultipleName  = "multiple"
	BackupTypeSingleName    = "single"
	BackupTypeDirectoryName = "directory"
)

var stateName = map[BackupType]string{
	BackupTypeMultiple:  BackupTypeMultipleName,
	BackupTypeSingle:    BackupTypeSingleName,
	BackupTypeDirectory: BackupTypeDirectoryName,
}
