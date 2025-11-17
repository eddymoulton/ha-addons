<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";

  interface Props extends HTMLButtonAttributes {
    label: string | Snippet;
    variant?:
      | "primary"
      | "secondary"
      | "success"
      | "danger"
      | "outlined"
      | "ghost";
    size?: "small" | "medium" | "large";
    fullWidth?: boolean;
    loading?: boolean;
    icon?: string;
    iconPosition?: "left" | "right";
    class?: string;
  }

  let {
    variant = "primary",
    size = "medium",
    label = "",
    fullWidth = false,
    loading = false,
    icon = undefined,
    iconPosition = "left",
    class: className = "",
    disabled = false,
    onclick = undefined,
    ...restProps
  }: Props = $props();

  const buttonClass = $derived(
    [
      "btn",
      `btn-${variant}`,
      `btn-${size}`,
      fullWidth && "btn-full-width",
      loading && "btn-loading",
      className,
    ]
      .filter(Boolean)
      .join(" ")
  );
</script>

<button
  class={buttonClass}
  disabled={disabled || loading}
  {onclick}
  {...restProps}
>
  {#if loading}
    <span class="btn-spinner" aria-hidden="true">⟳</span>
  {:else if icon && iconPosition === "left"}
    <span class="btn-icon" aria-hidden="true">{icon}</span>
  {/if}
  {#if typeof label === "function"}
    {@render label()}
  {:else}
    {label}
  {/if}
  {#if icon && iconPosition === "right" && !loading}
    <span class="btn-icon" aria-hidden="true">{icon}</span>
  {/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    border: none;
    border-radius: 4px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
    line-height: 1;
    position: relative;
    white-space: nowrap;
  }

  .btn:focus-visible {
    outline: 2px solid var(--primary-color);
    outline-offset: 2px;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-small {
    padding: 0.4rem 0.8rem;
    font-size: 0.85rem;
  }

  .btn-medium {
    padding: 0.6rem 1.2rem;
    font-size: 0.95rem;
  }

  .btn-large {
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    font-weight: 600;
    border-radius: 6px;
  }

  .btn-full-width {
    width: 100%;
  }

  .btn-primary {
    background: var(--primary-color);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--primary-color-dark);
  }

  .btn-secondary {
    background: var(--ha-card-border-color);
    color: var(--primary-text-color);
  }

  .btn-secondary:hover:not(:disabled) {
    background: var(--ha-card-border-color);
  }

  .btn-success {
    background: var(--success-color);
    color: white;
  }

  .btn-success:hover:not(:disabled) {
    background: var(--success-color-dark);
  }

  .btn-large.btn-success:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(76, 175, 80, 0.3);
  }

  .btn-danger {
    background: var(--error-color);
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background: #d32f2f;
  }

  .btn-outlined {
    background: transparent;
    color: var(--primary-color);
    border: 1px solid var(--primary-color);
  }

  .btn-outlined:hover:not(:disabled) {
    background: var(--primary-color);
    color: white;
  }

  .btn-ghost {
    background: transparent;
    color: var(--primary-text-color);
    border: 1px solid transparent;
  }

  .btn-ghost:hover:not(:disabled) {
    background: var(--ha-card-border-color);
    border-color: var(--ha-card-border-color);
  }

  .btn-ghost.btn-danger {
    color: var(--error-color);
  }

  .btn-ghost.btn-danger:hover:not(:disabled) {
    background: rgba(244, 67, 54, 0.1);
    border-color: transparent;
  }

  .btn-loading {
    color: transparent;
  }

  .btn-spinner {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    animation: spin 1s linear infinite;
    color: currentColor;
    font-size: 1.2em;
  }

  .btn-icon {
    display: inline-flex;
    font-size: 1.2em;
    line-height: 1;
  }

  @keyframes spin {
    from {
      transform: translate(-50%, -50%) rotate(0deg);
    }
    to {
      transform: translate(-50%, -50%) rotate(360deg);
    }
  }
</style>
