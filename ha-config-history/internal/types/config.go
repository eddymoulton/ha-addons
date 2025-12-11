package types

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
)

type ConfigBackupOptionGroup struct {
	GroupName string                 `json:"groupName"`
	Configs   []*ConfigBackupOptions `json:"configs"`
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

// TODO: V2 - Remove this migration function
// migrateToGroups converts the old configs array to the new grouped structure
func migrateToGroups(configs []*ConfigBackupOptions) []*ConfigBackupOptionGroup {
	if configs == nil {
		return []*ConfigBackupOptionGroup{}
	}

	groups := make(map[string]*ConfigBackupOptionGroup)

	for _, config := range configs {
		if config == nil {
			slog.Warn("Skipping nil config during migration")
			continue
		}

		var groupName string

		switch {
		case config.Name == "Configuration" || strings.Contains(strings.ToLower(config.Path), "configuration"):
			groupName = "Core Home Assistant"
		case config.Name == "Automations" || strings.Contains(strings.ToLower(config.Path), "automation"):
			groupName = "Automations"
		case config.Name == "Scenes" || strings.Contains(strings.ToLower(config.Path), "scene"):
			groupName = "Scenes"
		case config.Name == "ESP Home" || strings.Contains(strings.ToLower(config.Path), "esphome"):
			groupName = "ESP Home"
		case config.Name == "Storage" || strings.Contains(strings.ToLower(config.Path), ".storage"):
			groupName = "Storage & Settings"
		default:
			groupName = config.Name
		}

		if group, exists := groups[groupName]; exists {
			group.Configs = append(group.Configs, config)
		} else {
			groups[groupName] = &ConfigBackupOptionGroup{
				GroupName: groupName,
				Configs:   []*ConfigBackupOptions{config},
			}
		}
	}

	result := make([]*ConfigBackupOptionGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// saveConfig writes the AppSettings to the config file
func saveConfig(configPath string, appSettings *AppSettings) error {
	data, err := json.MarshalIndent(appSettings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func LoadConfig(configPath string) *AppSettings {
	defaultConfigGroups := []*ConfigBackupOptionGroup{
		{
			GroupName: "Core Home Assistant",
			Configs: []*ConfigBackupOptions{
				NewSingleConfigBackupOptions("Configuration", "configuration.yaml"),
				NewDirectoryConfigBackupOptions("Storage", ".storage",
					[]string{"core.*", "frontend.*", "person"},
					[]string{
						"core.analytics",
						"core.config_entries",
						"core.restore_state",
						"core.device_registry",
						"core.entity_registry",
						"core.uuid",
					})},
		},
		{
			GroupName: "Automations",
			Configs:   []*ConfigBackupOptions{NewMultipleConfigBackupOptions("Automations", "automations.yaml", "id", "alias")},
		},
		{
			GroupName: "Scenes",
			Configs:   []*ConfigBackupOptions{NewMultipleConfigBackupOptions("Scenes", "scenes.yaml", "id", "name")},
		},
		{
			GroupName: "ESP Home",
			Configs:   []*ConfigBackupOptions{NewDirectoryConfigBackupOptions("ESP Home", "esphome", []string{"*.yaml"}, []string{})},
		},
		{
			GroupName: "Dashboards",
			Configs: []*ConfigBackupOptions{NewDirectoryConfigBackupOptions("Storage", ".storage", []string{
				"lovelace.*",
				"lovelace_dashboards",
				"energy",
			}, []string{})},
		},
		{
			GroupName: "Helpers",
			Configs: []*ConfigBackupOptions{NewDirectoryConfigBackupOptions("Storage", ".storage", []string{
				"counter.*",
				"frontend.*",
				"input_boolean.*",
				"input_number.*",
				"input_select.*",
				"input_text.*",
				"input_*",
				"schedule",
				"timer",
			}, []string{})},
		},
	}

	appSettings := &AppSettings{
		HomeAssistantConfigDir:  "/homeassistant",
		BackupDir:               "/data/ha-config-history/backups",
		Port:                    ":40613",
		CronSchedule:            new(string),
		DefaultMaxBackups:       nil,
		DefaultMaxBackupAgeDays: nil,
		ConfigGroups:            defaultConfigGroups,
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

	if len(appSettings.Configs) > 0 && len(appSettings.ConfigGroups) > 0 {
		slog.Warn("Both old configs and new config groups exist. Using config groups and clearing old configs.")
		appSettings.Configs = nil
		return appSettings
	}

	// TODO: V2 - Remove migration code
	if len(appSettings.Configs) > 0 && len(appSettings.ConfigGroups) == 0 {
		slog.Info("Migrating configuration to grouped structure")

		migratedGroups := migrateToGroups(appSettings.Configs)
		if len(migratedGroups) == 0 {
			slog.Error("Migration failed: no groups created from existing configs")
			return appSettings
		}

		oldConfigs := appSettings.Configs
		appSettings.ConfigGroups = migratedGroups

		if err := saveConfig(configPath, appSettings); err != nil {
			slog.Error("Failed to save migrated configuration, reverting", "error", err)
			// Revert on save failure
			appSettings.ConfigGroups = nil
			appSettings.Configs = oldConfigs
			return appSettings
		}

		// Only clear old format after successful save
		appSettings.Configs = nil
		slog.Info("Successfully migrated configuration to grouped structure", "groups", len(appSettings.ConfigGroups))
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
