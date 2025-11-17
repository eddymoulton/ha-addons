<script lang="ts">
  import type { Snippet } from "svelte";

  /**
   * ListContent - A reusable scrollable content area for lists
   * Provides consistent styling for list content with overflow
   */
  type Props = { children: Snippet };
  let { children }: Props = $props();

  const contentClass = $derived(["list-content"].filter(Boolean).join(" "));
</script>

<div class={contentClass}>
  {@render children()}
</div>

<style>
  .list-content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
  }
</style>
