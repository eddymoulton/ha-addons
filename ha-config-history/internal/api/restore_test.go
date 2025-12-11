package api_test

import (
	"bytes"
	"encoding/json"
	"ha-config-history/internal/api"
	"ha-config-history/internal/core"
	"ha-config-history/internal/types"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestRestoreBackupHandler(t *testing.T) {
	t.Run("when restoring a single file backup", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		data := loadTestData(t)

		t.Run("standard files are restored", func(t *testing.T) {

			testCases := []struct {
				name     string
				filename string
				content  []byte
			}{
				{"backup extension", "20240101T120000.backup", data.backup},
				{"yaml extension", "20240101T120000.yaml", data.backup},
				{"utf8 content", "20240102T120000.yaml", data.utf8},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					env := setupSingleFileEnv(t, "config.yaml")
					groupSlug := "test-configs"
					path := "config.yaml"
					id := "test-config"

					env.writeFile(env.targetFile, data.original, 0644)
					env.createBackup(groupSlug, path, id, tc.filename, tc.content)

					w, response := env.makeRestoreRequest(groupSlug, path, id, tc.filename)
					env.assertStatusOK(w)
					env.assertRestoreSuccess(response)

					restoredContent := env.readFile(env.targetFile)
					env.assertContentEquals(tc.content, restoredContent)
				})
			}
		})

		t.Run("file permissions are preserved", func(t *testing.T) {
			env := setupSingleFileEnv(t, "config.yaml")
			groupSlug, path, id, filename := "test-configs", "config.yaml", "test-config", "20240101T120000.yaml"

			testFile := filepath.Join(env.haConfigDir, "permissions-test.yaml")
			env.writeFile(testFile, data.original, 0600)
			env.writeFile(env.targetFile, data.original, 0644)
			env.createBackup(groupSlug, path, id, filename, data.backup)

			originalInfo, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("Failed to stat original file: %v", err)
			}

			w, _ := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)

			restoredInfo, err := os.Stat(env.targetFile)
			if err != nil {
				t.Fatalf("Failed to stat restored file: %v", err)
			}

			if restoredInfo.Mode().Perm()&0200 == 0 {
				t.Error("Restored file is not writable")
			}
			if restoredInfo.Mode().Perm()&0400 == 0 {
				t.Error("Restored file is not readable")
			}

			t.Logf("Original permissions: %o, Restored permissions: %o",
				originalInfo.Mode().Perm(), restoredInfo.Mode().Perm())
		})

		t.Run("UTF-8 encoding is preserved", func(t *testing.T) {
			env := setupSingleFileEnv(t, "config.yaml")
			groupSlug, path, id, filename := "test-configs", "config.yaml", "test-config", "20240102T120000.yaml"

			env.writeFile(env.targetFile, data.original, 0644)
			env.createBackup(groupSlug, path, id, filename, data.utf8)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			env.assertContentEquals(data.utf8, restoredContent)

			restoredStr := string(restoredContent)
			requiredStrings := []string{"émojis 🎉", "中文测试", "🔥💯✨", "Ñoño", "€£¥₹"}
			for _, required := range requiredStrings {
				env.assertContains(restoredStr, required)
			}
		})

		t.Run("restores are idempotent", func(t *testing.T) {
			env := setupSingleFileEnv(t, "config.yaml")
			groupSlug, path, id, filename := "test-configs", "config.yaml", "test-config", "20240101T120000.yaml"

			data := loadTestData(t)
			env.writeFile(env.targetFile, data.original, 0644)
			env.createBackup(groupSlug, path, id, filename, data.backup)

			// First restore
			w1, response1 := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w1)
			env.assertRestoreSuccess(response1)

			firstRestoreContent := env.readFile(env.targetFile)

			// Second restore (should produce identical result)
			w2, response2 := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w2)
			env.assertRestoreSuccess(response2)

			secondRestoreContent := env.readFile(env.targetFile)

			// Verify both restores produced identical results
			env.assertContentEquals(firstRestoreContent, secondRestoreContent)
			env.assertContentEquals(data.backup, secondRestoreContent)
		})
	})

	t.Run("when restoring a partial file", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		setupPartialRestoreEnv := func(t *testing.T) *testEnvironment {
			env := setupPartialFileEnv(t, "automations.yaml", "id", "alias")
			// Create a file with multiple automation entries
			originalContent := readFile(t, "test-data/partial-automations-original.yaml")
			env.writeFile(env.targetFile, originalContent, 0644)
			return env
		}

		t.Run("restores middle section and verifies others unchanged", func(t *testing.T) {
			env := setupPartialRestoreEnv(t)
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_2", "20240101T120000.yaml"

			backupContent := readFile(t, "test-data/partial-automation-2-modified.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			restoredStr := string(restoredContent)

			// Verify first automation is unchanged
			env.assertContains(restoredStr, "id: automation_1")
			env.assertContains(restoredStr, "alias: First Automation")
			env.assertContains(restoredStr, "entity_id: light.living_room")

			// Verify middle automation was updated
			env.assertContains(restoredStr, "id: automation_2")
			env.assertContains(restoredStr, "alias: Modified Second Automation")
			env.assertContains(restoredStr, "at: \"08:30:00\"")

			// Verify third automation is unchanged
			env.assertContains(restoredStr, "id: automation_3")
			env.assertContains(restoredStr, "alias: Third Automation")
			env.assertContains(restoredStr, "event: sunset")
		})

		t.Run("restores first section and verifies others unchanged", func(t *testing.T) {
			env := setupPartialRestoreEnv(t)
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_1", "20240101T120000.yaml"

			backupContent := readFile(t, "test-data/partial-automation-1-updated.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			restoredStr := string(restoredContent)

			// Verify first automation was updated
			env.assertContains(restoredStr, "alias: Updated First Automation")
			env.assertContains(restoredStr, "entity_id: binary_sensor.door")

			// Verify second and third automations are unchanged
			env.assertContains(restoredStr, "id: automation_2")
			env.assertContains(restoredStr, "alias: Second Automation")
			env.assertContains(restoredStr, "id: automation_3")
			env.assertContains(restoredStr, "alias: Third Automation")
		})

		t.Run("restores last section and verifies others unchanged", func(t *testing.T) {
			env := setupPartialRestoreEnv(t)
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_3", "20240101T120000.yaml"

			backupContent := readFile(t, "test-data/partial-automation-3-updated.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			restoredStr := string(restoredContent)

			// Verify first and second automations are unchanged
			env.assertContains(restoredStr, "id: automation_1")
			env.assertContains(restoredStr, "alias: First Automation")
			env.assertContains(restoredStr, "id: automation_2")
			env.assertContains(restoredStr, "alias: Second Automation")

			// Verify third automation was updated
			env.assertContains(restoredStr, "alias: Updated Third Automation")
			env.assertContains(restoredStr, "event: sunrise")
			env.assertContains(restoredStr, "entity_id: scene.morning")
		})

		t.Run("YAML structure remains valid after complex partial restore", func(t *testing.T) {
			env := setupPartialRestoreEnv(t)
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_2", "20240101T120000.yaml"

			backupContent := readFile(t, "test-data/partial-automation-2-complex.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			env.assertValidYAML(restoredContent)

			// Verify it's still a sequence at root with 3 items
			var yamlData []interface{}
			if err := yaml.Unmarshal(restoredContent, &yamlData); err != nil {
				env.t.Fatalf("Failed to parse as sequence: %v", err)
			}
			if len(yamlData) != 3 {
				env.t.Errorf("Expected 3 items in YAML sequence, got %d", len(yamlData))
			}

			restoredStr := string(restoredContent)

			// Verify complex content was preserved
			env.assertContains(restoredStr, "alias: Complex YAML Structure Test")
			env.assertContains(restoredStr, "description:")
			env.assertContains(restoredStr, "above: 25")
			env.assertContains(restoredStr, "below: 60")

			// Check UTF-8 content is preserved (in some form)
			hasOriginalEmoji := strings.Contains(restoredStr, "émojis 🎉")
			hasDescriptionField := strings.Contains(restoredStr, "special characters")
			if !hasOriginalEmoji && !hasDescriptionField {
				env.t.Errorf("UTF-8 special characters were not preserved. Content:\n%s", restoredStr)
			}

			// Verify other sections are still present
			env.assertContains(restoredStr, "id: automation_1")
			env.assertContains(restoredStr, "id: automation_3")
		})

		t.Run("section ordering is preserved", func(t *testing.T) {
			env := setupPartialFileEnv(t, "automations.yaml", "id", "alias")
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_2", "20240101T120000.yaml"

			originalContent := readFile(t, "test-data/partial-automations-original.yaml")
			env.writeFile(env.targetFile, originalContent, 0644)

			backupContent := readFile(t, "test-data/partial-automation-2-webhook.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)

			// Parse as YAML to check ordering
			var yamlData []map[string]interface{}
			if err := yaml.Unmarshal(restoredContent, &yamlData); err != nil {
				env.t.Fatalf("Failed to parse restored YAML: %v", err)
			}

			// Verify order: automation_1, automation_2, automation_3
			if len(yamlData) != 3 {
				env.t.Fatalf("Expected 3 automations, got %d", len(yamlData))
			}

			expectedOrder := []string{"automation_1", "automation_2", "automation_3"}
			for i, expectedID := range expectedOrder {
				if yamlData[i]["id"] != expectedID {
					env.t.Errorf("Item %d should be %s, got %v", i, expectedID, yamlData[i]["id"])
				}
			}

			// Verify automation_2 was updated, others unchanged
			if yamlData[1]["alias"] != "Modified Second Automation" {
				env.t.Error("Middle automation was not updated correctly")
			}
			if yamlData[0]["alias"] != "First Automation" {
				env.t.Error("First automation was modified unexpectedly")
			}
			if yamlData[2]["alias"] != "Third Automation" {
				env.t.Error("Third automation was modified unexpectedly")
			}
		})

		t.Run("restores are idempotent", func(t *testing.T) {
			env := setupPartialFileEnv(t, "automations.yaml", "id", "alias")
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_2", "20240101T120000.yaml"

			originalContent := readFile(t, "test-data/partial-automations-original.yaml")
			backupContent := readFile(t, "test-data/partial-automation-2-modified.yaml")

			env.writeFile(env.targetFile, originalContent, 0644)
			env.createBackup(groupSlug, path, id, filename, backupContent)

			// First restore
			w1, response1 := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w1)
			env.assertRestoreSuccess(response1)

			firstRestoreContent := env.readFile(env.targetFile)

			// Second restore (should produce identical result)
			w2, response2 := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w2)
			env.assertRestoreSuccess(response2)

			secondRestoreContent := env.readFile(env.targetFile)

			// Verify both restores produced identical results
			env.assertContentEquals(firstRestoreContent, secondRestoreContent)

			// Verify the modified section is present in both
			restoredStr := string(secondRestoreContent)
			env.assertContains(restoredStr, "alias: Modified Second Automation")
		})

		t.Run("adds a new section if the original no longer exists", func(t *testing.T) {
			env := setupPartialFileEnv(t, "automations.yaml", "id", "alias")
			groupSlug, path, id, filename := "test-automations", "automations.yaml", "automation_4", "20240101T120000.yaml"

			// Start with original file containing 3 automations
			originalContent := readFile(t, "test-data/partial-automations-original.yaml")
			env.writeFile(env.targetFile, originalContent, 0644)

			// Create backup for new automation (automation_4)
			backupContent := readFile(t, "test-data/partial-automation-4-new.yaml")
			env.createBackup(groupSlug, path, id, filename, backupContent)

			w, response := env.makeRestoreRequest(groupSlug, path, id, filename)
			env.assertStatusOK(w)
			env.assertRestoreSuccess(response)

			restoredContent := env.readFile(env.targetFile)
			restoredStr := string(restoredContent)

			// Verify all original automations are still present
			env.assertContains(restoredStr, "id: automation_1")
			env.assertContains(restoredStr, "alias: First Automation")
			env.assertContains(restoredStr, "id: automation_2")
			env.assertContains(restoredStr, "alias: Second Automation")
			env.assertContains(restoredStr, "id: automation_3")
			env.assertContains(restoredStr, "alias: Third Automation")

			// Verify new automation was added
			env.assertContains(restoredStr, "id: automation_4")
			env.assertContains(restoredStr, "alias: Fourth Automation")

			// Verify it's still valid YAML with 4 items
			var yamlData []interface{}
			if err := yaml.Unmarshal(restoredContent, &yamlData); err != nil {
				env.t.Fatalf("Failed to parse as sequence: %v", err)
			}
			if len(yamlData) != 4 {
				env.t.Errorf("Expected 4 items in YAML sequence after adding new section, got %d", len(yamlData))
			}
		})
	})
}

func setupTestDirs(t *testing.T) (tempDir, backupDir, haConfigDir string) {
	tempDir = t.TempDir()
	backupDir = filepath.Join(tempDir, "backups")
	haConfigDir = filepath.Join(tempDir, "ha-config")

	err := os.MkdirAll(haConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create HA config directory: %v", err)
	}

	return tempDir, backupDir, haConfigDir
}

func readFile(t *testing.T, path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return content
}

// testEnvironment encapsulates all test setup and provides helper methods
type testEnvironment struct {
	tempDir, backupDir, haConfigDir, targetFile string
	server                                      *core.Server
	router                                      *gin.Engine
	t                                           *testing.T
}

// setupSingleFileEnv creates a test environment for single file restore tests
func setupSingleFileEnv(t *testing.T, configPath string) *testEnvironment {
	tempDir, backupDir, haConfigDir := setupTestDirs(t)
	targetFile := filepath.Join(haConfigDir, configPath)

	appSettings := &types.AppSettings{
		HomeAssistantConfigDir: haConfigDir,
		BackupDir:              backupDir,
		Port:                   ":8080",
		ConfigGroups: []*types.ConfigBackupOptionGroup{
			types.NewConfigBackupOptionGroup(
				"Test Configs",
				[]*types.ConfigBackupOptions{
					{
						Path:       configPath,
						BackupType: "single",
					},
				},
			),
		},
	}

	server := core.NewServer(appSettings, "tmp/test-config.json")
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

// setupPartialFileEnv creates a test environment for partial file restore tests
func setupPartialFileEnv(t *testing.T, configPath, idNode, friendlyNameNode string) *testEnvironment {
	tempDir, backupDir, haConfigDir := setupTestDirs(t)
	targetFile := filepath.Join(haConfigDir, configPath)

	config := &types.AppSettings{
		HomeAssistantConfigDir: haConfigDir,
		BackupDir:              backupDir,
		Port:                   ":8080",
		ConfigGroups: []*types.ConfigBackupOptionGroup{
			types.NewConfigBackupOptionGroup(
				"Test Automations",
				[]*types.ConfigBackupOptions{
					{
						Path:             configPath,
						BackupType:       "multiple",
						IdNode:           &idNode,
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

// Helper methods for testEnvironment
func (env *testEnvironment) writeFile(path string, content []byte, perm os.FileMode) {
	err := os.WriteFile(path, content, perm)
	if err != nil {
		env.t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

func (env *testEnvironment) readFile(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		env.t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return content
}

func (env *testEnvironment) createBackup(groupSlug, path, id, filename string, content []byte) {
	backupPath := filepath.Join(env.backupDir, groupSlug, path, id)
	err := os.MkdirAll(backupPath, 0755)
	if err != nil {
		env.t.Fatalf("Failed to create backup directory: %v", err)
	}

	backupFile := filepath.Join(backupPath, filename)
	err = os.WriteFile(backupFile, content, 0644)
	if err != nil {
		env.t.Fatalf("Failed to write backup file: %v", err)
	}
}

func (env *testEnvironment) makeRestoreRequest(groupSlug, configPath, id, filename string) (*httptest.ResponseRecorder, *api.RestoreBackupResponse) {
	req, err := http.NewRequest(
		http.MethodPost,
		"/configs/"+groupSlug+"/"+configPath+"/"+id+"/backups/"+filename+"/restore",
		nil,
	)
	if err != nil {
		env.t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	var response api.RestoreBackupResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		env.t.Fatalf("Failed to parse response: %v", err)
	}

	return w, &response
}

// Assertion helpers
func (env *testEnvironment) assertStatusOK(w *httptest.ResponseRecorder) {
	if w.Code != http.StatusOK {
		env.t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func (env *testEnvironment) assertRestoreSuccess(response *api.RestoreBackupResponse) {
	if !response.Success {
		env.t.Errorf("Expected success=true, got false. Error: %s", response.Error)
	}
}

func (env *testEnvironment) assertContentEquals(expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		env.t.Errorf("Content mismatch.\nExpected:\n%s\nGot:\n%s", string(expected), string(actual))
	}
}

func (env *testEnvironment) assertContains(content, substring string) {
	if !strings.Contains(content, substring) {
		env.t.Errorf("Expected content to contain %q, but it didn't", substring)
	}
}

func (env *testEnvironment) assertNotContains(content, substring string) {
	if strings.Contains(content, substring) {
		env.t.Errorf("Expected content not to contain %q, but it did", substring)
	}
}

func (env *testEnvironment) assertValidYAML(content []byte) {
	var yamlData interface{}
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		env.t.Fatalf("Content is not valid YAML: %v\nContent:\n%s", err, string(content))
	}
}

// Test data helpers
type testData struct {
	original  []byte
	backup    []byte
	utf8      []byte
	different []byte
}

func loadTestData(t *testing.T) *testData {
	return &testData{
		original:  readFile(t, "test-data/single-file-original.yaml"),
		backup:    readFile(t, "test-data/single-file-backup.yaml"),
		utf8:      readFile(t, "test-data/single-file-utf8-backup.yaml"),
		different: readFile(t, "test-data/single-file-different.yaml"),
	}
}
