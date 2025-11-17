<script>
  let { type = 'info', message = '', dismissable = false, onDismiss = null } = $props();

  const alertClass = $derived(`alert alert-${type}`);
</script>

{#if message}
  <div class={alertClass}>
    <div class="alert-content">
      <slot>{message}</slot>
    </div>
    {#if dismissable && onDismiss}
      <button class="alert-dismiss" onclick={onDismiss} aria-label="Dismiss alert">
        ×
      </button>
    {/if}
  </div>
{/if}

<style>
  .alert {
    padding: 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    position: relative;
  }

  .alert-content {
    flex: 1;
  }

  .alert-dismiss {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    margin-left: 1rem;
    opacity: 0.7;
    transition: opacity 0.2s ease;
  }

  .alert-dismiss:hover {
    opacity: 1;
  }

  .alert-error {
    background: rgba(244, 67, 54, 0.1);
    border: 1px solid var(--error-color, #f44336);
    color: var(--error-color, #f44336);
  }

  .alert-warning {
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid var(--warning-color, #ff9800);
    color: var(--warning-color, #ff9800);
  }

  .alert-success {
    background: rgba(76, 175, 80, 0.1);
    border: 1px solid var(--success-color, #4caf50);
    color: var(--success-color, #4caf50);
  }

  .alert-info {
    background: rgba(33, 150, 243, 0.1);
    border: 1px solid var(--info-color, #2196f3);
    color: var(--info-color, #2196f3);
  }

  .alert :global(ul) {
    margin: 0.5rem 0;
    padding-left: 1.5rem;
  }

  .alert :global(li) {
    margin: 0.25rem 0;
  }
</style>