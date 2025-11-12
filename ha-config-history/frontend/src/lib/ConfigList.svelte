<script lang="ts">
  import { onMount } from "svelte";
  import type { ConfigMetadata } from "./types";
  import { api } from "./api";
  import { formatFileSize } from "./utils";
  import LoadingState from "./LoadingState.svelte";

  export let onConfigClick: (config: ConfigMetadata) => void;
  export let selectedConfig: ConfigMetadata | null = null;

  let configs: ConfigMetadata[] = [];
  let loading = true;
  let error: string | null = null;
  let selectedGroup: string = "all";
  let searchQuery: string = "";
  let configToDelete: ConfigMetadata | null = null;
  let showDeleteConfirm = false;
  let deleting = false;

  $: groups = Array.from(new Set(configs.map((c) => c.group))).sort();
  $: filteredConfigs = configs
    .filter((c) => selectedGroup === "all" || c.group === selectedGroup)
    .filter(
      (c) =>
        searchQuery === "" ||
        c.friendlyName.toLowerCase().includes(searchQuery.toLowerCase()) ||
        c.id.toLowerCase().includes(searchQuery.toLowerCase())
    );

  async function loadConfigs() {
    loading = true;
    error = null;
    try {
      configs = await api.getConfigs();

      // Automatically select the first automation
      if (configs.length > 0 && !selectedConfig) {
        onConfigClick(configs[0]);
        selectedGroup = configs[0].group;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load configs";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadConfigs();
  });

  function handleGroupChange() {
    // Auto-select first config in the newly selected group
    if (filteredConfigs.length > 0) {
      onConfigClick(filteredConfigs[0]);
    }
  }

  function handleDeleteClick(config: ConfigMetadata, event: Event) {
    event.stopPropagation();
    configToDelete = config;
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    showDeleteConfirm = false;
    configToDelete = null;
  }

  async function confirmDelete() {
    if (!configToDelete) return;

    deleting = true;
    try {
      await api.deleteAllBackups(configToDelete.group, configToDelete.id);

      // If the deleted config was selected, clear selection
      if (selectedConfig?.id === configToDelete.id) {
        selectedConfig = null;
      }

      // Reload configs after successful deletion
      await loadConfigs();

      showDeleteConfirm = false;
      configToDelete = null;
    } catch (err) {
      error =
        err instanceof Error ? err.message : "Failed to delete config backups";
    } finally {
      deleting = false;
    }
  }
</script>

<div class="automation-list">
  <LoadingState
    {loading}
    {error}
    empty={!loading && !error && configs.length === 0}
    emptyMessage="No configs found"
    loadingMessage="Loading configs..."
  />

  {#if !loading && !error && configs.length > 0}
    <div class="filter-section">
      <div class="group-filter-row">
        <select
          id="group-filter"
          bind:value={selectedGroup}
          on:change={handleGroupChange}
          class="group-select"
        >
          {#each groups as group}
            <option value={group}>{group}</option>
          {/each}
        </select>
        <button
          class="refresh-btn"
          on:click={loadConfigs}
          type="button"
          title="Refresh configs"
          aria-label="Refresh configs"
        >
          ⟳
        </button>
      </div>
      <div class="search-box">
        <input
          type="text"
          placeholder="Search configs..."
          bind:value={searchQuery}
          class="search-input"
        />
        <div class="filter-count">
          {filteredConfigs.length} config{filteredConfigs.length !== 1
            ? "s"
            : ""}
        </div>
      </div>
    </div>

    {#if filteredConfigs.length === 0}
      <LoadingState empty={true} emptyMessage="No configs in this group" />
    {:else}
      <div class="grid">
        {#each filteredConfigs as config (config.id)}
          <div
            class="automation-card {selectedConfig?.id === config.id
              ? 'selected'
              : ''}"
            on:click={() => onConfigClick(config)}
            on:keydown={(e) => e.key === "Enter" && onConfigClick(config)}
            tabindex="0"
            role="button"
          >
            <div class="automation-header">
              <h3 class="automation-title">{config.friendlyName}</h3>
              <button
                class="delete-btn"
                on:click={(e) => handleDeleteClick(config, e)}
                type="button"
                title="Delete all backups"
                aria-label="Delete all backups"
              >
                🗑️
              </button>
            </div>

            <div class="automation-stats">
              <div class="stat">
                <span class="stat-label">Backups</span>
                <span class="stat-value">{config.backupCount}</span>
              </div>
              <div class="stat">
                <span class="stat-label">Total Size</span>
                <span class="stat-value"
                  >{formatFileSize(config.backupsSize)}</span
                >
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>

{#if showDeleteConfirm}
  <div
    class="modal-overlay"
    on:click={cancelDelete}
    on:keydown={(e) => e.key === "Escape" && cancelDelete()}
    role="presentation"
  >
    <div
      class="modal-content"
      on:click={(e) => e.stopPropagation()}
      on:keydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      aria-labelledby="delete-modal-title"
      tabindex="-1"
    >
      <h3 id="delete-modal-title">Delete All Backups?</h3>
      <h3>Delete All Backups?</h3>
      <p>Are you sure you want to delete ALL backups for this config?</p>
      {#if configToDelete}
        <p class="config-info">{configToDelete.friendlyName}</p>
        <p class="warning-text">
          ⚠️ This will delete {configToDelete.backupCount} backup{configToDelete.backupCount !==
          1
            ? "s"
            : ""} and cannot be undone!
        </p>
      {/if}
      <div class="modal-actions">
        <button
          class="cancel-btn"
          on:click={cancelDelete}
          type="button"
          disabled={deleting}
        >
          Cancel
        </button>
        <button
          class="confirm-delete-btn"
          on:click={confirmDelete}
          type="button"
          disabled={deleting}
        >
          {deleting ? "Deleting..." : "Delete All"}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .automation-list {
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--ha-card-background, #1c1c1e);
  }

  .filter-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--ha-card-background, #1c1c1e);
    min-height: 140px;
  }

  .search-box {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }

  .search-input {
    width: 80%;
    padding: 0.6rem;
    background: var(--ha-card-background, #2c2c2e);
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    border-radius: 4px;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.9rem;
  }

  .search-input:focus {
    outline: none;
    border-color: var(--primary-color, #03a9f4);
  }

  .group-filter-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .refresh-btn {
    background: transparent;
    color: var(--primary-color, #03a9f4);
    border: 1px solid var(--primary-color, #03a9f4);
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 1.2rem;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .refresh-btn:hover {
    background: var(--primary-color, #03a9f4);
    color: white;
  }

  .group-select {
    flex: 1;
    padding: 0.5rem;
    background: var(--ha-card-background, #2c2c2e);
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    border-radius: 4px;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.9rem;
    cursor: pointer;
  }

  .group-select:focus {
    outline: none;
    border-color: var(--primary-color, #03a9f4);
  }

  .filter-count {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.85rem;
    white-space: nowrap;
    align-content: end;
  }

  .grid {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .automation-card {
    background: var(--ha-card-background, #1c1c1e);
    border: 1px solid var(--ha-card-border-color, #2c2c2e);
    border-radius: 8px;
    padding: 0.5rem 1rem;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
    outline: none;
  }

  .automation-card:hover,
  .automation-card:focus {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    border-color: var(--primary-color, #03a9f4);
  }

  .automation-card.selected {
    border-color: var(--primary-color, #03a9f4);
    background: rgba(3, 169, 244, 0.1);
  }

  .automation-header {
    margin-bottom: 0.5rem;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .automation-title {
    color: var(--primary-text-color, #ffffff);
    font-size: 1.1rem;
    font-weight: 500;
    margin: 0 0 0.5rem 0;
    line-height: 1.3;
    flex: 1;
  }

  .delete-btn {
    background: transparent;
    border: none;
    color: var(--error-color, #f44336);
    cursor: pointer;
    font-size: 1.1rem;
    padding: 0.25rem;
    border-radius: 4px;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.7;
    flex-shrink: 0;
  }

  .delete-btn:hover {
    opacity: 1;
    background: rgba(244, 67, 54, 0.1);
    transform: scale(1.1);
  }

  .automation-stats {
    display: flex;
    gap: 1rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .stat-label {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .stat-value {
    color: var(--primary-text-color, #ffffff);
    font-size: 1rem;
    font-weight: 400;
  }

  .config-info {
    font-weight: 600;
    background: var(--ha-card-border-color, #2c2c2e);
    padding: 0.5rem;
    border-radius: 4px;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.95rem;
  }

  .warning-text {
    color: var(--error-color, #f44336);
    font-weight: 500;
    font-size: 0.9rem;
  }

  @media (max-width: 768px) {
    .filter-section {
      flex-direction: column;
      align-items: stretch;
      gap: 0.5rem;
    }

    .filter-count {
      text-align: center;
    }

    .automation-card {
      padding: 0.6rem 0.8rem;
    }
  }
</style>
