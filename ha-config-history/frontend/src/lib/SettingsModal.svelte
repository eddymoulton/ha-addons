<script lang="ts">
  import Modal from "./Modal.svelte";
  import type { AppSettings, UpdateSettingsResponse } from "./types";
  import { api } from "./api";
  import { getErrorMessage } from "./utils";
  import Button from "./components/Button.svelte";
  import Alert from "./components/Alert.svelte";
  import GeneralSettings from "./components/GeneralSettings.svelte";
  import GroupManagement from "./components/GroupManagement.svelte";

  type Props = {
    isOpen: boolean;
    onClose: () => void;
  };

  let { isOpen, onClose }: Props = $props();

  let settings: AppSettings | null = $state(null);
  let originalSettings: AppSettings | null = $state(null);
  let loading = $state(false);
  let saving = $state(false);
  let backingUp = $state(false);
  let backupSuccess = $state(false);
  let error: string | null = $state(null);
  let warnings: string[] = $state([]);
  let openSection: Sections = $state("general");
  let editingGroupIndex: number | null = $state(null);
  let newGroupName = $state("");

  export type Sections = "general" | "groups" | "configs" | null;

  function toggleSection(section: Sections) {
    openSection = openSection === section ? null : section;
  }

  $effect(() => {
    loadSettings();
  });

  async function loadSettings() {
    loading = true;
    error = null;
    try {
      settings = await api.getSettings();
      originalSettings = JSON.parse(JSON.stringify(settings));
    } catch (err) {
      error = getErrorMessage(err, "Failed to load settings");
    } finally {
      loading = false;
    }
  }

  function validateSettings(settings: AppSettings): string[] {
    const errors: string[] = [];

    if (!settings.homeAssistantConfigDir.trim()) {
      errors.push("Home Assistant Config Directory is required");
    }

    if (!settings.backupDir.trim()) {
      errors.push("Backup Directory is required");
    }

    if (settings.port && !settings.port.match(/^:\d+$/)) {
      errors.push("Port must be in format ':port' (e.g., ':40613')");
    }

    if (
      settings.cronSchedule &&
      settings.cronSchedule.trim() &&
      !settings.cronSchedule.match(/^[\d\*\/\-,\s]+$/)
    ) {
      errors.push("Cron schedule format appears invalid");
    }

    if (
      settings.defaultMaxBackups !== null &&
      settings.defaultMaxBackups !== undefined &&
      settings.defaultMaxBackups < 1
    ) {
      errors.push("Default Max Backups must be at least 1 or empty");
    }

    if (
      settings.defaultMaxBackupAgeDays !== null &&
      settings.defaultMaxBackupAgeDays !== undefined &&
      settings.defaultMaxBackupAgeDays < 1
    ) {
      errors.push("Default Max Age Days must be at least 1 or empty");
    }

    // Validate groups and configs
    const uniquePaths = new Set<string>();
    const uniqueGroupNames = new Set<string>();

    if (!settings.configGroups || settings.configGroups.length === 0) {
      errors.push("At least one config group is required");
    } else {
      settings.configGroups.forEach((group, groupIndex) => {
        // Validate group names
        if (!group.groupName.trim()) {
          errors.push(`Group #${groupIndex + 1}: Name is required`);
        }
        if (uniqueGroupNames.has(group.groupName)) {
          errors.push(`Group "${group.groupName}": Name must be unique`);
        } else {
          uniqueGroupNames.add(group.groupName);
        }

        // Validate configs within group
        if (!group.configs || group.configs.length === 0) {
          errors.push(
            `Group "${group.groupName}": Must contain at least one config`
          );
        } else {
          group.configs.forEach((config, configIndex) => {
            const configId = `Group "${group.groupName}" Config #${configIndex + 1}`;

            if (!config.name.trim()) {
              errors.push(`${configId}: Name is required`);
            }
            if (!config.path.trim()) {
              errors.push(`${configId}: Path is required`);
            }
            if (uniquePaths.has(config.path)) {
              errors.push(
                `Config "${config.name}": Path must be unique across all groups`
              );
            } else {
              uniquePaths.add(config.path);
            }
            if (
              config.maxBackups !== null &&
              config.maxBackups !== undefined &&
              config.maxBackups < 1
            ) {
              errors.push(
                `${configId}: Max Backups must be at least 1 or empty`
              );
            }
            if (
              config.maxBackupAgeDays !== null &&
              config.maxBackupAgeDays !== undefined &&
              config.maxBackupAgeDays < 1
            ) {
              errors.push(
                `${configId}: Max Age Days must be at least 1 or empty`
              );
            }
          });
        }
      });
    }

    return errors;
  }

  async function handleSave() {
    if (!settings) return;

    // Validate settings first
    const validationErrors = validateSettings(settings);
    if (validationErrors.length > 0) {
      error = validationErrors.join("; ");
      return;
    }

    saving = true;
    error = null;
    warnings = [];

    try {
      const response: UpdateSettingsResponse =
        await api.updateSettings(settings);

      if (response.success) {
        if (response.warnings && response.warnings.length > 0) {
          warnings = response.warnings;
        } else {
          // Close modal if no warnings
          onClose();
        }
      } else if (response.error) {
        error = response.error;
      }
    } catch (err) {
      error = getErrorMessage(err, "Failed to save settings");
    } finally {
      saving = false;
    }
  }

  async function handleBackupNow() {
    backingUp = true;
    backupSuccess = false;
    error = null;

    try {
      await api.triggerBackup();
      backupSuccess = true;
      // Reset success message after 3 seconds
      setTimeout(() => {
        backupSuccess = false;
      }, 3000);
    } catch (err) {
      error = getErrorMessage(err, "Failed to trigger backup");
    } finally {
      backingUp = false;
    }
  }
