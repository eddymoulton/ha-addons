package api_test

import (
	"ha-config-history/internal/api"
	"ha-config-history/internal/core"
	"ha-config-history/internal/types"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// setupKeyedFileEnv creates a test environment for keyed (mapping-rooted) restore tests.
func setupKeyedFileEnv(t *testing.T, configPath, friendlyNameNode string) *testEnvironment {
	tempDir, backupDir, haConfigDir := setupTestDirs(t)
	targetFile := filepath.Join(haConfigDir, configPath)

	config := &types.AppSettings{
		HomeAssistantConfigDir: haConfigDir,
		BackupDir:              backupDir,
		Port:                   ":8080",
		ConfigGroups: []*types.ConfigBackupOptionGroup{
			types.NewConfigBackupOptionGroup(
				"Test Scripts",
				[]*types.ConfigBackupOptions{
					{
						Path:             configPath,
						BackupType:       "keyed",
						FriendlyNameNode: &friendlyNameNode,
					},
				},
			),
		},
	}

	server := core.NewServer(config, "tmp/test-config.json")
	router := gin.New()
	router.POST("/configs/:group/:path/:id/backups/:filename/restore", api.RestoreBackupHandler(server))

	return &testEnvironment{
		tempDir:     tempDir,
		backupDir:   backupDir,
		haConfigDir: haConfigDir,
		targetFile:  targetFile,
		server:      server,
		router:      router,
		t:           t,
	}
}

func TestRestoreKeyedBackupHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupSlug, path := "test-scripts", "scripts.yaml"

	setupEnv := func(t *testing.T) *testEnvironment {
		env := setupKeyedFileEnv(t, "scripts.yaml", "alias")
		original := readFile(t, "test-data/keyed-scripts-original.yaml")
		env.writeFile(env.targetFile, original, 0644)
		return env
	}

	parseScripts := func(env *testEnvironment, content []byte) map[string]map[string]interface{} {
		scripts := map[string]map[string]interface{}{}
		if err := yaml.Unmarshal(content, &scripts); err != nil {
			env.t.Fatalf("failed to parse restored YAML: %v", err)
		}
		return scripts
	}

	t.Run("replaces an existing key and leaves others unchanged", func(t *testing.T) {
		env := setupEnv(t)
		id, filename := "script_2", "20240101T120000.backup"
		env.createBackup(groupSlug, path, id, filename, readFile(t, "test-data/keyed-script-2-modified.yaml"))

		w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
		env.assertStatusOK(w)
		env.assertRestoreSuccess(response)

		restored := env.readFile(env.targetFile)
		env.assertValidYAML(restored)

		scripts := parseScripts(env, restored)
		if len(scripts) != 3 {
			t.Fatalf("expected 3 scripts, got %d", len(scripts))
		}
		if scripts["script_2"]["alias"] != "Modified Second Script" {
			t.Errorf("script_2 was not updated, alias=%v", scripts["script_2"]["alias"])
		}
		if scripts["script_1"]["alias"] != "First Script" {
			t.Errorf("script_1 changed unexpectedly, alias=%v", scripts["script_1"]["alias"])
		}
		if scripts["script_3"]["alias"] != "Third Script" {
			t.Errorf("script_3 changed unexpectedly, alias=%v", scripts["script_3"]["alias"])
		}
	})

	t.Run("adds a new key when the script does not exist", func(t *testing.T) {
		env := setupEnv(t)
		id, filename := "script_4", "20240101T120000.backup"
		env.createBackup(groupSlug, path, id, filename, readFile(t, "test-data/keyed-script-4-new.yaml"))

		w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
		env.assertStatusOK(w)
		env.assertRestoreSuccess(response)

		restored := env.readFile(env.targetFile)
		env.assertValidYAML(restored)

		scripts := parseScripts(env, restored)
		if len(scripts) != 4 {
			t.Fatalf("expected 4 scripts after adding, got %d", len(scripts))
		}
		if scripts["script_4"]["alias"] != "Fourth Script" {
			t.Errorf("script_4 was not added correctly, alias=%v", scripts["script_4"]["alias"])
		}
		for _, k := range []string{"script_1", "script_2", "script_3"} {
			if _, ok := scripts[k]; !ok {
				t.Errorf("expected original script %s to remain", k)
			}
		}
	})

	t.Run("restores are idempotent", func(t *testing.T) {
		env := setupEnv(t)
		id, filename := "script_2", "20240101T120000.backup"
		env.createBackup(groupSlug, path, id, filename, readFile(t, "test-data/keyed-script-2-modified.yaml"))

		w1, r1 := env.makeRestoreRequest(groupSlug, path, id, filename)
		env.assertStatusOK(w1)
		env.assertRestoreSuccess(r1)
		first := env.readFile(env.targetFile)

		w2, r2 := env.makeRestoreRequest(groupSlug, path, id, filename)
		env.assertStatusOK(w2)
		env.assertRestoreSuccess(r2)
		second := env.readFile(env.targetFile)

		env.assertContentEquals(first, second)
	})

	t.Run("key order is preserved after restore", func(t *testing.T) {
		env := setupEnv(t)
		id, filename := "script_2", "20240101T120000.backup"
		env.createBackup(groupSlug, path, id, filename, readFile(t, "test-data/keyed-script-2-modified.yaml"))

		w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
		env.assertStatusOK(w)
		env.assertRestoreSuccess(response)

		restored := env.readFile(env.targetFile)

		var rootNode yaml.Node
		if err := yaml.Unmarshal(restored, &rootNode); err != nil {
			t.Fatalf("failed to parse restored YAML as yaml.Node: %v", err)
		}
		if len(rootNode.Content) == 0 {
			t.Fatal("expected a document node with content")
		}
		mappingNode := rootNode.Content[0]
		if mappingNode.Kind != yaml.MappingNode {
			t.Fatalf("expected root to be a MappingNode, got kind=%v", mappingNode.Kind)
		}

		// Collect top-level keys in document order (step 2: every other node starting at 0).
		var keys []string
		for i := 0; i+1 < len(mappingNode.Content); i += 2 {
			keys = append(keys, mappingNode.Content[i].Value)
		}

		expectedOrder := []string{"script_1", "script_2", "script_3"}
		if len(keys) != len(expectedOrder) {
			t.Fatalf("expected %d top-level keys, got %d: %v", len(expectedOrder), len(keys), keys)
		}
		for i, expected := range expectedOrder {
			if keys[i] != expected {
				t.Errorf("key at position %d: expected %q, got %q", i, expected, keys[i])
			}
		}
	})
}
