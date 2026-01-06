import { AuthRole, type AuthSnapshot } from '$lib/types';
import { createPollingResource } from '$lib/utils/poller.svelte';

class AuthState {
	private resource = createPollingResource<AuthSnapshot>('/api/me');

	loading = $derived(this.resource.loading);
	error = $derived(this.resource.error);
	lastFetch = $derived(this.resource.lastFetch);

	self = $derived(this.resource.value);
	isAuthenticated = $derived(this.self !== null);
	isAdmin = $derived(this.self ? this.self.role >= AuthRole.Admin : false);
	userRole = $derived(this.self?.role ?? null);
	organizationName = $derived(this.self?.organization_name ?? null);
	hasAdminAccess = $derived(this.self ? this.self.role >= 2 : false);

	fetchState() {
		return this.resource.fetchNow();
	}

	startPolling(interval: number = 30000) {
		return this.resource.startPolling(interval);
	}

	stopPolling() {
		this.resource.stopPolling();
	}

	logout() {
		this.stopPolling();
		this.resource.value = null;
		this.resource.loading = false;
		this.resource.error = null;
		this.resource.lastFetch = null;
	}

	wait(timeoutMs = 5000) {
		return this.resource.wait(timeoutMs);
	}
}

export const auth = new AuthState();
