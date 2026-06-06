package io_test

import (
	"ha-config-history/internal/io"
	"ha-config-history/internal/types"
	"strings"
	"testing"
)

func Test_ReadKeyedConfigsFromSingleFile(t *testing.T) {
	t.Run("Creates one config per top-level key in a mapping file", func(t *testing.T) {
		backups, err := io.ReadKeyedConfigsFromSingleFile("test-data",
			types.NewKeyedConfigBackupOptions("sample-keyed.yaml", "alias"))
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(backups) != 2 {
			t.Fatalf("Expected 2 config backups, got: %d", len(backups))
		}

		// Entry order follows document order: morning_routine, then notify_me.
		first := backups[0]
		if first.ID != "morning_routine" {
			t.Errorf("Expected first ID to be the map key 'morning_routine', got: %s", first.ID)
		}
		if first.Path != "sample-keyed.yaml" {
			t.Errorf("Expected Path 'sample-keyed.yaml', got: %s", first.Path)
		}
		if first.BackupType != "keyed" {
			t.Errorf("Expected BackupType 'keyed', got: %s", first.BackupType)
		}
		// Friendly name comes from the value's alias field.
		if first.FriendlyName != "Morning Routine" {
			t.Errorf("Expected FriendlyName from alias 'Morning Routine', got: %s", first.FriendlyName)
		}
		if first.Hash == "" {
			t.Error("Expected a non-empty hash")
		}
		// Blob is the value body only — it must NOT contain the top-level key.
		firstBlob := string(first.Blob)
		if !strings.Contains(firstBlob, "alias: \"Morning Routine\"") {
			t.Errorf("Expected blob to contain the alias field, got:\n%s", firstBlob)
		}
		if !strings.Contains(firstBlob, "service: light.turn_on") {
			t.Errorf("Expected blob to contain the sequence body, got:\n%s", firstBlob)
		}
		if strings.Contains(firstBlob, "morning_routine") {
			t.Errorf("Blob should be the value only and not contain the key 'morning_routine', got:\n%s", firstBlob)
		}

		// Second entry has no alias -> friendly name falls back to the key.
		second := backups[1]
		if second.ID != "notify_me" {
			t.Errorf("Expected second ID 'notify_me', got: %s", second.ID)
		}
		if second.FriendlyName != "notify_me" {
			t.Errorf("Expected FriendlyName to fall back to the key 'notify_me', got: %s", second.FriendlyName)
		}
		if strings.Contains(string(second.Blob), "notify_me") {
			t.Errorf("Blob should be the value only and not contain the key 'notify_me', got:\n%s", string(second.Blob))
		}
	})

	t.Run("Returns zero backups for an empty mapping", func(t *testing.T) {
		backups, err := io.ReadKeyedConfigsFromSingleFile("test-data",
			types.NewKeyedConfigBackupOptions("sample-keyed-empty.yaml", "alias"))
		if err != nil {
			t.Fatalf("Expected no error for empty mapping, got: %v", err)
		}
		if len(backups) != 0 {
			t.Errorf("Expected 0 config backups for empty mapping, got: %d", len(backups))
		}
	})

	t.Run("Returns zero backups for a null root", func(t *testing.T) {
		backups, err := io.ReadKeyedConfigsFromSingleFile("test-data",
			types.NewKeyedConfigBackupOptions("sample-keyed-null.yaml", "alias"))
		if err != nil {
			t.Fatalf("Expected no error for null root, got: %v", err)
		}
		if len(backups) != 0 {
			t.Errorf("Expected 0 config backups for null root, got: %d", len(backups))
		}
	})
}
