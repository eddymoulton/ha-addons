<script lang="ts">
  import type { Snippet } from "svelte";

  /**
   * ListHeader - A reusable header component for lists
   * Provides consistent header styling with title and optional action props
   */
  type Props = {
    title?: string;
    subtitle?: string;
    left?: Snippet;
    right?: Snippet;
    subtitleSnippet?: Snippet;
  };
  let {
    title = undefined,
    subtitle = undefined,
    left = undefined,
    right = undefined,
    subtitleSnippet = undefined,
  }: Props = $props();

  const headerClass = $derived(["list-header"].filter(Boolean).join(" "));
</script>

<div class={headerClass}>
  <div class="header-row">
    {#if left}
      {@render left()}
    {/if}
    {#if title}
      <h2>{title}</h2>
    {/if}
    {#if right}
      {@render right()}
    {/if}
  </div>
  {#if subtitle}
    <div class="header-subtitle">
      <div class="subtitle-text">{subtitle}</div>
    </div>
  {:else if subtitleSnippet}
    <div class="header-subtitle">
      {@render subtitleSnippet()}
    </div>
  {/if}
</div>

<style>
  .list-header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--ha-card-background);
    min-height: 140px;
  }

  .header-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .header-row h2 {
    margin: 0;
    color: var(--primary-text-color);
    font-size: 1.2rem;
    font-weight: 500;
    flex: 1;
  }

  .header-subtitle {
    margin-top: 0.5rem;
  }

  .subtitle-text {
    color: var(--secondary-text-color);
    font-size: 0.85rem;
  }
</style>
