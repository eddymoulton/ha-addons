<script lang="ts">
	/**
	 * ListContainer - A reusable container component for list layouts
	 * Provides consistent structure with header, filter, and content areas
	 */
	let { class: className = '' }: { class?: string } = $props();

	const containerClass = $derived(['list-container', className].filter(Boolean).join(' '));
</script>

<div class={containerClass}>
	<slot name="header" />
	<slot name="filter" />
	<slot name="content" />
</div>

<style>
	.list-container {
		height: calc(100vh - 84px);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		background: var(--ha-card-background, #1c1c1e);
	}
</style>
