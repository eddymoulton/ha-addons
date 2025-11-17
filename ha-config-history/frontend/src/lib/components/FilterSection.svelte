<script lang="ts">
  import type { Snippet } from "svelte";

  type Props = { children: Snippet };
  let { children }: Props = $props();

  const sectionClass = $derived(["filter-section"].filter(Boolean).join(" "));
</script>

<div class={sectionClass}>
  {@render children()}
</div>

<style>
  .filter-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--ha-card-background);
    min-height: 140px;
  }

  @media (max-width: 768px) {
    .filter-section {
      gap: 0.5rem;
    }
  }
</style>
