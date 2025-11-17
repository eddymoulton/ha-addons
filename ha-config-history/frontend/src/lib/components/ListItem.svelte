<script lang="ts">
  import type { Snippet } from "svelte";
  import Card from "./Card.svelte";

  type Props = {
    title: string;
    selected?: boolean;
    variant?: "default" | "current";
    hoverTransform?: "lift" | "slide" | "none";
    onclick?: (event: MouseEvent) => void;
    onkeydown?: (event: KeyboardEvent) => void;
    children?: Snippet;
    actions?: Snippet;
  };

  let {
    title,
    selected = false,
    variant = "default" as "default" | "current",
    hoverTransform = "lift" as "lift" | "slide" | "none",
    onclick = undefined,
    onkeydown = undefined,
    children = undefined,
    actions = undefined,
  }: Props = $props();
</script>

<Card {selected} {variant} {hoverTransform} {onclick} {onkeydown}>
  <div class="list-item-content">
    <div class="list-item-main">
      <div class="list-item-title">{title}</div>
      {@render children?.()}
    </div>
    <div class="list-item-actions">
      {@render actions?.()}
    </div>
  </div>
</Card>

<style>
  .list-item-title {
    color: var(--primary-text-color, #ffffff);
    font-size: 1rem;
    font-weight: 500;
    margin: 0;
    line-height: 1.3;
    flex: 1;
  }
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
