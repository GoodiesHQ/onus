import type { User } from '$lib/types';
import { createPollingResource } from '$lib/utils/poller.svelte';

class UsersState {
	private resource = createPollingResource<User[]>('/api/users');
	all = $derived(this.resource.value?.sort((a, b) => a.name.localeCompare(b.name)));
	loading = $derived(this.resource.loading);
	error = $derived(this.resource.error);
	lastFetch = $derived(this.resource.lastFetch);

	count = $derived(this.all?.length ?? 0);

	getUserById(id: string) {
		return this.all?.find((u) => u.id === id) ?? null;
	}

	fetchUsers() {
		return this.resource.fetchNow();
	}

	startPolling(interval: number = 60000) {
		return this.resource.startPolling(interval);
	}

	stopPolling() {
		this.resource.stopPolling();
	}

	wait(timeoutMs = 5000) {
		return this.resource.wait(timeoutMs);
	}
}

export const users = new UsersState();
