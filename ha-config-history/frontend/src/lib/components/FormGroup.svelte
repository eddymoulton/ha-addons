<script lang="ts">
  interface FormGroupProps {
    label: string;
    for: string;
    weight?: "light" | "heavy";
    helpText?: string;
    children: any;
  }

  let {
    label,
    for: forId,
    weight = "heavy",
    helpText,
    children,
  }: FormGroupProps = $props();
</script>

<div class="form-group form-group-{weight}">
  {#if label}
    <label for={forId}>
      {label}
      {#if helpText}
        <span class="help-text">{helpText}</span>
      {/if}
    </label>
  {/if}
  {@render children?.()}
</div>

<style>
  .form-group-heavy {
    margin-bottom: 1rem;
  }

  .form-group.form-group-heavy label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--primary-text-color);
    font-size: 0.9rem;
    font-weight: 500;
  }

  .form-group.form-group-light label {
    display: block;
    color: var(--secondary-text-color);
    font-size: 0.7rem;
    font-weight: 400;
  }

  .help-text {
    color: var(--secondary-text-color);
    font-size: 0.8rem;
    font-weight: 400;
  }
</style>
