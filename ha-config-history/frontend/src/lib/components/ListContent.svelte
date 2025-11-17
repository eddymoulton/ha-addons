<script lang="ts">
  /**
   * ListContent - A reusable scrollable content area for lists
   * Provides consistent styling for list content with overflow
   */
  type Props = { class?: string; padding?: string };
  let { class: className = "", padding = "1rem" }: Props = $props();

  const contentClass = $derived(
    ["list-content", className].filter(Boolean).join(" ")
  );
</script>

<div class={contentClass} style:padding>
  <slot />
</div>

<style>
  .list-content {
    flex: 1;
    overflow-y: auto;
  }
</style>
