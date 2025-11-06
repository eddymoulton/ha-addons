<script lang="ts">
  import { onMount } from "svelte";
  import type { ConfigMetadata, BackupInfo } from "./types";
  import { api } from "./api";
  import { formatFileSize, formatRelativeTime } from "./utils";
  import LoadingState from "./LoadingState.svelte";

  export let config: ConfigMetadata | null = null;
  export let onBackupClick: (backup: BackupInfo, allBackups: BackupInfo[]) => void;
  export let selectedBackup: BackupInfo | null = null;
  export let onBack: (() => void) | undefined = undefined;

  let backups: BackupInfo[] = [];
  let loading = false;
  let error: string | null = null;
  let isMobile = false;

  function checkMobile() {
    isMobile = window.innerWidth <= 1024;
  }

  onMount(() => {
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  });

  $: if (config) {
    loadBackups();
  }

  async function loadBackups() {
    if (!config) return;

    loading = true;
    error = null;

    try {
      backups = await api.getConfigBackups(config.group, config.id);

      // Automatically select the first backup
      if (backups.length > 0) {
        onBackupClick(backups[0], backups);
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load backups";
    } finally {
      loading = false;
    }
  }

  function handleBackupClick(backup: BackupInfo) {
    onBackupClick(backup, backups);
  }
</script>

<div class="backup-list-container">
  <div class="header">
    <div class="header-row">
      {#if onBack && isMobile}
        <button
          class="back-btn"
          on:click={onBack}
          type="button"
          aria-label="Back to configs"
        >
          ← Back
        </button>
      {/if}
      <h2>{config ? config.friendlyName : "Select an config"}</h2>
      {#if config}
        <button
          class="refresh-btn"
          on:click={loadBackups}
          type="button"
          title="Refresh backups"
          aria-label="Refresh backups"
        >
          ⟳
        </button>
      {/if}
    </div>
  </div>

  <LoadingState
    {loading}
    {error}
    empty={!config || (!loading && !error && backups.length === 0)}
    emptyMessage={!config ? "Select an config to view backups" : "No backups found for this config"}
    loadingMessage="Loading backups..."
  />

  {#if config && !loading && !error && backups.length > 0}
    <div class="backup-list">
      {#each backups as backup, index (backup.filename)}
        <div
          class="backup-item {index === 0
            ? 'current'
            : ''} {selectedBackup?.filename === backup.filename
            ? 'selected'
            : ''}"
          on:click={() => handleBackupClick(backup)}
          on:keydown={(e) => e.key === "Enter" && handleBackupClick(backup)}
          tabindex="0"
          role="button"
        >
          <div class="backup-header">
            <div class="backup-filename">
              {backup.date}
              {#if index === 0}
                <span class="current-badge">Current</span>
              {/if}
            </div>
            <div class="backup-size">{formatFileSize(backup.size)}</div>
          </div>

          <div class="backup-date">
            {formatRelativeTime(backup.date)}
          </div>
        </div>
      {/each}
    </div>

    <div class="footer">
      <p class="backup-count">
        {backups.length} backup{backups.length !== 1 ? "s" : ""} total
      </p>
    </div>
  {/if}
</div>

<style>
  .backup-list-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--ha-card-background, #1c1c1e);
  }

  .header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
  }

  .header-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .back-btn {
    background: transparent;
    color: var(--primary-color, #03a9f4);
    border: 1px solid var(--primary-color, #03a9f4);
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s;
  }

  .back-btn:hover {
    background: var(--primary-color, #03a9f4);
    color: white;
  }

  .header h2 {
    margin: 0;
    color: var(--primary-text-color, #ffffff);
    font-size: 1.2rem;
    font-weight: 500;
    flex: 1;
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
    flex-shrink: 0;
  }

  .refresh-btn:hover {
    background: var(--primary-color, #03a9f4);
    color: white;
  }

  .backup-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .backup-item {
    background: var(--ha-card-background, #2c2c2e);
    border: 1px solid var(--ha-card-border-color, #3c3c3e);
    border-radius: 6px;
    padding: 1rem;
    margin-bottom: 0.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
    outline: none;
  }

  .backup-item:hover,
  .backup-item:focus {
    background: var(--ha-card-border-color, #3c3c3e);
    border-color: var(--primary-color, #03a9f4);
    transform: translateX(4px);
  }

  .backup-item.selected {
    border-color: var(--primary-color, #03a9f4);
    background: rgba(3, 169, 244, 0.1);
  }

  .backup-item.current {
    border-color: var(--success-color, #4caf50);
    background: rgba(76, 175, 80, 0.1);
  }

  .backup-item.current:hover,
  .backup-item.current:focus {
    background: rgba(76, 175, 80, 0.2);
  }

  .backup-item.current.selected {
    background: rgba(76, 175, 80, 0.15);
  }

  .backup-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .backup-filename {
    color: var(--primary-text-color, #ffffff);
    font-family: monospace;
    font-size: 0.9rem;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .current-badge {
    background: var(--success-color, #4caf50);
    color: white;
    padding: 0.15rem 0.4rem;
    border-radius: 12px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .backup-size {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.85rem;
    font-weight: 500;
  }

  .backup-date {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.8rem;
  }

  .footer {
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
  }

  .backup-count {
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 0.9rem;
    margin: 0;
    text-align: center;
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
