<script lang="ts">
  import { onMount } from "svelte";
  import type { ConfigMetadata, BackupInfo } from "./types";
  import { api } from "./api";
  import { formatFileSize, formatRelativeTime, getErrorMessage } from "./utils";
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
      error = getErrorMessage(err, "Failed to load backups");
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

{#snippet leftContent()}
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
{/snippet}

{#snippet rightContent()}
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
{/snippet}

{#snippet subtitleContent()}
  <div class="backup-count">
    {backups.length} backup{backups.length !== 1 ? "s" : ""} total
  </div>
{/snippet}

<ListContainer>
  <ListHeader
    title={config ? config.friendlyName : "Select an config"}
    left={leftContent}
    right={rightContent}
    subtitleSnippet={subtitleContent}
  />

  <ListContent>
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
      <div class="list-grid">
        {#each backups as backup, index (backup.filename)}
          {#snippet actions()}
            <IconButton
              icon="🗑️"
              variant="ghost"
              size="small"
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
  .list-grid {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .current-badge {
    background: var(--success-color);
    color: white;
    padding: 0.15rem 0.4rem;
    margin-left: 1rem;
    border-radius: 12px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .backup-size {
    color: var(--secondary-text-color);
    font-size: 0.85rem;
    font-weight: 500;
    min-width: 50px;
    text-align: right;
  }

  .backup-date {
    color: var(--secondary-text-color);
    font-size: 0.8rem;
    display: flex;
    gap: 0.5rem;
  }

  .backup-count {
    color: var(--secondary-text-color);
    font-size: 0.85rem;
    white-space: nowrap;
    align-content: end;
  }

  .backup-info {
    font-family: monospace;
    background: var(--ha-card-border-color);
    padding: 0.5rem;
    border-radius: 4px;
    color: var(--primary-text-color);
    font-size: 0.85rem;
  }
</style>
