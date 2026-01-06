<script lang="ts">
	import { auth } from '$lib/stores';
	import { initializeAuthListeners } from '$lib/auth-listeners';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onDestroy, onMount } from 'svelte';
	import '../app.css';
	import favicon from '$lib/assets/favicon.ico';
	import ToastHost from '$lib/components/toastHost.svelte';

	let { children, data } = $props();
	let unsubscribeListeners: (() => void) | null = null;

	onMount(async () => {
		// Start polling for user updates every 30 seconds
		auth.self = data.self;
		await auth.startPolling(30000);

		// Initialize auth event listeners for 401/403 responses
		unsubscribeListeners = initializeAuthListeners();

		// Handle redirects based on auth state
		const isLoginOrAuth =
			$page.url.pathname.startsWith('/login') ||
			$page.url.pathname.startsWith('/auth') ||
			$page.url.pathname === '/';

		if (auth.isAuthenticated && isLoginOrAuth) {
			// If authenticated but on login or home page, redirect to the app
			await goto('/app/tasks/assigned');
		} else if (!auth.isAuthenticated && !isLoginOrAuth) {
			// If not authenticated and not on login page, redirect to login
			await goto('/login');
		}
	});

	// Cleanup on component destroy
	onDestroy(() => {
		auth.stopPolling();
		if (unsubscribeListeners) {
			unsubscribeListeners();
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<ToastHost />
