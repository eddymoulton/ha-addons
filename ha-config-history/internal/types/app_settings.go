package types

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
)

type GroupSlug string

type ConfigBackupOptionGroup struct {
	Name    string                 `json:"groupName"`
	Slug    GroupSlug              `json:"slug"`
	Configs []*ConfigBackupOptions `json:"configs"`
}

func NewConfigBackupOptionGroup(name string, configs []*ConfigBackupOptions) *ConfigBackupOptionGroup {
	slug := GroupSlug(strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	return &ConfigBackupOptionGroup{
		Name:    name,
		Slug:    slug,
		Configs: configs,
	}
}

type AppSettings struct {
	HomeAssistantConfigDir  string                     `json:"homeAssistantConfigDir"`
	BackupDir               string                     `json:"backupDir"`
	Port                    string                     `json:"port"`
	CronSchedule            *string                    `json:"cronSchedule,omitempty"`
	DefaultMaxBackups       *int                       `json:"defaultMaxBackups,omitempty"`
	DefaultMaxBackupAgeDays *int                       `json:"defaultMaxBackupAgeDays,omitempty"`
	ConfigGroups            []*ConfigBackupOptionGroup `json:"configGroups,omitempty"`
	Configs                 []*ConfigBackupOptions     `json:"configs,omitempty"` // Deprecated: kept for migration
}

type ConfigBackupOptions struct {
	Path                string   `json:"path"`
	BackupType          string   `json:"backupType"` // "multiple", "single", "directory"
	MaxBackups          *int     `json:"maxBackups,omitempty"`
	MaxBackupAgeDays    *int     `json:"maxBackupAgeDays,omitempty"`
	IdNode              *string  `json:"idNode,omitempty"`
	FriendlyNameNode    *string  `json:"friendlyNameNode,omitempty"`
	IncludeFilePatterns []string `json:"includeFilePatterns,omitempty"`
	ExcludeFilePatterns []string `json:"excludeFilePatterns,omitempty"`
}

func NewSingleConfigBackupOptions(path string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Path:       path,
		BackupType: "single",
	}
}

func NewMultipleConfigBackupOptions(path string, idNodeName string, friendNameNodeName string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Path:             path,
		BackupType:       "multiple",
		IdNode:           &idNodeName,
		FriendlyNameNode: &friendNameNodeName,
	}
}

func NewDirectoryConfigBackupOptions(path string, includeFilePatterns, excludeFilePatterns []string) *ConfigBackupOptions {
	return &ConfigBackupOptions{
		Path:                path,
		BackupType:          "directory",
		IncludeFilePatterns: includeFilePatterns,
		ExcludeFilePatterns: excludeFilePatterns,
	}
}

// saveAppSettings writes the AppSettings to the config file
func saveAppSettings(configPath string, appSettings *AppSettings) error {
	data, err := json.MarshalIndent(appSettings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func LoadAppSettings(appSettingsPath string) *AppSettings {
	defaultConfigGroups := []*ConfigBackupOptionGroup{
		NewConfigBackupOptionGroup("Core Home Assistant", []*ConfigBackupOptions{
			NewSingleConfigBackupOptions("configuration.yaml"),
			NewDirectoryConfigBackupOptions(".storage",
				[]string{"core.*", "frontend.*", "person"},
				[]string{
					"core.analytics",
					"core.config_entries",
					"core.restore_state",
					"core.device_registry",
					"core.entity_registry",
					"core.uuid",
				}),
		}),
		NewConfigBackupOptionGroup("Automations", []*ConfigBackupOptions{
			NewMultipleConfigBackupOptions("automations.yaml", "id", "alias"),
		}),
		NewConfigBackupOptionGroup("Scenes", []*ConfigBackupOptions{
			NewMultipleConfigBackupOptions("scenes.yaml", "id", "name"),
		}),
		NewConfigBackupOptionGroup("ESP Home", []*ConfigBackupOptions{
			NewDirectoryConfigBackupOptions("esphome", []string{"*.yaml"}, []string{"secrets.yaml"}),
		}),
		NewConfigBackupOptionGroup("Dashboards", []*ConfigBackupOptions{
			NewDirectoryConfigBackupOptions(
				".storage",
				[]string{
					"lovelace.*",
					"lovelace_dashboards",
					"energy",
				},
				[]string{}),
		}),
		NewConfigBackupOptionGroup("Helpers", []*ConfigBackupOptions{
			NewDirectoryConfigBackupOptions(
				".storage",
				[]string{
					"counter.*",
					"frontend.*",
					"input_boolean.*",
					"input_number.*",
					"input_select.*",
					"input_text.*",
					"input_*",
					"schedule",
					"timer",
				},
				[]string{}),
		}),
	}

	appSettings := &AppSettings{
		HomeAssistantConfigDir:  "/homeassistant",
		BackupDir:               "/data/backups",
		Port:                    ":40613",
		CronSchedule:            new(string),
		DefaultMaxBackups:       nil,
		DefaultMaxBackupAgeDays: nil,
		ConfigGroups:            defaultConfigGroups,
	}

	data, err := os.ReadFile(appSettingsPath)
	if err != nil {
		slog.Warn("Config file not found or unreadable, using defaults", "error", err)
		return appSettings
	}

	if err := json.Unmarshal(data, appSettings); err != nil {
		slog.Warn("Failed to parse config file, using defaults", "error", err)
		return appSettings
	}

	if len(appSettings.Configs) > 0 && len(appSettings.ConfigGroups) > 0 {
		slog.Warn("Both old configs and new config groups exist. Using config groups and clearing old configs.")
		appSettings.Configs = nil
		return appSettings
	}

	slog.Info("Loaded configuration",
		"homeassistantconfigdir", appSettings.HomeAssistantConfigDir,
		"backupDir", appSettings.BackupDir,
		"port", appSettings.Port,
		"configGroups", len(appSettings.ConfigGroups),
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
