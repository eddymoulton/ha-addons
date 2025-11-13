<script lang="ts">
  import Card from './Card.svelte';

  export let selected = false;
  export let variant: 'default' | 'current' = 'default';
  export let hoverTransform: 'lift' | 'slide' | 'none' = 'lift';
</script>

<Card
  {selected}
  {variant}
  {hoverTransform}
  on:click
  on:keydown
>
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