package io_test

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_ReadMultipleConfigsFromSingleFile(t *testing.T) {
	t.Run("Creates multiple configs from single file", func(t *testing.T) {
		fileName := "sample-multi.yaml"

		expected := []*types.ConfigBackup{
			{
				ConfigIdentifier: types.ConfigIdentifier{
					ID:    "example-1",
					Group: fileName,
				},
				FriendlyName: "Sample Multi Example 1",
				BackupType:   "multiple",
				FilePath:     "test-data/" + fileName,
				Hash:         "qJQn1jJ7Lx7ga0jRDb6560qQnU4=",
				Blob: []uint8(`id: "example-1"
alias: "Sample Multi Example 1"
description: "An example demonstrating multiple configurations."
config:
    settingA: true
    settingB: "valueA"
    nestedConfig:
        option1: 10
        option2: [1, 2, 3]
tags:
    - "example"
    - "multi-config"
    - "yaml"
`),
			},
			{
				ConfigIdentifier: types.ConfigIdentifier{
					ID:    "example-2",
					Group: fileName,
				},
				FriendlyName: "Sample Multi Example 2",
				BackupType:   "multiple",
				FilePath:     "test-data/" + fileName,
				Hash:         "nrgwnIiZKhZdxqYOgmkmkRVwyn0=",
				Blob: []uint8(`id: "example-2"
alias: "Sample Multi Example 2"
description: "An example demonstrating multiple configurations."
config:
    settingA: true
    settingB: "valueB"
    nestedConfig:
        option1: 10
        option2: [1, 2, 3]
tags:
    - "example"
    - "multi-config"
    - "yaml"
`),
			},
		}

		configBackups, err := io.ReadMultipleConfigsFromSingleFile("test-data",
			types.NewMultipleConfigBackupOptions(
				"multiple",
				fileName,
				"id",
				"alias",
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 2 {
			t.Errorf("Expected 2 config backups, got: %d", len(configBackups))
		}

		if diff := cmp.Diff(expected, configBackups, cmpopts.IgnoreFields(types.ConfigBackup{}, "ModifiedDate")); diff != "" {
			t.Errorf("Config backups do not match expected:\n%s", diff)
		}
	})
}

func Test_ReadSingleConfigFromSingleFile(t *testing.T) {
	t.Run("Creates single config from single file", func(t *testing.T) {
		fileName := "sample-single.yaml"

		expected :=
			&types.ConfigBackup{
				ConfigIdentifier: types.ConfigIdentifier{
					ID:    fileName,
					Group: fileName,
				},
				FriendlyName: fileName,
				BackupType:   "single",
				FilePath:     "test-data/" + fileName,
				Hash:         "0H41D6gI8C3iCdjTHiELk6-vA8g=",
				Blob: []uint8(`description: "An example of a single configuration file, with no name or id"
config:
  settingA: true
  settingB: "valueB"
  nestedConfig:
    option1: 10
    option2: [1, 2, 3]
tags:
  - "example"
  - "multi-config"
  - "yaml"
`),
			}

		configBackups, err := io.ReadSingleConfigFromSingleFile("test-data",
			types.NewSingleConfigBackupOptions(
				"multiple",
				fileName,
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if diff := cmp.Diff(expected, configBackups, cmpopts.IgnoreFields(types.ConfigBackup{}, "ModifiedDate")); diff != "" {
			t.Errorf("Config backups do not match expected:\n%s", diff)
		}
	})
}

func Test_ReadMultipleConfigsFromDirectory(t *testing.T) {
	t.Run("Creates multiple configs from a directory of files", func(t *testing.T) {
		directoryName := "sample-dir"

		expected := []*types.ConfigBackup{
			{
				ConfigIdentifier: types.ConfigIdentifier{ID: "random-file", Group: "sample-dir"},
				FriendlyName:     "random-file",
				Hash:             "M_oQwxaUJMNm2qj-MCmIQ_DBZ7w=",
				BackupType:       "directory",
				FilePath:         "test-data/sample-dir/random-file",
				Blob:             []uint8("This isn't yaml"),
			},
			{
				ConfigIdentifier: types.ConfigIdentifier{
					ID:    "sample-single-id.yaml",
					Group: directoryName,
				},
				FriendlyName: "sample-single-id.yaml",
				BackupType:   "directory",
				FilePath:     "test-data/" + directoryName + "/sample-single-id.yaml",
				Hash:         "kFKAAmZ2efLN5BnpAo2CD5wNnd8=",
				Blob: []uint8(
					`description: "An example of a single configuration file in a folder, with an id"
config:
  settingA: true
  settingB: "valueB"
  nestedConfig:
    option1: 10
    option2: [1, 2, 3]
tags:
  - "example"
  - "multi-config"
  - "yaml"
`),
			},
			{
				ConfigIdentifier: types.ConfigIdentifier{
					ID:    "sample-single.yaml",
					Group: directoryName,
				},
				FriendlyName: "sample-single.yaml",
				BackupType:   "directory",
				FilePath:     "test-data/" + directoryName + "/sample-single.yaml",
				Hash:         "4MkaqYb9oq4_zGrbJgePLeHc35A=",
				Blob: []uint8(`description: "An example of a single configuration file in a folder, with no id"
config:
  settingA: true
  settingB: "valueB"
  nestedConfig:
    option1: 10
    option2: [1, 2, 3]
tags:
  - "example"
  - "multi-config"
  - "yaml"
`),
			},
		}

		configBackups, err := io.ReadMultipleConfigsFromDirectory(
			"test-data",
			types.NewDirectoryConfigBackupOptions(
				"multiple",
				directoryName,
				[]string{},
				[]string{},
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 3 {
			t.Errorf("Expected 3 config backups, got: %d", len(configBackups))
		}

		if diff := cmp.Diff(expected, configBackups, cmpopts.IgnoreFields(types.ConfigBackup{}, "ModifiedDate")); diff != "" {
			t.Errorf("Config backups do not match expected:\n%s", diff)
		}
	})

	t.Run("Returns empty slice when no matching files in directory", func(t *testing.T) {
		directoryName := "sample-dir"

		configBackups, err := io.ReadMultipleConfigsFromDirectory(
			"test-data",
			types.NewDirectoryConfigBackupOptions(
				"multiple",
				directoryName,
				[]string{"*.nonexistent"},
				[]string{},
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 0 {
			t.Errorf("Expected 0 config backups, got: %d", len(configBackups))
		}
	})

	t.Run("Only includes files matching include patterns", func(t *testing.T) {
		directoryName := "sample-dir"

		configBackups, err := io.ReadMultipleConfigsFromDirectory(
			"test-data",
			types.NewDirectoryConfigBackupOptions(
				"multiple",
				directoryName,
				[]string{"*id.yaml"},
				[]string{},
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 1 {
			t.Errorf("Expected 1 config backup, got: %d", len(configBackups))
		}

		if configBackups[0].ID != "sample-single-id.yaml" {
			t.Errorf("Expected config backup ID to be 'sample-single-id.yaml', got: %s", configBackups[0].ID)
		}
	})

	t.Run("Excludes files matching exclude patterns", func(t *testing.T) {
		directoryName := "sample-dir"

		configBackups, err := io.ReadMultipleConfigsFromDirectory(
			"test-data",
			types.NewDirectoryConfigBackupOptions(
				"multiple",
				directoryName,
				[]string{},
				[]string{"*id.yaml", "random-file"},
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 1 {
			t.Errorf("Expected 1 config backup, got: %d", len(configBackups))
		}

		if configBackups[0].ID != "sample-single.yaml" {
			t.Errorf("Expected config backup ID to be 'sample-single.yaml', got: %s", configBackups[0].ID)
		}
	})

	t.Run("Includes and then excludes files based on patterns", func(t *testing.T) {
		directoryName := "sample-dir"

		configBackups, err := io.ReadMultipleConfigsFromDirectory(
			"test-data",
			types.NewDirectoryConfigBackupOptions(
				"multiple",
				directoryName,
				[]string{"*.yaml"},
				[]string{"*id.yaml"},
			))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(configBackups) != 1 {
			t.Errorf("Expected 1 config backup, got: %d", len(configBackups))
		}

		if configBackups[0].ID != "sample-single.yaml" {
			t.Errorf("Expected config backup ID to be 'sample-single.yaml', got: %s", configBackups[0].ID)
		}
	})
}
