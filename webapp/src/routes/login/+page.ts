import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	try {
		const res = await fetch('/auth');

		if (!res.ok) {
			throw new Error(`Failed to fetch auth providers: ${res.statusText}`);
		}

		const authProviders: string[] = await res.json();

		return {
			authProviders,
		};
	} catch (error) {
		console.error('Error loading auth providers:', error);
		return {
			authProviders: [],
		};
	}
};
