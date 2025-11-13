<script lang="ts">
  import type { Snippet } from "svelte";
  import IconButton from "./components/IconButton.svelte";

  type Props = {
    isOpen: boolean;
    title: string;
    onClose: () => void;
    children: Snippet;
    actions?: Snippet;
    size?: 'small' | 'medium' | 'large';
  };

  let { isOpen, title, onClose, children, actions, size = 'large' }: Props = $props();

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      onClose();
    }
  }

  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }
</script>

{#if isOpen}
  <div
    class="modal-backdrop"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
    tabindex="0"
    role="dialog"
    aria-modal="true"
    aria-labelledby="modal-title"
  >
    <div class="modal-content" class:small={size === 'small'} class:medium={size === 'medium'} class:large={size === 'large'}>
      <div class="modal-header">
        <h2 id="modal-title">{title}</h2>
        <IconButton
          icon="×"
          variant="ghost"
          size="large"
          onclick={onClose}
          type="button"
          aria-label="Close modal"
        />
      </div>

      <div class="modal-body">
        {@render children()}
      </div>

      {#if actions}
        <div class="modal-footer">
          {@render actions()}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-content {
    background: var(--ha-card-background, #1c1c1e);
    border: 1px solid var(--ha-card-border-color, #2c2c2e);
    border-radius: 8px;
    width: 100%;
    max-width: 800px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 1.5rem 1rem 1.5rem;
    border-bottom: 1px solid var(--ha-card-border-color, #2c2c2e);
    flex-shrink: 0;
  }

  .modal-header h2 {
    color: var(--primary-text-color, #ffffff);
    font-size: 1.5rem;
    font-weight: 500;
    margin: 0;
  }

  .modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
  }

  .modal-footer {
    padding: 1rem 1.5rem 1.5rem 1.5rem;
    border-top: 1px solid var(--ha-card-border-color, #2c2c2e);
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    flex-shrink: 0;
  }

  .modal-content.small {
    max-width: 450px;
  }

  .modal-content.medium {
    max-width: 600px;
  }

  .modal-content.large {
    max-width: 800px;
  }

  @media (max-width: 768px) {
    .modal-content {
      max-width: 100%;
      margin: 0;
      height: 100%;
      max-height: 100%;
      border-radius: 0;
    }

    .modal-backdrop {
      padding: 0;
    }
  }
</style>
