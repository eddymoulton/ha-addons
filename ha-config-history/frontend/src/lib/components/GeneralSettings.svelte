<script lang="ts">
  import type { Sections } from "$lib/SettingsModal.svelte";
  import type { AppSettings } from "../types";
  import FormGroup from "./FormGroup.svelte";
  import FormInput from "./FormInput.svelte";
  import Alert from "./Alert.svelte";

  type Props = {
    settings: AppSettings;
    originalSettings: AppSettings | null;
    openSection: Sections | null;
    onToggleSection: (section: Sections) => void;
  };

  let { settings, originalSettings, openSection, onToggleSection }: Props =
    $props();

  let fieldErrors: Record<string, string | null> = $state({});

  function hasChanged(field: string, value: any): boolean {
    if (!originalSettings || !settings) return false;
    return (
      JSON.stringify((originalSettings as any)[field]) !== JSON.stringify(value)
    );
  }

  function validateField(field: string, value: any): string | null {
    switch (field) {
      case "homeAssistantConfigDir":
        if (!value || !value.trim()) {
          return "Home Assistant Config Directory is required";
        }
        return null;
      case "backupDir":
        if (!value || !value.trim()) {
          return "Backup Directory is required";
        }
        return null;
      case "port":
        if (value && !value.match(/^:\d+$/)) {
          return "Port must be in format ':port' (e.g., ':40613')";
        }
        return null;
      case "cronSchedule":
        if (value && value.trim() && !value.match(/^[\d\*\/\-,\s]+$/)) {
          return "Cron schedule format appears invalid";
        }
        return null;
      case "defaultMaxBackups":
        if (value !== null && value !== undefined && value < 1) {
          return "Default Max Backups must be at least 1";
        }
        return null;
      case "defaultMaxBackupAgeDays":
        if (value !== null && value !== undefined && value < 1) {
          return "Default Max Age Days must be at least 1";
        }
        return null;
      default:
        return null;
    }
  }

  function handleFieldChange(field: string, value: any) {
    const error = validateField(field, value);
    fieldErrors[field] = error;
  }
</script>

<section class="settings-section">
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="section-heading"
    onclick={() => onToggleSection("general")}
    role="button"
    tabindex="0"
  >
    <span class="section-toggle">{openSection === "general" ? "▼" : "▶"}</span>
    General Settings
  </div>

  {#if openSection === "general"}
    <div class="section-content">
      {#each Object.entries(fieldErrors) as [field, error]}
        {#if error}
          <Alert type="error" message={error} />
        {/if}
      {/each}
      
      <FormGroup label="Home Assistant Config Directory" for="ha-config-dir">
        <FormInput
          id="ha-config-dir"
          type="text"
          bind:value={settings.homeAssistantConfigDir}
          placeholder="/config"
          oninput={() => handleFieldChange("homeAssistantConfigDir", settings.homeAssistantConfigDir)}
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
          oninput={() => handleFieldChange("backupDir", settings.backupDir)}
          changed={hasChanged("backupDir", settings.backupDir)}
        />
      </FormGroup>

      <FormGroup label="Server Port" for="port">
        <FormInput
          id="port"
          type="text"
          bind:value={settings.port}
          placeholder=":40613"
          oninput={() => handleFieldChange("port", settings.port)}
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
          oninput={() => handleFieldChange("cronSchedule", settings.cronSchedule)}
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
            oninput={() => handleFieldChange("defaultMaxBackups", settings.defaultMaxBackups)}
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
            oninput={() => handleFieldChange("defaultMaxBackupAgeDays", settings.defaultMaxBackupAgeDays)}
            min="1"
          />
        </FormGroup>
      </div>
    </div>
  {/if}
</section>

<style>
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

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  @media (max-width: 768px) {
    .form-row {
      grid-template-columns: 1fr;
    }
  }
</style>
