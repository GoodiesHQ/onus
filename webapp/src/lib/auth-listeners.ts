import { auth } from '$lib/stores/auth.svelte';
import { goto } from '$app/navigation';
import { toast } from './stores';

export function initializeAuthListeners(): () => void {
	// Listen for auth errors from fetch calls
	const originalFetch = window.fetch;
	window.fetch = async (...args) => {
		const response = await originalFetch(...args);

		// Check if this is an auth-related endpoint and if we got unauthorized
		const url = typeof args[0] === 'string' ? args[0] : (args[0]?.toString() ?? '');

		if ((url.includes('/api/') || url.includes('/auth/')) && response.status === 401) {
			// User is now unauthorized
			if (auth.self) {
				// Only log out if we think we were authenticated
				console.warn('User became unauthorized, logging out...');
				toast.error('Your session has expired. Please log in again.');
				auth.logout();
				await goto('/login');
			}
		}

		return response;
	};

	return () => {
		// Restore original fetch on cleanup
		window.fetch = originalFetch;
	};
}
