<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { getTaskPoller } from './poller.svelte';

	let { children } = $props();

	const tasks = getTaskPoller();

	// start polling once; let FilterTasks change URL
	onMount(async () => {
		await tasks.startPolling();
		await tasks.wait();
	});

	onDestroy(() => tasks.stopPolling());
</script>

{@render children?.()}
