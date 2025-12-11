<script lang="ts">
  import { onMount } from "svelte";
  import type {
    ConfigMetadata,
    BackupInfo,
    BackupDiffResponse,
    ComparisonMode,
    RestoreBackupResponse,
  } from "./types";
  import { api } from "./api";
  import Button from "./components/Button.svelte";
  import { formatRelativeTime, getErrorMessage } from "./utils";
  import LoadingState from "./LoadingState.svelte";
  import Alert from "./components/Alert.svelte";

  type DiffViewerProps = {
    config: ConfigMetadata | null;
    selectedBackup: BackupInfo | null;
    allBackups: BackupInfo[];
    onBack?: () => void;
  };

  let {
    config = null,
    selectedBackup = null,
    allBackups = [],
    onBack = undefined,
  }: DiffViewerProps = $props();

  let comparisonMode: ComparisonMode = $state("current");
  let secondBackupFilename: string | null = $state(null);
  let diffData: BackupDiffResponse | null = $state(null);
  let loading = $state(false);
  let error: string | null = $state(null);
  let restoringBackup: string | null = $state(null);
  let restoreSuccess: string | null = $state(null);
  let isMobile = $state(false);

  let currentBackup = $derived(allBackups.length > 0 ? allBackups[0] : null);
  let secondBackup: BackupInfo | null = $derived(
    secondBackupFilename
      ? allBackups.find((b) => b.filename === secondBackupFilename) || null
      : null
  );

  function checkMobile() {
    isMobile = window.innerWidth <= 1024;
  }

  onMount(() => {
    checkMobile();
    window.addEventListener("resize", checkMobile);
    return () => window.removeEventListener("resize", checkMobile);
  });

  $effect(() => {
    loadDiff();
  });

  async function loadDiff() {
    if (!config || !selectedBackup) return;

    loading = true;
    error = null;

    try {
      switch (comparisonMode) {
        case "previous": {
          // Find the previous backup (next in the list since sorted newest first)
          const currentIdx = allBackups.findIndex(
            (b) => b.filename === selectedBackup.filename
          );
          if (currentIdx >= 0 && currentIdx < allBackups.length - 1) {
            // There's a previous backup
            const previousBackup = allBackups[currentIdx + 1];
            diffData = await api.compareBackups(
              config.path,
              config.id,
              previousBackup.filename,
              selectedBackup.filename
            );
          } else {
            // No previous backup, just show the content
            diffData = {
              type: "content",
              content: await api.getBackupContent(
                config.path,
                config.id,
                selectedBackup.filename
              ),
              isFirstBackup: true,
            };
          }
          break;
        }
        case "current":
          if (
            currentBackup &&
            selectedBackup.filename !== currentBackup.filename
          ) {
            diffData = await api.compareBackups(
              config.path,
              config.id,
              selectedBackup.filename,
              currentBackup.filename
            );
          } else {
            diffData = {
              type: "content",
              content: await api.getBackupContent(
                config.path,
                config.id,
                selectedBackup.filename
              ),
              isFirstBackup: false,
            };
          }
          break;
        case "two-backups":
          if (secondBackup) {
            diffData = await api.compareBackups(
              config.path,
              config.id,
              secondBackup.filename,
              selectedBackup.filename
            );
          }
          break;
      }
    } catch (err) {
      error = getErrorMessage(err, "Failed to load diff");
    } finally {
      loading = false;
    }
  }

  function handleComparisonModeChange(mode: ComparisonMode) {
    comparisonMode = mode;
    loadDiff();
  }

  function renderDiff(unifiedDiff: string): string {
    const lines = unifiedDiff.split("\n");
    let lineNumber = 0;
    return lines
      .map((line) => {
        let lineClass = "";
        let showLineNumber = true;

        if (line.startsWith("+") && !line.startsWith("+++")) {
          lineClass = "added";
          lineNumber++;
        } else if (line.startsWith("-") && !line.startsWith("---")) {
          lineClass = "removed";
          lineNumber++;
        } else if (line.startsWith("@@")) {
          lineClass = "context";
          showLineNumber = false;
          // Extract line number from context line (e.g., @@ -1,3 +1,4 @@)
          const match = line.match(/@@ -(\d+),?\d* \+(\d+),?\d* @@/);
          if (match) {
            lineNumber = parseInt(match[2], 10);
          }
        } else if (line.startsWith("---") || line.startsWith("+++")) {
          lineClass = "file-header";
          showLineNumber = false;
        } else {
          lineClass = "unchanged";
          lineNumber++;
        }

        const lineNumDisplay = showLineNumber
          ? `<span class="line-number">${lineNumber}</span>`
          : '<span class="line-number"></span>';
        return `<div class="diff-line ${lineClass}">${lineNumDisplay}<span class="line-content">${escapeHtml(line)}</span></div>`;
      })
      .join("");
  }

  function escapeHtml(text: string): string {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }

  async function handleRestore(event: MouseEvent) {
    event.stopPropagation();

    if (!config || !selectedBackup) return;

    const confirmed = confirm(
      `Are you sure you want to restore backup "${selectedBackup.filename}"?\n\nThis will overwrite the current file.`
    );

    if (!confirmed) return;

    restoringBackup = selectedBackup.filename;
    error = null;
    restoreSuccess = null;

    try {
      const response: RestoreBackupResponse = await api.restoreBackup(
        config.path,
        config.id,
        selectedBackup.filename
      );

      if (response.success) {
        restoreSuccess = response.message || "Backup restored successfully!";
        setTimeout(() => {
          restoreSuccess = null;
        }, 5000);
      } else {
        error = response.error || "Failed to restore backup";
      }
    } catch (err) {
      error = getErrorMessage(err, "Failed to restore backup");
    } finally {
      restoringBackup = null;
    }
  }
