<script lang="ts">
  import type { Snippet } from "svelte";

  /**
   * ListContainer - A reusable container component for list layouts
   * Provides consistent structure with header, filter, and content areas
   */
  type Props = { children: Snippet };
  let { children }: Props = $props();

  const containerClass = $derived(["list-container"].filter(Boolean).join(" "));
</script>

<div class={containerClass}>
  {@render children()}
</div>

<style>
  .list-container {
    height: calc(100vh - 84px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--ha-card-background, #1c1c1e);
  }
</style>
