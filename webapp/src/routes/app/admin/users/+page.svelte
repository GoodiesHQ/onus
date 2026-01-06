<script lang="ts">
	import ListUsers from '$lib/components/listUsers.svelte';
	import Title from '$lib/components/title.svelte';
	import type { UserWithRole } from '$lib/types';
	import { createPollingResource } from '$lib/utils/poller.svelte';
	import { onMount } from 'svelte';

	let users = createPollingResource<UserWithRole[]>('/api/admin/users', {
		credentials: 'same-origin',
		interval: 30000,
	});

	onMount(() => {
		users.startPolling();

		return () => {
			users.stopPolling();
		};
	});
</script>

<Title text="User Management" />

<ListUsers {users} />