</script>

<div class="diff-viewer-container">
  <div class="header">
    {#if onBack && isMobile}
      <Button
        label="Back"
        variant="outlined"
        size="small"
        onclick={onBack}
        type="button"
        aria-label="Back to backups"
        icon="←"
      ></Button>
    {/if}
    <div class="title-container">
      <h2>{selectedBackup ? selectedBackup.filename : "Select a backup"}</h2>
      <Button
        label="Restore"
        variant="success"
        size="small"
        onclick={handleRestore}
        type="button"
        disabled={selectedBackup?.filename === currentBackup?.filename}
        title={selectedBackup?.filename === currentBackup?.filename
          ? "Cannot restore current backup"
          : "Restore this backup"}
      ></Button>
    </div>
    <div class="diff-controls">
      <div class="comparison-modes">
        <Button
          label="vs Current"
          variant={comparisonMode === "current" ? "primary" : "secondary"}
          size="small"
          onclick={() => handleComparisonModeChange("current")}
          type="button"
        ></Button>
        <Button
          label="vs Previous"
          variant={comparisonMode === "previous" ? "primary" : "secondary"}
          size="small"
          onclick={() => handleComparisonModeChange("previous")}
          type="button"
        ></Button>
        <Button
          label="Compare Two"
          variant={comparisonMode === "two-backups" ? "primary" : "secondary"}
          size="small"
          onclick={() => handleComparisonModeChange("two-backups")}
          type="button"
        ></Button>
      </div>

      {#if comparisonMode === "two-backups"}
        <div class="backup-selector">
          <label for="second-backup">Compare with:</label>
          <select
            id="second-backup"
            bind:value={secondBackupFilename}
            onchange={loadDiff}
          >
            <option value={null}>Select backup...</option>
            {#each allBackups as backup (backup.filename)}
              {#if backup.filename !== selectedBackup?.filename}
                <option value={backup.filename}>{backup.filename}</option>
              {/if}
            {/each}
          </select>
        </div>
      {/if}
    </div>
  </div>

  <LoadingState
    empty={!selectedBackup}
    emptyMessage="Select a backup to view diff"
  />

  {#if selectedBackup}
    <div class="diff-viewer">
      <LoadingState {loading} {error} loadingMessage="Loading diff..." />

      {#if !loading && !error}
        {#if restoreSuccess}
          <Alert type="success" message={restoreSuccess} />
        {:else if diffData}
          <div class="diff-content">
            {#if diffData.type === "diff"}
              <div class="diff-header">
                <div class="file-info">
                  <span class="file-old">
                    {diffData.oldFilename || "Previous"}
                    {#if diffData.oldFilename}
                      <small>
                        ({formatRelativeTime(
                          allBackups.find(
                            (b) => b.filename === diffData!.oldFilename
                          )?.date || ""
                        )})
                      </small>
                    {/if}
                  </span>
                  →
                  <span class="file-new">
                    {diffData.newFilename || "Current"}
                    {#if diffData.newFilename}
                      <small>
                        ({formatRelativeTime(
                          allBackups.find(
                            (b) => b.filename === diffData!.newFilename
                          )?.date || ""
                        )})
                      </small>
                    {/if}
                  </span>
                </div>
              </div>
              <div class="diff-body">
                {@html renderDiff(diffData.unifiedDiff || "")}
              </div>
            {:else}
              <div class="content-view">
                <div class="content-header">
                  <h4>Raw Content</h4>
                  {#if diffData.isFirstBackup}
                    <span class="first-backup-notice">
                      This is the first backup - no previous version to compare
                    </span>
                  {/if}
                </div>
                <pre class="content-body">{diffData.content}</pre>
              </div>
            {/if}
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>

<style>
  .diff-viewer-container {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 84px);
    background: var(--ha-card-background);
  }

  .title-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color);
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--ha-card-background);
    min-height: 140px;
  }

  .header h2 {
    margin: 0;
    color: var(--primary-text-color);
    font-size: 1.2rem;
    font-weight: 500;
    word-break: break-all;
  }

  .diff-viewer {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    padding: 1rem;
  }

  .diff-controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    flex-shrink: 0;
    flex-wrap: wrap;
    position: sticky;
    top: 0;
    z-index: 9;
    background: var(--ha-card-background);
  }

  .comparison-modes {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .backup-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .backup-selector label {
    color: var(--secondary-text-color);
    font-size: 0.9rem;
  }

  .backup-selector select {
    background: var(--ha-card-background);
    color: var(--primary-text-color);
    border: 1px solid var(--ha-card-border-color);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.9rem;
  }

  .diff-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .diff-header {
    margin-bottom: 1rem;
    flex-shrink: 0;
  }

  .file-info {
    display: flex;
    align-items: center;
    gap: 1rem;
    color: var(--secondary-text-color);
    font-size: 0.9rem;
    flex-wrap: wrap;
  }

  .file-old,
  .file-new {
    font-family: monospace;
    color: var(--primary-text-color);
  }

  .file-old small,
  .file-new small {
    color: var(--secondary-text-color);
    font-size: 0.8em;
  }

  .diff-body {
    flex: 1;
    overflow-y: auto;
    background: var(--ha-card-background);
    border: 1px solid var(--ha-card-border-color);
    border-radius: 4px;
    font-family: "Courier New", monospace;
    font-size: 0.85rem;
    line-height: 1.4;
  }

  :global(.diff-line) {
    padding: 0.1rem 0.5rem;
    white-space: pre;
    border-left: 3px solid transparent;
    display: flex;
    align-items: center;
  }

  :global(.diff-line .line-number) {
    display: inline-block;
    min-width: 3rem;
    margin-right: 1rem;
    text-align: right;
    color: var(--secondary-text-color);
    user-select: none;
    flex-shrink: 0;
  }

  :global(.diff-line .line-content) {
    flex: 1;
    white-space: pre;
  }

  :global(.diff-line.added) {
    background: rgba(76, 175, 80, 0.2);
    color: var(--success-color);
    border-left-color: var(--success-color);
  }

  :global(.diff-line.removed) {
    background: rgba(244, 67, 54, 0.2);
    color: var(--error-color);
    border-left-color: var(--error-color);
  }

  :global(.diff-line.context) {
    background: rgba(3, 169, 244, 0.1);
    color: var(--primary-color);
    font-weight: 600;
  }

  :global(.diff-line.file-header) {
    background: var(--ha-card-border-color);
    color: var(--secondary-text-color);
    font-weight: 600;
  }

  :global(.diff-line.unchanged) {
    color: var(--primary-text-color);
  }

  .content-view {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .content-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    flex-shrink: 0;
  }

  .content-header h4 {
    color: var(--primary-text-color);
    margin: 0;
  }

  .first-backup-notice {
    color: var(--warning-color);
    font-size: 0.9rem;
    font-style: italic;
  }

  .content-body {
    flex: 1;
    background: var(--ha-card-background);
    border: 1px solid var(--ha-card-border-color);
    border-radius: 4px;
    padding: 1rem;
    overflow-y: auto;
    color: var(--primary-text-color);
    font-family: "Courier New", monospace;
    font-size: 0.85rem;
    line-height: 1.4;
    margin: 0;
  }

  @media (max-width: 1024px) {
    .diff-controls {
      flex-direction: column;
      align-items: stretch;
      gap: 0.5rem;
    }

    .comparison-modes {
      justify-content: center;
    }
  }
</style>
