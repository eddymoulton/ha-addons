<script lang="ts">
  import ConfigList from "./ConfigList.svelte";
  import BackupList from "./BackupList.svelte";
  import DiffViewer from "./DiffViewer.svelte";
  import SettingsModal from "./SettingsModal.svelte";
  import ResizeHandle from "./ResizeHandle.svelte";
  import Button from "./components/Button.svelte";
  import type { ConfigMetadata, BackupInfo } from "./types";

  let selectedConfig: ConfigMetadata | null = $state(null);
  let selectedBackup: BackupInfo | null = $state(null);
  let allBackups: BackupInfo[] = $state([]);
  let showSettings = $state(false);

  // Column widths (in pixels)
  const MIN_COLUMN_WIDTH = 250;
  const DEFAULT_CONFIG_WIDTH = 600;
  const DEFAULT_BACKUP_WIDTH = 350;

  let configColumnWidth = $state(DEFAULT_CONFIG_WIDTH);
  let backupColumnWidth = $state(DEFAULT_BACKUP_WIDTH);

  async function handleConfigClick(config: ConfigMetadata) {
    selectedConfig = config;
    selectedBackup = null;
    allBackups = [];
  }

  async function handleBackupClick(backup: BackupInfo, backups: BackupInfo[]) {
    selectedBackup = backup;
    allBackups = backups;
  }

  function handleBackToConfigs() {
    selectedConfig = null;
    selectedBackup = null;
    allBackups = [];
  }

  function handleBackToBackups() {
    selectedBackup = null;
    allBackups = [];
  }

  function handleOpenSettings() {
    showSettings = true;
  }

  function handleCloseSettings() {
    showSettings = false;
  }

  function handleConfigResize(event: CustomEvent<{ deltaX: number }>) {
    configColumnWidth = Math.max(
      MIN_COLUMN_WIDTH,
      configColumnWidth + event.detail.deltaX
    );
    saveColumnWidths();
  }

  function handleBackupResize(event: CustomEvent<{ deltaX: number }>) {
    backupColumnWidth = Math.max(
      MIN_COLUMN_WIDTH,
      backupColumnWidth + event.detail.deltaX
    );
    saveColumnWidths();
  }

  function saveColumnWidths() {
    localStorage.setItem(
      "columnWidths",
      JSON.stringify({
        config: configColumnWidth,
        backup: backupColumnWidth,
      })
    );
  }
</script>

<main class="app">
  <header class="app-header">
    <h1>Home Assistant Config History</h1>
    <Button
      label="Settings"
      variant="primary"
      size="small"
      type="button"
      onclick={handleOpenSettings}
    />
  </header>

  <div
    class="three-column-layout"
    class:has-config={selectedConfig}
    class:has-backup={selectedBackup}
    style="--config-width: {configColumnWidth}px; --backup-width: {backupColumnWidth}px;"
  >
    <div class="column column-configs">
      <ConfigList onConfigClick={handleConfigClick} {selectedConfig} />
    </div>

    <ResizeHandle on:resize={handleConfigResize} />

    <div class="column column-backups">
      <BackupList
        config={selectedConfig}
        onBackupClick={handleBackupClick}
        {selectedBackup}
        onBack={handleBackToConfigs}
      />
    </div>

    <ResizeHandle on:resize={handleBackupResize} />

    <div class="column column-diff">
      <DiffViewer
        config={selectedConfig}
        {selectedBackup}
        {allBackups}
        onBack={handleBackToBackups}
      />
    </div>
  </div>

  <SettingsModal isOpen={showSettings} onClose={handleCloseSettings} />
</main>

