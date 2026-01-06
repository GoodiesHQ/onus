import type { Task } from '$lib/types';
import { createPollingResource, type PollingResource } from '$lib/utils/poller.svelte';

let singleton: PollingResource<Task[]> | null = null;

export function getTaskPoller(): PollingResource<Task[]> {
	if (!singleton) {
		singleton = createPollingResource<Task[]>('/api/tasks');
	}
	return singleton;
}
