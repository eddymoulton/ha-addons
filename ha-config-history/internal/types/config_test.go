package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateToGroups(t *testing.T) {
	tests := []struct {
		name     string
		configs  []*ConfigBackupOptions
		expected []string // expected group names
	}{
		{
			name:     "empty configs",
			configs:  []*ConfigBackupOptions{},
			expected: []string{},
		},
		{
			name: "core configuration",
			configs: []*ConfigBackupOptions{
				{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
			},
			expected: []string{"Core Home Assistant"},
		},
		{
			name: "automations and scenes",
			configs: []*ConfigBackupOptions{
				{Name: "Automations", Path: "automations.yaml", BackupType: "multiple"},
				{Name: "Scenes", Path: "scenes.yaml", BackupType: "multiple"},
			},
			expected: []string{"Automations", "Scenes"},
		},
		{
			name: "mixed configurations",
			configs: []*ConfigBackupOptions{
				{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
				{Name: "Automations", Path: "automations.yaml", BackupType: "multiple"},
				{Name: "Custom Script", Path: "scripts/custom.yaml", BackupType: "single"},
			},
			expected: []string{"Core Home Assistant", "Automations", "Custom Script"},
		},
		{
			name: "storage configurations",
			configs: []*ConfigBackupOptions{
				{Name: "Storage", Path: ".storage", BackupType: "directory"},
				{Name: "Core Storage", Path: ".storage/core.config", BackupType: "single"},
			},
			expected: []string{"Storage & Settings"},
		},
		{
			name: "esphome configurations",
			configs: []*ConfigBackupOptions{
				{Name: "ESP Home", Path: "esphome", BackupType: "directory"},
				{Name: "ESP Device", Path: "esphome/device.yaml", BackupType: "single"},
			},
			expected: []string{"ESP Home"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := migrateToGroups(tt.configs)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d groups, got %d", len(tt.expected), len(result))
				return
			}

			// Check that all expected group names exist
			groupNames := make(map[string]bool)
			for _, group := range result {
				groupNames[group.GroupName] = true
			}

			for _, expectedName := range tt.expected {
				if !groupNames[expectedName] {
					t.Errorf("expected group '%s' not found", expectedName)
				}
			}

			// Verify configs are properly assigned
			totalConfigs := 0
			for _, group := range result {
				totalConfigs += len(group.Configs)
				if len(group.Configs) == 0 {
					t.Errorf("group '%s' has no configs", group.GroupName)
				}
			}

			if totalConfigs != len(tt.configs) {
				t.Errorf("expected %d total configs, got %d", len(tt.configs), totalConfigs)
			}
		})
	}
}

func TestConfigMigration(t *testing.T) {
	// Create temporary directory for test config files
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	tests := []struct {
		name            string
		initialConfig   *AppSettings
		expectMigration bool
		expectError     bool
	}{
		{
			name: "config loading with old format (triggers conflict resolution)",
			initialConfig: &AppSettings{
				HomeAssistantConfigDir: "/homeassistant",
				BackupDir:              "/data/backups",
				Port:                   ":40613",
				Configs: []*ConfigBackupOptions{
					{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					{Name: "Automations", Path: "automations.yaml", BackupType: "multiple"},
				},
				// Note: LoadConfig starts with defaults that include ConfigGroups,
				// so this triggers conflict resolution, not migration
			},
			expectMigration: false, // This will trigger conflict resolution
			expectError:     false,
		},
		{
			name: "no migration needed - already grouped",
			initialConfig: &AppSettings{
				HomeAssistantConfigDir: "/homeassistant",
				BackupDir:              "/data/backups",
				Port:                   ":40613",
				Configs:                nil,
				ConfigGroups: []*ConfigBackupOptionGroup{
					{
						GroupName: "Test Group",
						Configs: []*ConfigBackupOptions{
							{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
						},
					},
				},
			},
			expectMigration: false,
			expectError:     false,
		},
		{
			name: "conflicting configuration - both formats exist",
			initialConfig: &AppSettings{
				HomeAssistantConfigDir: "/homeassistant",
				BackupDir:              "/data/backups",
				Port:                   ":40613",
				Configs: []*ConfigBackupOptions{
					{Name: "Old Config", Path: "old.yaml", BackupType: "single"},
				},
				ConfigGroups: []*ConfigBackupOptionGroup{
					{
						GroupName: "New Group",
						Configs: []*ConfigBackupOptions{
							{Name: "New Config", Path: "new.yaml", BackupType: "single"},
						},
					},
				},
			},
			expectMigration: false,
			expectError:     false,
		},
		{
			name: "empty configs - no migration",
			initialConfig: &AppSettings{
				HomeAssistantConfigDir: "/homeassistant",
				BackupDir:              "/data/backups",
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
			data, err := json.MarshalIndent(tt.initialConfig, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal initial config: %v", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("failed to write initial config file: %v", err)
			}

			// Load config (this should trigger migration if needed)
			result := LoadConfig(configPath)

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
			if result.HomeAssistantConfigDir != tt.initialConfig.HomeAssistantConfigDir {
				t.Error("HomeAssistantConfigDir was not preserved")
			}
			if result.BackupDir != tt.initialConfig.BackupDir {
				t.Error("BackupDir was not preserved")
			}
			if result.Port != tt.initialConfig.Port {
				t.Error("Port was not preserved")
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	config := &AppSettings{
		HomeAssistantConfigDir: "/test",
		BackupDir:              "/test-backups",
		Port:                   ":8080",
		ConfigGroups: []*ConfigBackupOptionGroup{
			{
				GroupName: "Test Group",
				Configs: []*ConfigBackupOptions{
					{Name: "Test Config", Path: "test.yaml", BackupType: "single"},
				},
			},
		},
	}

	// Test saving configuration
	err := saveConfig(configPath, config)
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
	if loadedConfig.HomeAssistantConfigDir != config.HomeAssistantConfigDir {
		t.Error("HomeAssistantConfigDir mismatch after save/load")
	}
	if len(loadedConfig.ConfigGroups) != len(config.ConfigGroups) {
		t.Error("ConfigGroups count mismatch after save/load")
	}
}

func TestMigrationRobustness(t *testing.T) {
	tests := []struct {
		name        string
		configs     []*ConfigBackupOptions
		shouldPanic bool
	}{
		{
			name:        "nil configs slice",
			configs:     nil,
			shouldPanic: false,
		},
		{
			name: "configs with nil elements",
			configs: []*ConfigBackupOptions{
				{Name: "Valid Config", Path: "valid.yaml", BackupType: "single"},
				nil, // This should not cause a panic after our fix
			},
			shouldPanic: false, // We handle nil configs gracefully now
		},
		{
			name: "configs with empty names",
			configs: []*ConfigBackupOptions{
				{Name: "", Path: "empty-name.yaml", BackupType: "single"},
				{Name: "Valid Config", Path: "valid.yaml", BackupType: "single"},
			},
			shouldPanic: false,
		},
		{
			name: "configs with empty paths",
			configs: []*ConfigBackupOptions{
				{Name: "Empty Path Config", Path: "", BackupType: "single"},
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.shouldPanic {
					if tt.shouldPanic {
						t.Error("expected panic but none occurred")
					} else {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()

			result := migrateToGroups(tt.configs)

			if !tt.shouldPanic {
				// Basic validation for non-panicking cases
				if result == nil {
					t.Error("migrateToGroups returned nil")
				}
			}
		})
	}
}

// TestDirectMigration tests the migration logic in isolation
func TestDirectMigration(t *testing.T) {
	// Create temporary directory for test config files
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Test direct migration scenario
	configs := []*ConfigBackupOptions{
		{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
		{Name: "Automations", Path: "automations.yaml", BackupType: "multiple"},
		{Name: "Custom Script", Path: "scripts/custom.yaml", BackupType: "single"},
	}

	// Create a config with only old format (no defaults)
	appSettings := &AppSettings{
		HomeAssistantConfigDir: "/homeassistant",
		BackupDir:              "/data/backups",
		Port:                   ":40613",
		Configs:                configs,
		ConfigGroups:           nil, // No groups initially
	}

	// Write to file
	data, err := json.MarshalIndent(appSettings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Now test migration logic by simulating what happens in LoadConfig
	// Read the file back
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	// Start with empty settings (not the defaults)
	testSettings := &AppSettings{}
	if err := json.Unmarshal(fileData, testSettings); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Now check if migration should occur
	if len(testSettings.Configs) > 0 && len(testSettings.ConfigGroups) == 0 {
		t.Log("Migration should trigger")

		// Perform migration
		migratedGroups := migrateToGroups(testSettings.Configs)
		if len(migratedGroups) == 0 {
			t.Fatal("Migration failed: no groups created from existing configs")
		}

		// Backup old configuration before clearing
		oldConfigs := testSettings.Configs
		testSettings.ConfigGroups = migratedGroups

		// Save migrated configuration
		if err := saveConfig(configPath, testSettings); err != nil {
			t.Fatalf("Failed to save migrated configuration: %v", err)
		}

		// Only clear old format after successful save
		testSettings.Configs = nil

		// Verify migration results
		if len(testSettings.ConfigGroups) == 0 {
			t.Error("expected migration to create config groups")
		}
		if len(testSettings.Configs) > 0 {
			t.Error("expected old configs to be cleared after migration")
		}

		// Verify the saved file
		savedData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read saved config: %v", err)
		}

		var savedConfig AppSettings
		if err := json.Unmarshal(savedData, &savedConfig); err != nil {
			t.Fatalf("failed to unmarshal saved config: %v", err)
		}

		if len(savedConfig.ConfigGroups) == 0 {
			t.Error("migration was not saved to file")
		}

		// Verify all configs were migrated
		totalConfigs := 0
		for _, group := range savedConfig.ConfigGroups {
			totalConfigs += len(group.Configs)
		}
		if totalConfigs != len(oldConfigs) {
			t.Errorf("expected %d total configs in groups, got %d", len(oldConfigs), totalConfigs)
		}

		t.Logf("Successfully migrated %d configs to %d groups", len(oldConfigs), len(savedConfig.ConfigGroups))

	} else {
		t.Fatal("Migration conditions not met")
	}
}
