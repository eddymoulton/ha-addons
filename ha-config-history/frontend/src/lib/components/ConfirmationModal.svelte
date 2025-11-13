<script lang="ts">
  import type { Snippet } from "svelte";
  import Modal from "../Modal.svelte";
  import Button from "./Button.svelte";

  type Props = {
    isOpen: boolean;
    title: string;
    message: string;
    onClose: () => void;
    onConfirm: () => void;
    confirmText?: string;
    cancelText?: string;
    variant?: 'primary' | 'danger' | 'warning';
    disabled?: boolean;
    children?: Snippet;
  };

  let { 
    isOpen, 
    title, 
    message, 
    onClose, 
    onConfirm, 
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    variant = 'primary',
    disabled = false,
    children
  }: Props = $props();
</script>

<Modal {isOpen} {title} onClose={onClose} size="small">
  <p class="message">{message}</p>
  
  {#if children}
    {@render children()}
  {/if}

  {#snippet actions()}
    <Button
      variant="secondary"
      onclick={onClose}
      type="button"
      {disabled}
    >
      {cancelText}
    </Button>
    <Button
      {variant}
      onclick={onConfirm}
      type="button"
      {disabled}
    >
      {confirmText}
    </Button>
  {/snippet}
</Modal>

<style>
  .message {
    margin: 0 0 1rem 0;
    color: var(--secondary-text-color, #9b9b9b);
    line-height: 1.5;
  }
</style>