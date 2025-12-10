<script lang="ts">
  interface FormSelectProps {
    id?: string;
    value: string;
    children?: any;
    style?: "group-select" | null;
    onchange?: (event: Event) => void;
  }

  let {
    id,
    value = $bindable(),
    children,
    style = null,
    onchange,
    ...restProps
  }: FormSelectProps = $props();
</script>

<select
  {id}
  bind:value
  {onchange}
  class="form-select {style === 'group-select' ? 'group-select' : ''}"
  {...restProps}
>
  {@render children?.()}
</select>

<style>
  .form-select {
    width: 100%;
    padding: 0.6rem;
    background: var(--ha-card-background);
    border: 1px solid var(--ha-card-border-color);
    border-radius: 4px;
    color: var(--primary-text-color);
    font-size: 0.9rem;
  }

  .form-select:focus {
    outline: none;
    border-color: var(--primary-color);
  }

  /* Group select specific styles */
  .form-select.group-select {
    flex: 1;
    padding: 0.5rem;
    background: var(--ha-card-background);
    border: 1px solid var(--ha-card-border-color);
    border-radius: 4px;
    color: var(--primary-text-color);
    font-size: 0.9rem;
  }

  .form-select.group-select:focus {
    outline: none;
    border-color: var(--primary-color);
  }
</style>
