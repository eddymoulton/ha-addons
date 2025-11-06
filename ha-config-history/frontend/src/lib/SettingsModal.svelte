<script lang="ts">
  import Modal from "./Modal.svelte";
  import type {
    AppSettings,
    ConfigBackupOptions,
    UpdateSettingsResponse,
  } from "./types";
  import { api } from "./api";
  import { onMount } from "svelte";

  export let isOpen = false;
  export let onClose: () => void;

  let settings: AppSettings | null = null;
  let originalSettings: AppSettings | null = null;
  let loading = true;
  let saving = false;
  let backingUp = false;
  let backupSuccess = false;
  let error: string | null = null;
  let warnings: string[] = [];
  let editingConfigIndex: number | null = null;

  function hasChanged(field: string, value: any): boolean {
    if (!originalSettings || !settings) return false;
    return (
      JSON.stringify((originalSettings as any)[field]) !== JSON.stringify(value)
    );
  }

  $: if (isOpen && !settings) {
    loadSettings();
  }

  async function loadSettings() {
    loading = true;
    error = null;
    try {
      settings = await api.getSettings();
      originalSettings = JSON.parse(JSON.stringify(settings));
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load settings";
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

    // Validate configs
    settings.configs.forEach((config, index) => {
      if (!config.name.trim()) {
        errors.push(`Config #${index + 1}: Name is required`);
      }
      if (!config.path.trim()) {
        errors.push(`Config #${index + 1}: Path is required`);
      }
      if (
        config.maxBackups !== null &&
        config.maxBackups !== undefined &&
        config.maxBackups < 1
      ) {
        errors.push(
          `Config "${config.name}": Max Backups must be at least 1 or empty`
        );
      }
      if (
        config.maxBackupAgeDays !== null &&
        config.maxBackupAgeDays !== undefined &&
        config.maxBackupAgeDays < 1
      ) {
        errors.push(
          `Config "${config.name}": Max Age Days must be at least 1 or empty`
        );
      }
    });

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
      error = err instanceof Error ? err.message : "Failed to save settings";
    } finally {
      saving = false;
    }
  }

  function addConfig() {
    if (!settings) return;
    const newConfig: ConfigBackupOptions = {
      name: "",
      path: "",
      backupType: "multiple",
      idNode: "id",
      friendlyNameNode: "alias",
    };
    settings.configs = [...settings.configs, newConfig];
    editingConfigIndex = settings.configs.length - 1;
  }

  function removeConfig(index: number) {
    if (!settings) return;
    const config = settings.configs[index];
    if (
      !confirm(
        `Delete config "${config.name || "(Unnamed)"}"? This cannot be undone.`
      )
    ) {
      return;
    }
    settings.configs = settings.configs.filter((_, i) => i !== index);
    if (editingConfigIndex === index) {
      editingConfigIndex = null;
    }
  }

  function duplicateConfig(index: number) {
    if (!settings) return;
    const configToDuplicate = settings.configs[index];
    const duplicated: ConfigBackupOptions = {
      ...configToDuplicate,
      name: `${configToDuplicate.name} (copy)`,
    };
    settings.configs = [...settings.configs, duplicated];
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
      error = err instanceof Error ? err.message : "Failed to trigger backup";
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
      {#if error}
        <div class="alert alert-error">{error}</div>
      {/if}

      {#if backupSuccess}
        <div class="alert alert-success">Backup completed successfully!</div>
      {/if}

      {#if warnings.length > 0}
        <div class="alert alert-warning">
          <strong>Warnings:</strong>
          <ul>
            {#each warnings as warning}
              <li>{warning}</li>
            {/each}
          </ul>
          <button class="btn-small" on:click={onClose}>Close Anyway</button>
        </div>
      {/if}

      <section class="settings-section">
        <h3>General Settings</h3>

        <div class="form-group">
          <label for="ha-config-dir">Home Assistant Config Directory</label>
          <input
            id="ha-config-dir"
            type="text"
            bind:value={settings.homeAssistantConfigDir}
            placeholder="/config"
            class:changed={hasChanged(
              "homeAssistantConfigDir",
              settings.homeAssistantConfigDir
            )}
          />
        </div>

        <div class="form-group">
          <label for="backup-dir">Backup Directory</label>
          <input
            id="backup-dir"
            type="text"
            bind:value={settings.backupDir}
            placeholder="./backups"
            class:changed={hasChanged("backupDir", settings.backupDir)}
          />
        </div>

        <div class="form-group">
          <label for="port">Server Port</label>
          <input
            id="port"
            type="text"
            bind:value={settings.port}
            placeholder=":40613"
            class:changed={hasChanged("port", settings.port)}
          />
        </div>

        <div class="form-group">
          <label for="cron-schedule">
            Cron Schedule
            <span class="help-text"
              >(Leave empty to disable, e.g., "0 2 * * *" for daily at 2 AM)</span
            >
          </label>
          <input
            id="cron-schedule"
            type="text"
            bind:value={settings.cronSchedule}
            placeholder="0 2 * * *"
            class:changed={hasChanged("cronSchedule", settings.cronSchedule)}
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="max-backups">
              Default Max Backups
              <span class="help-text">(Leave empty for unlimited)</span>
            </label>
            <input
              id="max-backups"
              type="number"
              bind:value={settings.defaultMaxBackups}
              placeholder="unlimited"
              min="1"
            />
          </div>

          <div class="form-group">
            <label for="max-age">
              Default Max Age (Days)
              <span class="help-text">(Leave empty for unlimited)</span>
            </label>
            <input
              id="max-age"
              type="number"
              bind:value={settings.defaultMaxBackupAgeDays}
              placeholder="unlimited"
              min="1"
            />
          </div>
        </div>

        <div class="backup-action">
          <button
            class="btn-backup"
            type="button"
            on:click={handleBackupNow}
            disabled={backingUp}
          >
            {backingUp ? "Running Backup..." : "⚡ Backup Now"}
          </button>
          <span class="backup-help"
            >Manually trigger a backup of all configured files</span
          >
        </div>
      </section>

      <section class="settings-section">
        <div class="section-header">
          <h3>Config Backup Options</h3>
          <button class="btn-add" type="button" on:click={addConfig}
            >+ Add Config</button
          >
        </div>

        {#if settings.configs.length === 0}
          <div class="empty-state">
            No config backup options defined. Click "Add Config" to create one.
          </div>
        {:else}
          <div class="config-list">
            {#each settings.configs as config, index (index)}
              <div class="config-item">
                <div class="config-header">
                  <div class="config-title">
                    <strong>{config.name || "(Unnamed)"}</strong>
                    <span class="config-type">{config.backupType}</span>
                  </div>
                  <div class="config-actions">
                    <button
                      class="btn-icon"
                      type="button"
                      on:click={() =>
                        (editingConfigIndex =
                          editingConfigIndex === index ? null : index)}
                      title="Edit"
                    >
                      {editingConfigIndex === index ? "▼" : "▶"}
                    </button>
                    <button
                      class="btn-icon"
                      type="button"
                      on:click={() => duplicateConfig(index)}
                      title="Duplicate"
                    >
                      ⧉
                    </button>
                    <button
                      class="btn-icon btn-danger"
                      type="button"
                      on:click={() => removeConfig(index)}
                      title="Remove"
                    >
                      ×
                    </button>
                  </div>
                </div>

                {#if editingConfigIndex === index}
                  <div class="config-details">
                    <div class="form-group">
                      <label for="config-name-{index}">Name</label>
                      <input
                        id="config-name-{index}"
                        type="text"
                        bind:value={config.name}
                        placeholder="Config name"
                      />
                    </div>

                    <div class="form-group">
                      <label for="config-path-{index}">Path</label>
                      <input
                        id="config-path-{index}"
                        type="text"
                        bind:value={config.path}
                        placeholder="relative/path/to/file.yaml"
                      />
                    </div>

                    <div class="form-group">
                      <label for="config-type-{index}">Backup Type</label>
                      <select
                        id="config-type-{index}"
                        bind:value={config.backupType}
                      >
                        <option value="multiple">Multiple</option>
                        <option value="single">Single</option>
                        <option value="directory">Directory</option>
                      </select>
                    </div>

                    {#if config.backupType === "multiple"}
                      <div class="form-row">
                        <div class="form-group">
                          <label for="config-id-node-{index}">ID Node</label>
                          <input
                            id="config-id-node-{index}"
                            type="text"
                            bind:value={config.idNode}
                            placeholder="id"
                          />
                        </div>

                        <div class="form-group">
                          <label for="config-friendly-node-{index}"
                            >Friendly Name Node</label
                          >
                          <input
                            id="config-friendly-node-{index}"
                            type="text"
                            bind:value={config.friendlyNameNode}
                            placeholder="alias"
                          />
                        </div>
                      </div>
                    {/if}

                    <div class="form-row">
                      <div class="form-group">
                        <label for="config-max-backups-{index}"
                          >Max Backups</label
                        >
                        <input
                          id="config-max-backups-{index}"
                          type="number"
                          bind:value={config.maxBackups}
                          placeholder="Default"
                          min="1"
                        />
                      </div>

                      <div class="form-group">
                        <label for="config-max-age-{index}"
                          >Max Age (Days)</label
                        >
                        <input
                          id="config-max-age-{index}"
                          type="number"
                          bind:value={config.maxBackupAgeDays}
                          placeholder="Default"
                          min="1"
                        />
                      </div>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </section>

      <div class="modal-actions">
        <button
          class="btn btn-secondary"
          type="button"
          on:click={onClose}
          disabled={saving}
        >
          Cancel
        </button>
        <button
          class="btn btn-primary"
          type="button"
          on:click={handleSave}
          disabled={saving}
        >
          {saving ? "Saving..." : "Save Settings"}
        </button>
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
    color: var(--secondary-text-color, #9b9b9b);
  }

  .alert {
    padding: 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .alert-error {
    background: rgba(244, 67, 54, 0.1);
    border: 1px solid var(--error-color, #f44336);
    color: var(--error-color, #f44336);
  }

  .alert-warning {
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid var(--warning-color, #ff9800);
    color: var(--warning-color, #ff9800);
  }

  .alert-success {
    background: rgba(76, 175, 80, 0.1);
    border: 1px solid var(--success-color, #4caf50);
    color: var(--success-color, #4caf50);
  }

  .alert ul {
    margin: 0.5rem 0;
    padding-left: 1.5rem;
  }

  .alert li {
    margin: 0.25rem 0;
  }

  .settings-section {
    background: var(--ha-card-background, #1c1c1e);
    border: 1px solid var(--ha-card-border-color, #2c2c2e);
    border-radius: 8px;
    padding: 1.5rem;
  }

  .settings-section h3 {
    margin: 0 0 1rem 0;
    color: var(--primary-text-color, #ffffff);
    font-size: 1.1rem;
    font-weight: 500;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header h3 {
    margin: 0;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.9rem;
    font-weight: 500;
  }

  .help-text {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.8rem;
    font-weight: 400;
  }

  .form-group input,
  .form-group select {
    width: 100%;
    padding: 0.6rem;
    background: var(--ha-card-background, #2c2c2e);
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    border-radius: 4px;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.9rem;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: var(--primary-color, #03a9f4);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .btn-add {
    background: var(--primary-color, #03a9f4);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    transition: background-color 0.2s;
  }

  .btn-add:hover {
    background: var(--primary-color-dark, #0288d1);
  }

  .backup-action {
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--ha-card-border-color, #3c3c3e);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .btn-backup {
    background: var(--success-color, #4caf50);
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 600;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    align-self: flex-start;
  }

  .btn-backup:hover:not(:disabled) {
    background: var(--success-color-dark, #45a049);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(76, 175, 80, 0.3);
  }

  .btn-backup:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  .backup-help {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.85rem;
    font-style: italic;
  }

  .btn-small {
    background: var(--primary-color, #03a9f4);
    color: white;
    border: none;
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.85rem;
    margin-top: 0.5rem;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    color: var(--secondary-text-color, #9b9b9b);
    font-style: italic;
  }

  .config-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .config-item {
    background: var(--ha-card-background, #2c2c2e);
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    border-radius: 6px;
    padding: 1rem;
  }

  .config-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .config-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: var(--primary-text-color, #ffffff);
  }

  .config-type {
    background: var(--ha-card-border-color, #3c3c3e);
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    font-size: 0.75rem;
    color: var(--secondary-text-color, #9b9b9b);
    text-transform: uppercase;
  }

  .config-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-icon {
    background: transparent;
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    color: var(--primary-text-color, #ffffff);
    width: 2rem;
    height: 2rem;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.2rem;
    transition: all 0.2s;
  }

  .btn-icon:hover {
    background: var(--ha-card-border-color, #3c3c3e);
  }

  .btn-icon.btn-danger {
    color: var(--error-color, #f44336);
    border-color: var(--error-color, #f44336);
  }

  .btn-icon.btn-danger:hover {
    background: rgba(244, 67, 54, 0.1);
  }

  .config-details {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color, #3c3c3e);
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color, #2c2c2e);
  }

  .btn {
    padding: 0.6rem 1.5rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    border: none;
    transition: background-color 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: var(--ha-card-border-color, #3c3c3e);
    color: var(--primary-text-color, #ffffff);
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--ha-card-border-color, #4c4c4e);
  }

  .btn-primary {
    background: var(--success-color, #4caf50);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--success-color-dark, #45a049);
  }

  @media (max-width: 768px) {
    .form-row {
      grid-template-columns: 1fr;
    }

    .config-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
    }
  }
</style>
