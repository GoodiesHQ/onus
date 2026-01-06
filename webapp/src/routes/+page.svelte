<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		goto('/login');
	});
</script>

<div class="flex min-h-screen items-center justify-center">
	<div class="loading loading-lg loading-spinner"></div>
</div>
