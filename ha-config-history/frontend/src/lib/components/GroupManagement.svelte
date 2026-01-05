<script lang="ts">
  import type {
    AppSettings,
    ConfigBackupOptions,
    ConfigBackupOptionGroup,
    BackupType,
  } from "../types";
  import IconButton from "./IconButton.svelte";
  import Button from "./Button.svelte";
  import FormGroup from "./FormGroup.svelte";
  import FormInput from "./FormInput.svelte";
  import FormSelect from "./FormSelect.svelte";
  import Alert from "./Alert.svelte";
  import type { Sections } from "$lib/SettingsModal.svelte";

  type Props = {
    settings: AppSettings;
    openSection: Sections | null;
    onToggleSection: (section: Sections) => void;
    editingGroupIndex: number | null;
    newGroupName: string;
    onEditingGroupIndexChange: (index: number | null) => void;
    onNewGroupNameChange: (name: string) => void;
  };

  let {
    settings,
    openSection,
    onToggleSection,
    editingGroupIndex,
    newGroupName,
    onEditingGroupIndexChange,
    onNewGroupNameChange,
  }: Props = $props();

  let groupError: string | null = $state(null);
  let configError: string | null = $state(null);

  function getConfigGroups(): ConfigBackupOptionGroup[] {
    return settings.configGroups || [];
  }

  function getGroupNames(): string[] {
    return getConfigGroups().map((group) => group.groupName);
  }

  function validateGroupName(name: string): string | null {
    const trimmed = name.trim();
    if (!trimmed) {
      return "Group name cannot be empty";
    }
    if (trimmed.length > 100) {
      return "Group name cannot exceed 100 characters";
    }
    if (/[<>:"/\\|?*]/.test(trimmed)) {
      return "Group name contains invalid characters";
    }
    const reserved = ["null", "undefined", "admin", "root", "system"];
    if (reserved.some((r) => r.toLowerCase() === trimmed.toLowerCase())) {
      return `Group name '${trimmed}' is reserved`;
    }
    if (
      getGroupNames().some(
        (existing) => existing.toLowerCase() === trimmed.toLowerCase()
      )
    ) {
      return `Group name '${trimmed}' already exists`;
    }
    return null;
  }

  function validateConfigPath(path: string): string | null {
    const trimmed = path.trim();
    if (!trimmed) {
      return "Config path cannot be empty";
    }
    if (trimmed.length > 500) {
      return "Config path cannot exceed 500 characters";
    }
    if (trimmed.includes("..")) {
      return "Config path cannot contain '..' sequences";
    }
    if (trimmed.startsWith("/") && !trimmed.startsWith("/homeassistant")) {
      return "Absolute paths must be within homeassistant directory";
    }
    return null;
  }

  function validateConfigInGroup(
    config: ConfigBackupOptions,
    groupIndex: number
  ): string | null {
    const pathError = validateConfigPath(config.path || "");
    if (pathError) {
      return pathError;
    }

    if (!["single", "multiple", "directory"].includes(config.backupType)) {
      return "Invalid backup type";
    }

    if (config.backupType === "multiple") {
      if (!config.idNode?.trim()) {
        return "ID node is required for multiple backup type";
      }
      if (!config.friendlyNameNode?.trim()) {
        return "Friendly name node is required for multiple backup type";
      }
    }

    if (
      config.maxBackups !== null &&
      config.maxBackups !== undefined &&
      config.maxBackups < 1
    ) {
      return "Max backups must be at least 1";
    }

    if (
      config.maxBackupAgeDays !== null &&
      config.maxBackupAgeDays !== undefined &&
      config.maxBackupAgeDays < 1
    ) {
      return "Max backup age days must be at least 1";
    }

    // Check for duplicate paths across all groups
    for (let i = 0; i < (settings.configGroups?.length || 0); i++) {
      const group = settings.configGroups![i];
      for (let j = 0; j < group.configs.length; j++) {
        if (i === groupIndex) continue; // Skip current group
        if (group.configs[j].path === config.path) {
          return `Path '${config.path}' is already used in group '${group.groupName}'`;
        }
      }
    }

    return null;
  }

  function clearErrors() {
    groupError = null;
    configError = null;
  }

  function getFriendlyBackupTypeName(backupType: BackupType): string {
    switch (backupType) {
      case "multiple":
        return "List of YAML Configurations";
      case "directory":
        return "Directory of Files";
      case "single":
        return "Single File";
    }
  }

  function addGroup() {
    clearErrors();
    if (!newGroupName.trim()) {
      groupError = "Group name cannot be empty";
      return;
    }

    const validation = validateGroupName(newGroupName);
    if (validation) {
      groupError = validation;
      return;
    }

    const newGroup: ConfigBackupOptionGroup = {
      groupName: newGroupName.trim(),
      configs: [],
    };

    if (!settings.configGroups) {
      settings.configGroups = [];
    }

    settings.configGroups = [...settings.configGroups, newGroup];
    onNewGroupNameChange("");
    onToggleSection("groups");
    onEditingGroupIndexChange(settings.configGroups.length - 1);
  }

  function removeGroup(index: number) {
    clearErrors();
    if (index < 0 || index >= (settings.configGroups?.length || 0)) {
      groupError = "Invalid group index";
      return;
    }

    const group = settings.configGroups[index];
    if (
      !confirm(
        `Delete group "${group.groupName}"? This will also delete ${group.configs.length} config(s). This cannot be undone.`
      )
    ) {
      return;
    }

    try {
      settings.configGroups = settings.configGroups.filter(
        (_, i) => i !== index
      );
      if (editingGroupIndex === index) {
        onEditingGroupIndexChange(null);
      }
    } catch (err) {
      groupError = "Failed to remove group";
      console.error("Error removing group:", err);
    }
  }

  function addConfigToGroup(groupIndex: number) {
    clearErrors();
    if (groupIndex < 0 || groupIndex >= (settings.configGroups?.length || 0)) {
      groupError = "Invalid group index";
      return;
    }

    try {
      onToggleSection("groups");

      const newConfig: ConfigBackupOptions = {
        path: "",
        backupType: "multiple",
        idNode: "id",
        friendlyNameNode: "alias",
      };

      settings.configGroups[groupIndex].configs = [
        ...settings.configGroups[groupIndex].configs,
        newConfig,
      ];
      onEditingGroupIndexChange(groupIndex);
    } catch (err) {
      configError = "Failed to add config to group";
      console.error("Error adding config to group:", err);
    }
  }

  function removeConfigFromGroup(groupIndex: number, configIndex: number) {
    clearErrors();
    if (groupIndex < 0 || groupIndex >= (settings.configGroups?.length || 0)) {
      groupError = "Invalid group index";
      return;
    }

    const group = settings.configGroups[groupIndex];
    if (configIndex < 0 || configIndex >= group.configs.length) {
      configError = "Invalid config index";
      return;
    }

    const config = group.configs[configIndex];
    if (!confirm(`Remove config from group? This cannot be undone.`)) {
      return;
    }

    try {
      settings.configGroups[groupIndex].configs = group.configs.filter(
        (_, i) => i !== configIndex
      );
    } catch (err) {
      configError = "Failed to remove config from group";
      console.error("Error removing config from group:", err);
    }
  }

  function moveConfigToGroup(
    fromGroupIndex: number,
    configIndex: number,
    toGroupName: string
  ) {
    clearErrors();
    if (
      fromGroupIndex < 0 ||
      fromGroupIndex >= (settings.configGroups?.length || 0)
    ) {
      groupError = "Invalid source group index";
      return;
    }

    const fromGroup = settings.configGroups[fromGroupIndex];
    if (configIndex < 0 || configIndex >= fromGroup.configs.length) {
      configError = "Invalid config index";
      return;
    }

    const config = fromGroup.configs[configIndex];
    const toGroupIndex = settings.configGroups.findIndex(
      (g) => g.groupName === toGroupName
    );

    if (toGroupIndex === -1) {
      configError = `Target group '${toGroupName}' not found`;
      return;
    }

    const validation = validateConfigInGroup(config, toGroupIndex);
    if (validation) {
      configError = validation;
      return;
    }

    try {
      settings.configGroups[fromGroupIndex].configs = fromGroup.configs.filter(
        (_, i) => i !== configIndex
      );

      settings.configGroups[toGroupIndex].configs = [
        ...settings.configGroups[toGroupIndex].configs,
        config,
      ];
    } catch (err) {
      configError = "Failed to move config to group";
      console.error("Error moving config to group:", err);
    }
  }

  function duplicateConfigInGroup(groupIndex: number, configIndex: number) {
    clearErrors();
    if (groupIndex < 0 || groupIndex >= (settings.configGroups?.length || 0)) {
      groupError = "Invalid group index";
      return;
    }

    const group = settings.configGroups[groupIndex];
    if (configIndex < 0 || configIndex >= group.configs.length) {
      configError = "Invalid config index";
      return;
    }

    try {
      const configToDuplicate = group.configs[configIndex];
      const duplicated: ConfigBackupOptions = {
        ...configToDuplicate,
      };

      const validation = validateConfigInGroup(duplicated, groupIndex);
      if (validation) {
        configError = validation;
        return;
      }

      settings.configGroups[groupIndex].configs = [
        ...group.configs,
        duplicated,
      ];
    } catch (err) {
      configError = "Failed to duplicate config";
      console.error("Error duplicating config:", err);
    }
  }

  function handleConfigChange(
    config: ConfigBackupOptions,
    groupIndex: number,
    field: string
  ) {
    clearErrors();

    if (field === "path" || field === "name" || field === "backupType") {
      const validation = validateConfigInGroup(config, groupIndex);
      if (validation) {
        configError = validation;
      }
    }
  }
</script>

{#snippet groupSettings(group: ConfigBackupOptionGroup, groupIndex: number)}
  <div class="group-item">
    <div class="group-header">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="group-title"
        onclick={() =>
          onEditingGroupIndexChange(
            editingGroupIndex === groupIndex ? null : groupIndex
          )}
        role="button"
        tabindex="0"
      >
        <span class="group-toggle">
          {editingGroupIndex === groupIndex ? "▼" : "▶"}
        </span>
        <strong>{group.groupName}</strong>
        <span class="config-count">
          {group.configs.length} config{group.configs.length !== 1 ? "s" : ""}
        </span>
      </div>
      <div class="group-actions">
        <IconButton
          icon="+"
          variant="outlined"
          size="medium"
          type="button"
          onclick={() => addConfigToGroup(groupIndex)}
          title="Add Config to Group"
          aria-label="Add Config to Group"
        />
        <IconButton
          icon="×"
          variant="danger"
          size="medium"
          type="button"
          onclick={() => removeGroup(groupIndex)}
          title="Delete Group"
          aria-label="Delete Group"
        />
      </div>
    </div>

    {#if editingGroupIndex === groupIndex}
      <div class="group-details">
        <FormGroup label="Group Name" for="group-name-{groupIndex}">
          <FormInput
            id="group-name-{groupIndex}"
            type="text"
            bind:value={group.groupName}
            placeholder="Group name"
          />
        </FormGroup>

        <div class="group-configs">
          <h4>Configs in this group:</h4>
          {#if group.configs.length === 0}
            <div class="empty-configs">
              No configs in this group. Click the "+" button above to add one.
            </div>
          {:else}
            <div class="config-list">
              {#each group.configs as config, configIndex (configIndex)}
                {@render configSettings(config, configIndex, group, groupIndex)}
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/snippet}

{#snippet configSettings(
  config: ConfigBackupOptions,
  configIndex: number,
  group: ConfigBackupOptionGroup,
  groupIndex: number
)}
  <div class="config-item config-in-group">
    <div class="config-header">
      <div class="config-title"></div>
      <div class="config-actions">
        {#if settings.configGroups.length > 1}
          <FormSelect
            id="move-config-{groupIndex}-{configIndex}"
            value={group.groupName}
            onchange={(e) => {
              const target = e.target as HTMLSelectElement;
              if (
                target.value &&
                target.value !== settings.configGroups[groupIndex].groupName
              ) {
                moveConfigToGroup(groupIndex, configIndex, target.value);
              }
              target.value = settings.configGroups[groupIndex].groupName; // Reset
            }}
          >
            <option value={settings.configGroups[groupIndex].groupName}>
              Move to...
            </option>
            {#each getGroupNames() as groupName}
              {#if groupName !== settings.configGroups[groupIndex].groupName}
                <option value={groupName}>{groupName}</option>
              {/if}
            {/each}
          </FormSelect>
        {/if}
        <IconButton
          icon="⧉"
          variant="outlined"
          size="large"
          type="button"
          onclick={() => duplicateConfigInGroup(groupIndex, configIndex)}
          title="Duplicate"
          aria-label="Duplicate"
        />
        <IconButton
          icon="×"
          variant="danger"
          size="large"
          type="button"
          onclick={() => removeConfigFromGroup(groupIndex, configIndex)}
          title="Remove from group"
          aria-label="Remove from group"
        />
      </div>
    </div>

    <div class="config-details">
      <div class="config-inline-form">
        <FormGroup
          label="Path"
          for={groupIndex + "." + configIndex + ".path"}
          weight="light"
        >
          <FormInput
            id={groupIndex + "." + configIndex + ".path"}
            type="text"
            bind:value={config.path}
            oninput={() => handleConfigChange(config, groupIndex, "path")}
          />
        </FormGroup>
        <FormGroup
          label="Type"
          for={groupIndex + "." + configIndex + ".type"}
          weight="light"
        >
          <FormSelect
            id={groupIndex + "." + configIndex + ".type"}
            bind:value={config.backupType}
            onchange={() =>
              handleConfigChange(config, groupIndex, "backupType")}
          >
            <option value="multiple">
              {getFriendlyBackupTypeName("multiple")}
            </option>
            <option value="single">
              {getFriendlyBackupTypeName("single")}
            </option>
            <option value="directory">
              {getFriendlyBackupTypeName("directory")}
            </option>
          </FormSelect>
        </FormGroup>
      </div>

      {#if config.backupType === "multiple"}
        <div class="config-inline-form">
          <FormGroup
            label="ID Node"
            for={groupIndex + "." + configIndex + ".idNode"}
            weight="light"
          >
            <FormInput
              id={groupIndex + "." + configIndex + ".idNode"}
              type="text"
              bind:value={config.idNode}
              placeholder="id"
            />
          </FormGroup>
          <FormGroup
            label="Friendly Name Node"
            for={groupIndex + "." + configIndex + ".friendlyNameNode"}
            weight="light"
          >
            <FormInput
              id={groupIndex + "." + configIndex + ".friendlyNameNode"}
              type="text"
              bind:value={config.friendlyNameNode}
              placeholder="alias"
            />
          </FormGroup>
        </div>
      {/if}

      {#if config.backupType === "directory"}
        <div class="config-inline-form">
          <FormGroup
            label="Include patterns"
            for={groupIndex + "." + configIndex + ".include"}
            weight="light"
          >
            <FormInput
              id={groupIndex + "." + configIndex + ".include"}
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
            label="Exclude Patterns"
            for={groupIndex + "." + configIndex + ".exclude"}
            weight="light"
          >
            <FormInput
              id={groupIndex + "." + configIndex + ".exclude"}
              type="text"
              value={config.excludeFilePatterns?.join(", ") || ""}
              oninput={(e) => {
                const value = e.currentTarget.value.trim();
                config.excludeFilePatterns = value
                  ? value.split(",").map((p) => p.trim())
                  : [];
              }}
              placeholder="*.backup, secrets"
            />
          </FormGroup>
        </div>
      {/if}

      <div class="config-inline-form">
        <FormGroup
          label="Max backups"
          for={groupIndex + "." + configIndex + ".maxBackups"}
          weight="light"
        >
          <FormInput
            id={groupIndex + "." + configIndex + ".maxBackups"}
            type="number"
            bind:value={config.maxBackups}
            min="1"
          />
        </FormGroup>
        <FormGroup
          label="Max Age in Days"
          for={groupIndex + "." + configIndex + ".maxBackupAgeDays"}
          weight="light"
        >
          <FormInput
            id={groupIndex + "." + configIndex + ".maxBackupAgeDays"}
            type="number"
            bind:value={config.maxBackupAgeDays}
            min="1"
          />
        </FormGroup>
      </div>
    </div>
  </div>
{/snippet}

<section class="settings-section">
  <div class="section-header">
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="section-heading"
      onclick={() => onToggleSection("groups")}
      role="button"
      tabindex="0"
    >
      <span class="section-toggle">
        {openSection === "groups" ? "▼" : "▶"}
      </span>
      Group Management
    </div>
    <div class="add-group-controls">
      <FormInput
        type="text"
        placeholder="New group name..."
        value={newGroupName}
        oninput={(e) => {
          const value = e.currentTarget.value;
          onNewGroupNameChange(value);
          if (value.trim()) {
            const validation = validateGroupName(value);
            groupError = validation;
          } else {
            groupError = null;
          }
        }}
      />
      <Button
        label="Add Group"
        variant="primary"
        size="small"
        type="button"
        onclick={addGroup}
        disabled={!newGroupName.trim()}
        icon="+"
      ></Button>
    </div>
  </div>

  {#if openSection === "groups"}
    <div class="section-content">
      <Alert type="error" message={groupError} />
      <Alert type="error" message={configError} />

      {#if !settings.configGroups || settings.configGroups.length === 0}
        <div class="empty-state">
          No groups defined. Add a group above to get started.
        </div>
      {:else}
        <div class="group-list">
          {#each settings.configGroups as group, groupIndex (groupIndex)}
            {@render groupSettings(group, groupIndex)}
          {/each}
        </div>
      {/if}
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

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header {
    margin: 0;
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
    cursor: pointer;
    user-select: none;
    flex: 1;
    transition: color 0.2s;
  }

  .config-title:hover {
    color: var(--primary-color);
  }

  .config-actions {
    display: flex;
    gap: 0.5rem;
  }

  .config-details {
    margin-top: 1rem;
  }

  .add-group-controls {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .group-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .group-item {
    background: var(--ha-card-background);
    border: 2px solid var(--primary-color);
    border-radius: 8px;
    padding: 1rem;
  }

  .group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .group-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: var(--primary-text-color);
    cursor: pointer;
    user-select: none;
    flex: 1;
    transition: color 0.2s;
  }

  .group-title:hover {
    color: var(--primary-color);
  }

  .group-toggle {
    font-size: 0.8rem;
    display: inline-flex;
    align-items: center;
    transition: transform 0.2s;
  }

  .config-count {
    background: var(--primary-color);
    color: white;
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    font-size: 0.75rem;
    text-transform: uppercase;
  }

  .group-actions {
    display: flex;
    gap: 0.5rem;
  }

  .group-details {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ha-card-border-color);
  }

  .group-configs {
    margin-top: 1.5rem;
  }

  .group-configs h4 {
    margin: 0 0 1rem 0;
    color: var(--primary-text-color);
    font-size: 1rem;
    font-weight: 500;
  }

  .empty-configs {
    text-align: center;
    padding: 1.5rem;
    color: var(--secondary-text-color);
    font-style: italic;
    background: var(--ha-card-border-color);
    border-radius: 4px;
  }

  .config-in-group {
    border: 1px solid var(--ha-card-border-color);
    margin-bottom: 0.5rem;
    background: var(--ha-card-background);
  }

  .config-inline-form {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  @media (max-width: 768px) {
    .config-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
    }

    .add-group-controls {
      flex-direction: column;
      width: 100%;
    }

    .config-inline-form {
      grid-template-columns: 1fr;
    }

    .group-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
    }
  }
</style>
