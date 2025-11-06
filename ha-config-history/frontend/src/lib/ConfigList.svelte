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
      <div class="search-box">
        <input
          type="text"
          placeholder="Search configs..."
          bind:value={searchQuery}
          class="search-input"
        />
      </div>
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
      <div class="filter-count">
        {filteredConfigs.length} config{filteredConfigs.length !== 1 ? "s" : ""}
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
    padding: 1rem;
    border-bottom: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
  }

  .search-box {
    width: 100%;
  }

  .search-input {
    width: 100%;
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
    padding: 0.75rem 1rem;
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
  }

  .automation-title {
    color: var(--primary-text-color, #ffffff);
    font-size: 1.1rem;
    font-weight: 500;
    margin: 0 0 0.5rem 0;
    line-height: 1.3;
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
