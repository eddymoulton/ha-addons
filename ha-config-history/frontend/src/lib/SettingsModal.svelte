<script lang="ts">
  import Modal from "./Modal.svelte";
  import type {
    AppSettings,
    ConfigBackupOptions,
    UpdateSettingsResponse,
  } from "./types";
  import { api } from "./api";
  import { getErrorMessage } from "./utils";
  import IconButton from "./components/IconButton.svelte";
  import Button from "./components/Button.svelte";
  import FormGroup from "./components/FormGroup.svelte";
  import FormInput from "./components/FormInput.svelte";
  import Alert from "./components/Alert.svelte";
  import FormSelect from "./components/FormSelect.svelte";

  type Props = {
    isOpen: boolean;
    onClose: () => void;
  };

  let { isOpen, onClose }: Props = $props();

  let settings: AppSettings | null = $state(null);
  let originalSettings: AppSettings | null = null;
  let loading = $state(false);
  let saving = $state(false);
  let backingUp = $state(false);
  let backupSuccess = $state(false);
  let error: string | null = $state(null);
  let warnings: string[] = $state([]);
  let editingConfigIndex: number | null = $state(null);
  let openSection: Sections = $state("general");

  type Sections = "general" | "configs" | null;

  function toggleSection(section: Sections) {
    openSection = openSection === section ? null : section;
  }

  function hasChanged(field: string, value: any): boolean {
    if (!originalSettings || !settings) return false;
    return (
      JSON.stringify((originalSettings as any)[field]) !== JSON.stringify(value)
    );
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

    // Validate configs
    const uniquePaths = new Set<string>();
    settings.configs.forEach((config, index) => {
      if (!config.name.trim()) {
        errors.push(`Config #${index + 1}: Name is required`);
      }
      if (!config.path.trim()) {
        errors.push(`Config #${index + 1}: Path is required`);
      }
      if (uniquePaths.has(config.path)) {
        errors.push(`Config "${config.name}": Path must be unique`);
      } else {
        uniquePaths.add(config.path);
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
      error = getErrorMessage(err, "Failed to save settings");
    } finally {
      saving = false;
    }
  }

  function addConfig() {
    if (!settings) return;
    openSection = "configs";

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
      error = getErrorMessage(err, "Failed to trigger backup");
    } finally {
      backingUp = false;
    }
  }
</script>

{#snippet generalSettings()}
  {#if settings}
    <section class="settings-section">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="section-heading"
        onclick={() => toggleSection("general")}
        role="button"
        tabindex="0"
      >
        <span class="section-toggle"
          >{openSection === "general" ? "▼" : "▶"}</span
        >
        General Settings
      </div>

      {#if openSection === "general"}
        <div class="section-content">
          <FormGroup
            label="Home Assistant Config Directory"
            for="ha-config-dir"
          >
            <FormInput
              id="ha-config-dir"
              type="text"
              bind:value={settings.homeAssistantConfigDir}
              placeholder="/config"
              changed={hasChanged(
                "homeAssistantConfigDir",
                settings.homeAssistantConfigDir
              )}
            />
          </FormGroup>

          <FormGroup label="Backup Directory" for="backup-dir">
            <FormInput
              id="backup-dir"
              type="text"
              bind:value={settings.backupDir}
              placeholder="./backups"
              changed={hasChanged("backupDir", settings.backupDir)}
            />
          </FormGroup>

          <FormGroup label="Server Port" for="port">
            <FormInput
              id="port"
              type="text"
              bind:value={settings.port}
              placeholder=":40613"
              changed={hasChanged("port", settings.port)}
            />
          </FormGroup>

          <FormGroup
            label="Cron Schedule"
            for="cron-schedule"
            helpText="(Leave empty to disable, e.g., &quot;0 2 * * *&quot; for daily at 2 AM)"
          >
            <FormInput
              id="cron-schedule"
              type="text"
              bind:value={settings.cronSchedule}
              placeholder="0 2 * * *"
              changed={hasChanged("cronSchedule", settings.cronSchedule)}
            />
          </FormGroup>

          <div class="form-row">
            <FormGroup
              label="Default Max Backups"
              for="max-backups"
              helpText="(Leave empty for unlimited)"
            >
              <FormInput
                id="max-backups"
                type="number"
                bind:value={settings.defaultMaxBackups}
                placeholder="unlimited"
                min="1"
              />
            </FormGroup>

            <FormGroup
              label="Default Max Age (Days)"
              for="max-age"
              helpText="(Leave empty for unlimited)"
            >
              <FormInput
                id="max-age"
                type="number"
                bind:value={settings.defaultMaxBackupAgeDays}
                placeholder="unlimited"
                min="1"
              />
            </FormGroup>
          </div>

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
            <span class="backup-help"
              >Manually trigger a backup of all configured files</span
            >
          </div>
        </div>
      {/if}
    </section>
  {/if}
{/snippet}

{#snippet configItem(config: ConfigBackupOptions, index: number)}
  {#if settings}
    <div class="config-item">
      <div class="config-header">
        <div class="config-title">
          <strong>{config.name || "(Unnamed)"}</strong>
          <span class="config-type">{config.backupType}</span>
        </div>
        <div class="config-actions">
          <IconButton
            icon={editingConfigIndex === index ? "▼" : "▶"}
            variant="outlined"
            size="medium"
            type="button"
            onclick={() =>
              (editingConfigIndex =
                editingConfigIndex === index ? null : index)}
            title="Edit"
            aria-label="Edit"
          />
          <IconButton
            icon="⧉"
            variant="outlined"
            size="medium"
            type="button"
            onclick={() => duplicateConfig(index)}
            title="Duplicate"
            aria-label="Duplicate"
          />
          <IconButton
            icon="×"
            variant="danger"
            size="medium"
            type="button"
            onclick={() => removeConfig(index)}
            title="Remove"
            aria-label="Remove"
          />
        </div>
      </div>

      {#if editingConfigIndex === index}
        <div class="config-details">
          <FormGroup label="Name" for="config-name-{index}">
            <FormInput
              id="config-name-{index}"
              type="text"
              bind:value={config.name}
              placeholder="Config name"
            />
          </FormGroup>

          <FormGroup label="Path" for="config-path-{index}">
            <FormInput
              id="config-path-{index}"
              type="text"
              bind:value={config.path}
              placeholder="relative/path/to/file_or_directory"
            />
          </FormGroup>

          <FormGroup label="Backup Type" for="config-type-{index}">
            <FormSelect id="config-type-{index}" bind:value={config.backupType}>
              <option value="multiple">
                Single YAML file with a list of items
              </option>
              <option value="single">Single file</option>
              <option value="directory">Directory of files</option>
            </FormSelect>
          </FormGroup>

          {#if config.backupType === "multiple"}
            <div class="form-row">
              <FormGroup label="ID Node" for="config-id-node-{index}">
                <FormInput
                  id="config-id-node-{index}"
                  type="text"
                  bind:value={config.idNode}
                  placeholder="id"
                />
              </FormGroup>

              <FormGroup
                label="Friendly Name Node"
                for="config-friendly-node-{index}"
              >
                <FormInput
                  id="config-friendly-node-{index}"
                  type="text"
                  bind:value={config.friendlyNameNode}
                  placeholder="alias"
                />
              </FormGroup>
            </div>
          {/if}

          {#if config.backupType === "directory"}
            <FormGroup
              label="Include File Patterns"
              for="config-include-patterns-{index}"
              helpText="(Comma-separated glob patterns, e.g., *.yaml, *.json)"
            >
              <FormInput
                id="config-include-patterns-{index}"
                type="text"
                value={config.includeFilePatterns?.join(", ") || ""}
                oninput={(e) => {
                  const value = e.currentTarget.value.trim();
                  config.includeFilePatterns = value
                    ? value.split(",").map((p) => p.trim())
                    : [];
                }}
                placeholder="*.yaml, *.json"
              />
            </FormGroup>

            <FormGroup
              label="Exclude File Patterns"
              for="config-exclude-patterns-{index}"
              helpText="(Comma-separated glob patterns, e.g., *.backup)"
            >
              <FormInput
                id="config-exclude-patterns-{index}"
                type="text"
                value={config.excludeFilePatterns?.join(", ") || ""}
                oninput={(e) => {
                  const value = e.currentTarget.value.trim();
                  config.excludeFilePatterns = value
                    ? value.split(",").map((p) => p.trim())
                    : [];
                }}
                placeholder="*.backup, temp/*"
              />
            </FormGroup>
          {/if}

          <div class="form-row">
            <FormGroup label="Max Backups" for="config-max-backups-{index}">
              <FormInput
                id="config-max-backups-{index}"
                type="number"
                bind:value={config.maxBackups}
                placeholder="Default"
                min="1"
              />
            </FormGroup>

            <FormGroup label="Max Age (Days)" for="config-max-age-{index}">
              <FormInput
                id="config-max-age-{index}"
                type="number"
                bind:value={config.maxBackupAgeDays}
                placeholder="Default"
                min="1"
              />
            </FormGroup>
          </div>
        </div>
      {/if}
    </div>
  {/if}
{/snippet}

{#snippet configBackupOptions()}
  {#if settings}
    <section class="settings-section">
      <div class="section-header">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div
          class="section-heading"
          onclick={() => toggleSection("configs")}
          role="button"
          tabindex="0"
        >
          <span class="section-toggle">
            {openSection === "configs" ? "▼" : "▶"}
          </span>
          Config Backup Options
        </div>
        <Button
          label="Add Config"
          variant="primary"
          size="small"
          type="button"
          onclick={addConfig}
          icon="+"
        ></Button>
      </div>

      {#if openSection === "configs"}
        <div class="section-content">
          {#if settings.configs.length === 0}
            <div class="empty-state">
              No config backup options defined. Click "Add Config" to create
              one.
            </div>
          {:else}
            <div class="config-list">
              {#each settings.configs as config, index (index)}
                {@render configItem(config, index)}
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </section>
  {/if}
{/snippet}

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

      {@render generalSettings()}
      {@render configBackupOptions()}

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

  .settings-section .section-heading {
    margin: 0 0 1rem 0;
    color: var(--primary-text-color);
    font-size: 1.1rem;
    font-weight: 500;
  }

  .section-heading {
    cursor: pointer;
    user-select: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0 !important;
    transition: color 0.2s;
  }

  .section-heading:hover {
    color: var(--primary-color);
  }

  .section-toggle {
    font-size: 0.8rem;
    display: inline-flex;
    align-items: center;
    transition: transform 0.2s;
  }

  .section-content {
    margin-top: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header {
    margin: 0;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .backup-action {
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--ha-card-border-color);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .backup-help {
    color: var(--secondary-text-color);
    font-size: 0.85rem;
    font-style: italic;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    color: var(--secondary-text-color);
    font-style: italic;
  }

  .config-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .config-item {
    background: var(--ha-card-background);
    border: 1px solid var(--ha-card-border-color);
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
    color: var(--primary-text-color);
  }

  .config-type {
    background: var(--ha-card-border-color);
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    font-size: 0.75rem;
    color: var(--secondary-text-color);
    text-transform: uppercase;
  }

  .config-actions {
    display: flex;
    gap: 0.5rem;
  }

  .config-details {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color);
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color);
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
