<script lang="ts">
  type Props = { class?: string };
  let { class: className = "" }: Props = $props();

  const sectionClass = $derived(
    ["filter-section", className].filter(Boolean).join(" ")
  );
</script>

<div class={sectionClass}>
  <slot />
</div>

<style>
  .filter-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--ha-card-background, #1c1c1e);
    min-height: 140px;
  }

  @media (max-width: 768px) {
    .filter-section {
      flex-direction: column;
      align-items: stretch;
      gap: 0.5rem;
    }
  }
</style>
