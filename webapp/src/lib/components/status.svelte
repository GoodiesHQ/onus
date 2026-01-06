<script lang="ts">
	type StatusProps = {
		animate?: boolean;
		color: 'info' | 'success' | 'warning' | 'error';
	};

	let { animate = false, color = 'info' }: StatusProps = $props();

	let statusClass = $derived('status-' + color);
</script>

<div class="inline-grid *:[grid-area:1/1]">
	{#if animate}
		<div class="status {statusClass} animate-ping"></div>
	{/if}
	<div class="status {statusClass}"></div>
</div>