</script>

<Modal {isOpen} title="Settings" {onClose}>
  {#if loading}
    <div class="loading">Loading settings...</div>
  {:else if settings}
    <div class="settings-form">
      <Alert type="error" message={error} />

      {#if backupSuccess}
        <Alert type="success" message="Backup completed successfully!" />
      {/if}

      {#if warnings.length > 0}
        <Alert type="warning">
          <strong>Warnings:</strong>
          <ul>
            {#each warnings as warning}
              <li>{warning}</li>
            {/each}
          </ul>
          <Button
            size="small"
            variant="primary"
            onclick={onClose}
            label="Close Anyway"
          />
        </Alert>
      {/if}

      <div class="backup-action">
        <Button
          label={backingUp ? "Running Backup..." : "Backup Now"}
          variant="success"
          size="large"
          type="button"
          onclick={handleBackupNow}
          disabled={backingUp}
          loading={backingUp}
          icon={backingUp ? undefined : "⚡"}
        />
        <span class="backup-help">
          Manually trigger a backup of all configured files
        </span>
      </div>

      {#if settings}
        <GeneralSettings
          {settings}
          {originalSettings}
          {openSection}
          onToggleSection={toggleSection}
        />
        <GroupManagement
          {settings}
          {openSection}
          onToggleSection={toggleSection}
          {editingGroupIndex}
          {newGroupName}
          onEditingGroupIndexChange={(index) => (editingGroupIndex = index)}
          onNewGroupNameChange={(name) => (newGroupName = name)}
        />
      {/if}

      <div class="modal-actions">
        <Button
          label="Cancel"
          variant="secondary"
          type="button"
          onclick={onClose}
          disabled={saving}
        />
        <Button
          label={saving ? "Saving..." : "Save Settings"}
          variant="success"
          type="button"
          onclick={handleSave}
          disabled={saving}
          loading={saving}
        />
      </div>
    </div>
  {/if}
</Modal>

<style>
  .settings-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    max-height: 70vh;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: var(--secondary-text-color);
  }

  .backup-action {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .backup-help {
    color: var(--secondary-text-color);
    font-size: 0.85rem;
    font-style: italic;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color);
  }
</style>
