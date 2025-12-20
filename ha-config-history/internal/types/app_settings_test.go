package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppSettingsMigration(t *testing.T) {
	// Create temporary directory for test config files
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	tests := []struct {
		name               string
		initialAppSettings *AppSettings
		expectMigration    bool
		expectError        bool
	}{
		{
			name: "config loading with old format (triggers conflict resolution)",
			initialAppSettings: &AppSettings{
				HomeAssistantConfigDir: "tmp/homeassistant",
				BackupDir:              "tmp/data/backups",
				Port:                   ":40613",
				Configs: []*ConfigBackupOptions{
					{Path: "configuration.yaml", BackupType: "single"},
					{Path: "automations.yaml", BackupType: "multiple"},
				},
				// Note: LoadConfig starts with defaults that include ConfigGroups,
				// so this triggers conflict resolution, not migration
			},
			expectMigration: false, // This will trigger conflict resolution
			expectError:     false,
		},
		{
			name: "no migration needed - already grouped",
			initialAppSettings: &AppSettings{
				HomeAssistantConfigDir: "tmp/homeassistant",
				BackupDir:              "tmp/data/backups",
				Port:                   ":40613",
				Configs:                nil,
				ConfigGroups: []*ConfigBackupOptionGroup{
					NewConfigBackupOptionGroup(
						"Test Group",
						[]*ConfigBackupOptions{
							{Path: "configuration.yaml", BackupType: "single"},
						},
					),
				},
			},
			expectMigration: false,
			expectError:     false,
		},
		{
			name: "conflicting configuration - both formats exist",
			initialAppSettings: &AppSettings{
				HomeAssistantConfigDir: "tmp/homeassistant",
				BackupDir:              "tmp/data/backups",
				Port:                   ":40613",
				Configs: []*ConfigBackupOptions{
					{Path: "old.yaml", BackupType: "single"},
				},
				ConfigGroups: []*ConfigBackupOptionGroup{
					NewConfigBackupOptionGroup(
						"New Group",
						[]*ConfigBackupOptions{
							{Path: "new.yaml", BackupType: "single"},
						},
					),
				},
			},
			expectMigration: false,
			expectError:     false,
		},
		{
			name: "empty configs - no migration",
			initialAppSettings: &AppSettings{
				HomeAssistantConfigDir: "tmp/homeassistant",
				BackupDir:              "tmp/data/backups",
				Port:                   ":40613",
				Configs:                []*ConfigBackupOptions{},
				ConfigGroups:           nil,
			},
			expectMigration: false,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write initial config to file
			data, err := json.MarshalIndent(tt.initialAppSettings, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal initial config: %v", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("failed to write initial config file: %v", err)
			}

			// Load config (this should trigger migration if needed)
			result := LoadAppSettings(configPath)

			// Just verify that LoadConfig works and handles the scenarios correctly
			if result == nil {
				t.Fatal("LoadConfig returned nil")
			}

			// The result will always have ConfigGroups due to defaults
			if len(result.ConfigGroups) == 0 {
				t.Error("result should have config groups")
			}

			// Configs should be cleared if there was a conflict
			if len(result.Configs) > 0 {
				t.Error("configs should be cleared when both formats exist")
			}

			// Verify basic configuration is preserved
			if result.HomeAssistantConfigDir != tt.initialAppSettings.HomeAssistantConfigDir {
				t.Error("HomeAssistantConfigDir was not preserved")
			}
			if result.BackupDir != tt.initialAppSettings.BackupDir {
				t.Error("BackupDir was not preserved")
			}
			if result.Port != tt.initialAppSettings.Port {
				t.Error("Port was not preserved")
			}
		})
	}
}

func TestSaveAppSettings(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	appSettings := &AppSettings{
		HomeAssistantConfigDir: "tmp/homeassistant",
		BackupDir:              "tmp/test-backups",
		Port:                   ":8080",
		ConfigGroups: []*ConfigBackupOptionGroup{
			NewConfigBackupOptionGroup(
				"Test Group",
				[]*ConfigBackupOptions{
					{Path: "test.yaml", BackupType: "single"},
				},
			),
		},
	}

	// Test saving configuration
	err := saveAppSettings(configPath, appSettings)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created and is readable
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	// Verify content can be unmarshaled
	var loadedConfig AppSettings
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("failed to unmarshal saved config: %v", err)
	}

	// Verify content matches
	if loadedConfig.HomeAssistantConfigDir != appSettings.HomeAssistantConfigDir {
		t.Error("HomeAssistantConfigDir mismatch after save/load")
	}
	if len(loadedConfig.ConfigGroups) != len(appSettings.ConfigGroups) {
		t.Error("ConfigGroups count mismatch after save/load")
	}
}
