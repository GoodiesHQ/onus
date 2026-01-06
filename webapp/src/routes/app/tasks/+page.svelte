<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		// Redirect to default task view
		goto('/app/tasks/assigned');
	});
</script>
