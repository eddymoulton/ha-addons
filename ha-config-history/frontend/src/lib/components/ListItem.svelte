<script lang="ts">
  import Card from "./Card.svelte";

  type Props = {
    selected?: boolean;
    variant?: "default" | "current";
    hoverTransform?: "lift" | "slide" | "none";
    onclick?: (event: MouseEvent) => void;
    onkeydown?: (event: KeyboardEvent) => void;
  };

  let {
    selected = false,
    variant = "default" as "default" | "current",
    hoverTransform = "lift" as "lift" | "slide" | "none",
    onclick = undefined,
    onkeydown = undefined,
  }: Props = $props();
</script>

<Card {selected} {variant} {hoverTransform} {onclick} {onkeydown}>
  <div class="list-item-content">
    <div class="list-item-main">
      <slot name="title" />
      <slot name="content" />
    </div>
    <div class="list-item-actions">
      <slot name="actions" />
    </div>
  </div>
</Card>

<style>
  .list-item-content {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    min-height: 2rem;
  }

  .list-item-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .list-item-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  @media (max-width: 768px) {
    .list-item-content {
      flex-direction: column;
      gap: 0.5rem;
    }

    .list-item-actions {
      align-self: flex-end;
    }
  }
</style>
