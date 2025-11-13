<script lang="ts">
  import { onMount } from "svelte";
  import type { ConfigMetadata } from "./types";
  import { api } from "./api";
  import { formatFileSize } from "./utils";
  import LoadingState from "./LoadingState.svelte";
  import Button from "./components/Button.svelte";
  import IconButton from "./components/IconButton.svelte";
  import ListContainer from "./components/ListContainer.svelte";
  import FilterSection from "./components/FilterSection.svelte";
  import ListContent from "./components/ListContent.svelte";
  import ListItem from "./components/ListItem.svelte";
  import ConfirmationModal from "./components/ConfirmationModal.svelte";

  type ConfigListProps = {
    onConfigClick: (config: ConfigMetadata) => void;
    selectedConfig: ConfigMetadata | null;
  };

  let { onConfigClick, selectedConfig }: ConfigListProps = $props();

  let configs: ConfigMetadata[] = $state([]);
  let loading = $state(false);
  let error: string | null = $state(null);
  let selectedGroup: string = $state("all");
  let searchQuery: string = $state("");
  let configToDelete: ConfigMetadata | null = $state(null);
  let showDeleteConfirm = $state(false);
  let deleting = $state(false);

  let groups = $derived(
    Array.from(new Set(configs.map((c) => c.group))).sort()
  );
  let filteredConfigs = $derived(
    configs
      .filter((c) => selectedGroup === "all" || c.group === selectedGroup)
      .filter(
        (c) =>
          searchQuery === "" ||
          c.friendlyName.toLowerCase().includes(searchQuery.toLowerCase()) ||
          c.id.toLowerCase().includes(searchQuery.toLowerCase())
      )
  );

  async function loadConfigs() {
    loading = true;
    error = null;
    try {
      configs = await api.getConfigs();
      selectedGroup = groups[0];
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load configs";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadConfigs();
  });

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

<ListContainer>
  <FilterSection slot="filter">
    <LoadingState
      {loading}
      {error}
      empty={!loading && !error && configs.length === 0}
      emptyMessage="No configs found"
      loadingMessage="Loading configs..."
    />

    {#if !loading && !error && configs.length > 0}
      <div class="group-filter-row">
        <select
          id="group-filter"
          bind:value={selectedGroup}
          class="group-select"
        >
          {#each groups as group}
            <option value={group}>{group}</option>
          {/each}
        </select>
        <Button
          variant="outlined"
          size="small"
          onclick={loadConfigs}
          type="button"
          title="Refresh configs"
          aria-label="Refresh configs"
          icon="⟳"
        >
          Refresh
        </Button>
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
    {/if}
  </FilterSection>

  <ListContent slot="content">
    {#if !loading && !error && configs.length > 0}
      {#if filteredConfigs.length === 0}
        <LoadingState empty={true} emptyMessage="No configs in this group" />
      {:else}
        <div class="grid">
          {#each filteredConfigs as config (config.id)}
            <ListItem
              selected={selectedConfig?.id === config.id}
              hoverTransform="lift"
              on:click={() => onConfigClick(config)}
              on:keydown={(e) => e.key === "Enter" && onConfigClick(config)}
            >
              <div slot="title" class="automation-title">{config.friendlyName}</div>
              <div slot="content" class="automation-stats">
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
              <div slot="actions">
                <IconButton
                  icon="🗑️"
                  variant="ghost"
                  size="small"
                  class="btn-danger"
                  onclick={(e) => handleDeleteClick(config, e)}
                  type="button"
                  title="Delete all backups"
                  aria-label="Delete all backups"
                />
              </div>
            </ListItem>
          {/each}
        </div>
      {/if}
    {/if}
  </ListContent>
</ListContainer>

<ConfirmationModal
  isOpen={showDeleteConfirm}
  title="Delete All Backups?"
  message="Are you sure you want to delete ALL backups for this config?"
  onClose={cancelDelete}
  onConfirm={confirmDelete}
  confirmText={deleting ? "Deleting..." : "Delete All"}
  variant="danger"
  disabled={deleting}
>
  {#if configToDelete}
    <p class="config-info">{configToDelete.friendlyName}</p>
    <p class="warning-text">
      ⚠️ This will delete {configToDelete.backupCount} backup{configToDelete.backupCount !== 1 ? "s" : ""} and cannot be undone!
    </p>
  {/if}
</ConfirmationModal>

<style>

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
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }


  .automation-title {
    color: var(--primary-text-color, #ffffff);
    font-size: 1rem;
    font-weight: 500;
    margin: 0;
    line-height: 1.3;
    flex: 1;
  }

  .automation-stats {
    display: flex;
    gap: 2rem;
  }

  .stat {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
  }

  .stat-label {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .stat-value {
    color: var(--primary-text-color, #ffffff);
    font-size: 0.9rem;
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

  }
</style>
