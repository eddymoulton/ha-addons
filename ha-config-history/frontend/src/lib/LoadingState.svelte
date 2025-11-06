<script lang="ts">
  export let loading: boolean = false;
  export let error: string | null = null;
  export let empty: boolean = false;
  export let emptyMessage: string = "No items found";
  export let loadingMessage: string = "Loading...";
</script>

{#if loading}
  <div class="loading" role="status" aria-live="polite">{loadingMessage}</div>
{:else if error}
  <div class="error" role="alert">Error: {error}</div>
{:else if empty}
  <div class="empty">{emptyMessage}</div>
{/if}

<style>
  .loading,
  .error,
  .empty {
    text-align: center;
    padding: 3rem 1.5rem;
    color: var(--secondary-text-color, #9b9b9b);
    font-size: 1rem;
  }

  .error {
    color: var(--error-color, #f44336);
  }
</style>
