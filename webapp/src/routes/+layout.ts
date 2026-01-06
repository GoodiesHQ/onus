// This is a client-side only application
// Authentication checks and redirects are handled in +layout.svelte
// No server-side load function needed for static deployment

import { redirect } from '@sveltejs/kit';
import type { Load } from '@sveltejs/kit';

export const ssr = false;
export const prerender = false;

export const load: Load = async ({ fetch, url }) => {
	// Allow unauthenticated access to login and registration pages
	if (url.pathname === '/login' || url.pathname === '/register') {
		return {};
	}

	// Fetch current user info
	const res = await fetch('/api/me', { credentials: 'include' });

	if (!res.ok) {
		throw redirect(302, '/login?redirect=' + encodeURIComponent(url.pathname));
	}

	const self = await res.json();

	return {
		self,
	};
};
