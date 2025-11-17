<script lang="ts">
  import type { HTMLButtonAttributes } from "svelte/elements";
  import Button from "./Button.svelte";

  interface Props extends HTMLButtonAttributes {
    icon: string;
    variant?:
      | "primary"
      | "secondary"
      | "success"
      | "danger"
      | "outlined"
      | "ghost";
    size?: "small" | "medium" | "large";
    loading?: boolean;
    class?: string;
    label?: string;
  }

  let {
    icon,
    variant = "ghost",
    size = "medium",
    loading = false,
    class: className = "",
    onclick = undefined,
    onmouseenter = undefined,
    onmouseleave = undefined,
    onfocus = undefined,
    onblur = undefined,
    ...restProps
  }: Props = $props();

  const iconButtonClass = $derived(
    ["icon-btn", `icon-btn-${size}`, className].filter(Boolean).join(" ")
  );
</script>

{#snippet label()}
  {#if !loading}
    <span class="icon-btn-content" aria-hidden="true">{icon}</span>
  {/if}
{/snippet}

<Button
  {label}
  class={iconButtonClass}
  {variant}
  {loading}
  {onclick}
  {onmouseenter}
  {onmouseleave}
  {onfocus}
  {onblur}
  aria-label={restProps["aria-label"]}
  {...restProps}
></Button>

<style>
  :global(.icon-btn) {
    padding: 0.25rem !important;
    aspect-ratio: 1;
    min-width: unset;
  }

  :global(.icon-btn-small) {
    width: 1.75rem;
    height: 1.75rem;
    font-size: 1rem;
  }

  :global(.icon-btn-medium) {
    width: 2rem;
    height: 2rem;
    font-size: 1.1rem;
  }

  :global(.icon-btn-large) {
    width: 2.5rem;
    height: 2.5rem;
    font-size: 1.3rem;
  }

  :global(.icon-btn.btn-ghost:hover:not(:disabled)) {
    transform: scale(1.1);
  }

  :global(.icon-btn.btn-ghost.btn-danger:hover:not(:disabled)) {
    opacity: 1;
    background: rgba(244, 67, 54, 0.1);
  }

  :global(.icon-btn.btn-ghost.btn-danger) {
    opacity: 0.7;
  }

  .icon-btn-content {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }
</style>
