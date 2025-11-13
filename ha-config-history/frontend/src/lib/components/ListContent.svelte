<script lang="ts">
	/**
	 * ListContent - A reusable scrollable content area for lists
	 * Provides consistent styling for list content with overflow
	 */
	interface $$Props {
		class?: string;
		padding?: string;
	}

	export let padding: $$Props['padding'] = '1rem';
	let className: $$Props['class'] = '';
	export { className as class };

	$: contentClass = ['list-content', className].filter(Boolean).join(' ');
</script>

<div class={contentClass} style:padding>
	<slot />
</div>

<style>
	.list-content {
		flex: 1;
		overflow-y: auto;
	}
</style>
