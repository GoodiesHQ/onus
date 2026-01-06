import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const ssr = false;

export const load: PageLoad = ({ params }) => {
	switch (params.scope) {
		case 'assigned':
		case 'requested':
			return { scope: params.scope };
		default:
			throw error(404, 'Unknown scope');
	}
};
