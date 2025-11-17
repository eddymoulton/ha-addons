<script lang="ts">
  type Props = {
    selected?: boolean;
    variant?: "default" | "current";
    clickable?: boolean;
    hoverTransform?: "lift" | "slide" | "none";
    onclick?: (event: MouseEvent) => void;
    onkeydown?: (event: KeyboardEvent) => void;
  };

  let {
    selected = false,
    variant = "default" as "default" | "current",
    clickable = true,
    hoverTransform = "lift" as "lift" | "slide" | "none",
    onclick = undefined,
    onkeydown = undefined,
  }: Props = $props();
</script>

<div
  class="card"
  class:selected
  class:current={variant === "current"}
  class:clickable
  class:hover-lift={hoverTransform === "lift"}
  class:hover-slide={hoverTransform === "slide"}
  {onclick}
  {onkeydown}
  role={clickable ? "button" : undefined}
  tabindex={clickable ? 0 : undefined}
>
  <slot />
</div>

<style>
  .card {
    background: var(--ha-card-background, #1c1c1e);
    border: 1px solid var(--ha-card-border-color, #2c2c2e);
    border-radius: 8px;
    padding: 0.5rem 1rem;
    transition: all 0.2s ease;
    position: relative;
    outline: none;
  }

  .card.clickable {
    cursor: pointer;
  }

  .card.clickable.hover-lift:hover,
  .card.clickable.hover-lift:focus {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    border-color: var(--primary-color, #03a9f4);
  }

  .card.clickable.hover-slide:hover,
  .card.clickable.hover-slide:focus {
    background: var(--ha-card-border-color, #3c3c3e);
    border-color: var(--primary-color, #03a9f4);
    transform: translateX(4px);
  }

  .card.selected {
    border-color: var(--primary-color, #03a9f4);
    background: rgba(3, 169, 244, 0.1);
  }

  .card.current {
    border-color: var(--success-color, #4caf50);
    background: rgba(76, 175, 80, 0.1);
  }

  .card.current.clickable:hover,
  .card.current.clickable:focus {
    background: rgba(76, 175, 80, 0.2);
  }

  .card.current.selected {
    background: rgba(76, 175, 80, 0.15);
  }
</style>
