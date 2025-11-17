<script lang="ts">
  import { onMount } from "svelte";
  import type { ConfigMetadata, BackupInfo } from "./types";
  import { api } from "./api";
  import { formatFileSize, formatRelativeTime } from "./utils";
  import LoadingState from "./LoadingState.svelte";
  import Button from "./components/Button.svelte";
  import IconButton from "./components/IconButton.svelte";
  import ListContainer from "./components/ListContainer.svelte";
  import ListHeader from "./components/ListHeader.svelte";
  import ListContent from "./components/ListContent.svelte";
  import ListItem from "./components/ListItem.svelte";
  import ConfirmationModal from "./components/ConfirmationModal.svelte";

  type Props = {
    config: ConfigMetadata | null;
    onBackupClick: (backup: BackupInfo, allBackups: BackupInfo[]) => void;
    selectedBackup: BackupInfo | null;
    onBack?: (() => void) | undefined;
  };

  let {
    config = null,
    onBackupClick,
    selectedBackup = null,
    onBack = undefined,
  }: Props = $props();

  let backups: BackupInfo[] = $state([]);
  let loading = $state(false);
  let error: string | null = $state(null);
  let isMobile = $state(false);
  let backupToDelete: BackupInfo | null = $state(null);
  let showDeleteConfirm = $state(false);
  let deleting = $state(false);

  function checkMobile() {
    isMobile = window.innerWidth <= 1024;
  }

  onMount(() => {
    checkMobile();
    window.addEventListener("resize", checkMobile);
    return () => window.removeEventListener("resize", checkMobile);
  });

  $effect(() => {
    loadBackups();
  });

  async function loadBackups() {
    if (!config) return;

    loading = true;
    error = null;

    try {
      backups = await api.getConfigBackups(config.group, config.id);
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load backups";
    } finally {
      loading = false;
    }
  }

  function handleBackupClick(backup: BackupInfo) {
    onBackupClick(backup, backups);
  }

  function handleDeleteClick(backup: BackupInfo, event: Event) {
    event.stopPropagation();
    backupToDelete = backup;
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    showDeleteConfirm = false;
    backupToDelete = null;
  }

  async function confirmDelete() {
    if (!config || !backupToDelete) return;

    deleting = true;
    try {
      await api.deleteBackup(config.group, config.id, backupToDelete.filename);

      // Reload backups after successful deletion
      await loadBackups();

      showDeleteConfirm = false;
      backupToDelete = null;
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to delete backup";
    } finally {
      deleting = false;
    }
  }
</script>

<ListContainer>
  <ListHeader
    slot="header"
    title={config ? config.friendlyName : "Select an config"}
  >
    <svelte:fragment slot="left">
      {#if onBack && isMobile}
        <Button
          label="Back"
          variant="outlined"
          size="small"
          onclick={onBack}
          type="button"
          aria-label="Back to configs"
          icon="←"
        ></Button>
      {/if}
    </svelte:fragment>
    <svelte:fragment slot="right">
      {#if config}
        <Button
          label="Refresh"
          variant="outlined"
          size="small"
          onclick={loadBackups}
          type="button"
          title="Refresh backups"
          aria-label="Refresh backups"
          icon="⟳"
        ></Button>
      {/if}
    </svelte:fragment>
    <svelte:fragment slot="subtitle">
      <div class="backup-count">
        {backups.length} backup{backups.length !== 1 ? "s" : ""} total
      </div>
    </svelte:fragment>
  </ListHeader>

  <ListContent slot="content">
    <LoadingState
      {loading}
      {error}
      empty={!config || (!loading && !error && backups.length === 0)}
      emptyMessage={!config
        ? "Select an config to view backups"
        : "No backups found for this config"}
      loadingMessage="Loading backups..."
    />

    {#if config && !loading && !error && backups.length > 0}
      <div class="backup-list">
        {#each backups as backup, index (backup.filename)}
          {#snippet actions()}
            <IconButton
              icon="🗑️"
              variant="ghost"
              size="small"
              class="btn-danger"
              onclick={(e) => handleDeleteClick(backup, e)}
              type="button"
              title="Delete backup"
              aria-label="Delete backup"
            />
          {/snippet}

          <ListItem
            selected={selectedBackup?.filename === backup.filename}
            variant={index === 0 ? "current" : "default"}
            hoverTransform="slide"
            onclick={() => handleBackupClick(backup)}
            onkeydown={(e) => e.key === "Enter" && handleBackupClick(backup)}
            title={backup.date}
            {actions}
          >
            <div class="backup-date">
              {formatRelativeTime(backup.date)}
              <div class="backup-size">{formatFileSize(backup.size)}</div>
              {#if index === 0}
                <span class="current-badge">Current</span>
              {/if}
            </div>
          </ListItem>
        {/each}
      </div>
    {/if}
  </ListContent>
</ListContainer>

<ConfirmationModal
  isOpen={showDeleteConfirm}
  title="Delete Backup?"
  message="Are you sure you want to delete this backup?"
  onClose={cancelDelete}
  onConfirm={confirmDelete}
  confirmText={deleting ? "Deleting..." : "Delete"}
  variant="danger"
  disabled={deleting}
>
  {#if backupToDelete}
    <p class="backup-info">{backupToDelete.filename}</p>
  {/if}
</ConfirmationModal>

<style>
  .backup-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .current-badge {
    background: var(--success-color, #4caf50);
    color: white;
    padding: 0.15rem 0.4rem;
    margin-left: 1rem;
    border-radius: 12px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .backup-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .backup-size {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.85rem;
    font-weight: 500;
    min-width: 50px;
    text-align: right;
  }

  .backup-date {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.8rem;
    display: flex;
    gap: 0.5rem;
  }

  .backup-count {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.85rem;
    white-space: nowrap;
    align-content: end;
  }

  .backup-info {
    font-family: monospace;
    background: var(--ha-card-border-color, #2c2c2e);
    padding: 0.5rem;
    border-radius: 4px;
    color: var(--primary-text-color, #ffffff);
    font-size: 0.85rem;
  }

  @media (max-width: 1024px) {
    .backup-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .backup-filename {
      font-size: 0.8rem;
    }
  }
</style>