<style>
  :global(html) {
    --primary-color: #03a9f4;
    --primary-color-dark: #0288d1;
    --success-color: #4caf50;
    --success-color-dark: #45a049;
    --error-color: #f44336;
    --warning-color: #ff9800;
    --primary-text-color: #ffffff;
    --secondary-text-color: #9b9b9b;
    --ha-card-background: #1c1c1e;
    --ha-card-border-color: #2c2c2e;
    --breakpoint-mobile: 1024px;
    --breakpoint-tablet: 1440px;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    background: #0c0c0e;
    color: var(--primary-text-color);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      sans-serif;
    line-height: 1.5;
  }

  :global(*) {
    box-sizing: border-box;
  }

  :global(:focus) {
    outline: 2px solid var(--primary-color);
    outline-offset: 2px;
  }

  :global(:focus:not(:focus-visible)) {
    outline: none;
  }

  :global(button, select, input) {
    font-family: inherit;
  }

  :global(button:disabled) {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(.sr-only) {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  .app-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    background: var(--ha-card-background, #1c1c1e);
    border-bottom: 2px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 100;
    height: 84px;
  }

  .app-header h1 {
    color: var(--primary-text-color, #ffffff);
    font-size: 1.5rem;
    margin: 0;
    font-weight: 500;
  }

  .three-column-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .column {
    overflow: hidden;
    display: flex;
    flex-direction: column;
    min-width: 0; /* Allow flex items to shrink below content size */
  }

  .column-configs {
    width: var(--config-width, 350px);
    flex-shrink: 0;
  }

  .column-backups {
    width: var(--backup-width, 400px);
    flex-shrink: 0;
  }

  .column-diff {
    flex: 1;
    min-width: 300px;
  }

  /* Mobile-first responsive behavior */
  /* Uses --breakpoint-mobile (1024px) */
  @media (max-width: 1024px) {
    .app-header h1 {
      font-size: 1.2rem;
    }

    .three-column-layout {
      flex-direction: column;
    }

    .column-configs,
    .column-backups,
    .column-diff {
      width: 100%;
      flex: 1;
    }

    /* Hide resize handles on mobile */
    .three-column-layout :global(.resize-handle) {
      display: none;
    }

    /* Hide columns based on selection state */
    .column-backups,
    .column-diff {
      display: none;
    }

    /* Show backups when config is selected */
    .three-column-layout.has-config .column-configs {
      display: none;
    }

    .three-column-layout.has-config .column-backups {
      display: flex;
    }

    /* Show diff when backup is selected */
    .three-column-layout.has-backup .column-backups {
      display: none;
    }

    .three-column-layout.has-backup .column-diff {
      display: flex;
    }
  }

  /* Tablet view: show 2 columns */
  /* Uses --breakpoint-mobile + 1px (1025px) and --breakpoint-tablet (1440px) */
  @media (min-width: 1025px) and (max-width: 1440px) {
    /* Hide diff column and second resize handle until backup is selected */
    .column-diff {
      display: none;
    }

    .three-column-layout :global(.resize-handle:last-of-type) {
      display: none;
    }

    .three-column-layout.has-backup .column-backups {
      display: flex;
    }

    .three-column-layout.has-backup .column-diff {
      display: flex;
    }

    .three-column-layout.has-backup :global(.resize-handle:first-of-type) {
      display: block;
    }

    .three-column-layout.has-backup :global(.resize-handle:last-of-type) {
      display: block;
    }
  }

  /* Desktop view: show all 3 columns when needed */
  /* Uses --breakpoint-tablet + 1px (1441px) */
  @media (min-width: 1441px) {
    /* Show all columns on desktop, but adjust visibility based on state */
    .column-backups {
      display: none;
    }

    .column-diff {
      display: none;
    }

    /* Hide resize handles for columns that aren't shown */
    .three-column-layout :global(.resize-handle) {
      display: none;
    }

    .three-column-layout.has-config .column-backups {
      display: flex;
    }

    .three-column-layout.has-config :global(.resize-handle:first-of-type) {
      display: block;
    }

    .three-column-layout.has-backup .column-diff {
      display: flex;
    }

    .three-column-layout.has-backup :global(.resize-handle) {
      display: block;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    :global(*) {
      animation-duration: 0.01ms !important;
      animation-iteration-count: 1 !important;
      transition-duration: 0.01ms !important;
    }
  }
</style>
