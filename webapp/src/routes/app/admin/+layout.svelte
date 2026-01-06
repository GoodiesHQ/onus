<script lang="ts">
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(async () => {
		// If user doesn't have admin access, redirect to app
		await auth.wait();
		if (!auth.hasAdminAccess) {
			goto('/app');
		}
	});
</script>

{#if auth.hasAdminAccess}
	{@render children()}
{/if}
