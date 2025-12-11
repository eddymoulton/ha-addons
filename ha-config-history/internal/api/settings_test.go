package api

import (
	"ha-config-history/internal/types"
	"testing"
)

func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		expectErr bool
	}{
		{"valid name", "Core Home Assistant", false},
		{"empty name", "", true},
		{"whitespace only", "   ", true},
		{"too long", string(make([]byte, 101)), true},
		{"invalid characters colon", "Group:Name", true},
		{"invalid characters slash", "Group/Name", true},
		{"invalid characters backslash", "Group\\Name", true},
		{"invalid characters pipe", "Group|Name", true},
		{"invalid characters asterisk", "Group*Name", true},
		{"invalid characters question", "Group?Name", true},
		{"reserved name admin", "admin", true},
		{"reserved name Admin", "Admin", true},
		{"reserved name ADMIN", "ADMIN", true},
		{"reserved name null", "null", true},
		{"reserved name undefined", "undefined", true},
		{"reserved name root", "root", true},
		{"reserved name system", "system", true},
		{"valid special characters", "Group-Name_123", false},
		{"valid with spaces", "My Config Group", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupName(tt.groupName)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateGroupName(%q) error = %v, expectErr %v", tt.groupName, err, tt.expectErr)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{"valid relative path", "configuration.yaml", false},
		{"valid nested path", "config/automations.yaml", false},
		{"valid homeassistant absolute", "/homeassistant/config.yaml", false},
		{"empty path", "", true},
		{"whitespace only", "   ", true},
		{"too long", string(make([]byte, 501)), true},
		{"directory traversal", "../config.yaml", true},
		{"directory traversal nested", "config/../secrets.yaml", true},
		{"invalid absolute path", "/etc/passwd", true},
		{"valid with spaces", "my config file.yaml", false},
		{"hidden file", ".storage/core.config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigPath(tt.path)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateConfigPath(%q) error = %v, expectErr %v", tt.path, err, tt.expectErr)
			}
		})
	}
}

func TestValidateConfigGroups(t *testing.T) {
	tests := []struct {
		name         string
		configGroups []*types.ConfigBackupOptionGroup
		expectErr    bool
	}{
		{
			name:         "empty groups",
			configGroups: []*types.ConfigBackupOptionGroup{},
			expectErr:    false,
		},
		{
			name:         "nil groups",
			configGroups: nil,
			expectErr:    false,
		},
		{
			name: "valid single group",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "valid multiple groups",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					},
				},
				{
					GroupName: "Automations",
					Configs: []*types.ConfigBackupOptions{
						{
							Name:             "Automations",
							Path:             "automations.yaml",
							BackupType:       "multiple",
							IdNode:           stringPtr("id"),
							FriendlyNameNode: stringPtr("alias"),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "nil group",
			configGroups: []*types.ConfigBackupOptionGroup{
				nil,
			},
			expectErr: true,
		},
		{
			name: "invalid group name",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "admin",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate group names",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					},
				},
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Secrets", Path: "secrets.yaml", BackupType: "single"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "empty group configs",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Empty Group",
					Configs:   []*types.ConfigBackupOptions{},
				},
			},
			expectErr: true,
		},
		{
			name: "nil config",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						nil,
					},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate config paths",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Group 1",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "single"},
					},
				},
				{
					GroupName: "Group 2",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration Copy", Path: "configuration.yaml", BackupType: "single"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid backup type",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "Configuration", Path: "configuration.yaml", BackupType: "invalid"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "empty config name",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{Name: "", Path: "configuration.yaml", BackupType: "single"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "multiple backup type missing idNode",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Automations",
					Configs: []*types.ConfigBackupOptions{
						{
							Name:             "Automations",
							Path:             "automations.yaml",
							BackupType:       "multiple",
							FriendlyNameNode: stringPtr("alias"),
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "multiple backup type missing friendlyNameNode",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Automations",
					Configs: []*types.ConfigBackupOptions{
						{
							Name:       "Automations",
							Path:       "automations.yaml",
							BackupType: "multiple",
							IdNode:     stringPtr("id"),
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid maxBackups",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{
							Name:       "Configuration",
							Path:       "configuration.yaml",
							BackupType: "single",
							MaxBackups: intPtr(0),
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid maxBackupAgeDays",
			configGroups: []*types.ConfigBackupOptionGroup{
				{
					GroupName: "Core Home Assistant",
					Configs: []*types.ConfigBackupOptions{
						{
							Name:             "Configuration",
							Path:             "configuration.yaml",
							BackupType:       "single",
							MaxBackupAgeDays: intPtr(0),
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigGroups(tt.configGroups)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateConfigGroups() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
