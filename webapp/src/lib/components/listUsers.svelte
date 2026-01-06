<script lang="ts">
	import type { UserWithRole } from '$lib/types';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import Label from './label.svelte';
	import ListUserEntry from './listUserEntry.svelte';
	import { toast } from '$lib/stores';

	type ListUsersProps = {
		loadingText?: string;
		users: PollingResource<UserWithRole[]>;
	};

	let { loadingText = 'Loading users...', users }: ListUsersProps = $props();

	let filter = $state('');

	const filteredUsers = $derived.by(() => {
		if (filter.trim() === '') {
			return users.value;
		}
		const lowerFilter = filter.trim().toLowerCase();
		return users.value?.filter(
			(u: UserWithRole) =>
				u.name.toLowerCase().includes(lowerFilter) || u.email.toLowerCase().includes(lowerFilter),
		);
	});

	let confirmDisableOpen = $state(false);
	let disableReasonInput = $state('');
	let disabling = $state(false);

	// Which user is currently being confirmed/disabled
	let pendingDisableUser = $state<UserWithRole | null>(null);

	// Track per-user busy state (disable row controls)
	let busyUserIds = $state<Set<string>>(new Set());

	function setBusy(userId: string, busy: boolean) {
		const next = new Set(busyUserIds);
		if (busy) next.add(userId);
		else next.delete(userId);
		busyUserIds = next;
	}

	function isBusy(userId: string): boolean {
		return busyUserIds.has(userId);
	}

	async function enableUser(user: UserWithRole): Promise<void> {
		setBusy(user.user_id, true);
		try {
			const r = await fetch(`/api/admin/users/${user.user_id}/enable`, {
				method: 'POST',
				credentials: 'same-origin',
			});

			if (!r.ok) {
				const detail = await r.text().catch(() => r.statusText);
				toast.error('Failed to enable user.', { detail: detail || r.statusText });
				return;
			}

			await users.fetchNow();
			toast.success('User enabled.');
		} catch (e) {
			toast.error('Failed to enable user.', { detail: String(e) });
		} finally {
			setBusy(user.user_id, false);
		}
	}

	function requestDisable(user: UserWithRole) {
		pendingDisableUser = user;
		disableReasonInput = '';
		confirmDisableOpen = true;

		// mark row busy while modal is open
		setBusy(user.user_id, true);
	}

	function cancelDisable() {
		if (pendingDisableUser) {
			setBusy(pendingDisableUser.user_id, false);
		}
		pendingDisableUser = null;
		confirmDisableOpen = false;
		disableReasonInput = '';
		disabling = false;
	}

	async function doDisableConfirmed(): Promise<void> {
		if (!pendingDisableUser) {
			return;
		}

		disabling = true;
		try {
			const r = await fetch(`/api/admin/users/${pendingDisableUser.user_id}/disable`, {
				method: 'POST',
				credentials: 'same-origin',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ reason: disableReasonInput.trim() }),
			});

			if (!r.ok) {
				const detail = await r.text().catch(() => r.statusText);
				toast.error('Failed to disable user.', { detail: detail || r.statusText });
				return;
			}

			await users.fetchNow();
			toast.success('User disabled.');
			cancelDisable(); // closes + clears busy
		} catch (e) {
			toast.error('Failed to disable user.', { detail: String(e) });
		} finally {
			disabling = false;
			// if still open (error path), keep busy until cancel/confirm succeeds
		}
	}
</script>

<div class="mb-4 md:mb-8">
	<Label target="user-filter" text="Filter Users" />
	<input id="user-filter" type="text" class="input input-sm w-full" bind:value={filter} />
</div>
<div class="card mt-0 bg-base-100 shadow-md">
	<div class="card-body p-0">
		<div class="overflow-x-auto">
			<table class="table w-full table-fixed table-sm">
				<colgroup>
					<col class="w-[30%]" />
					<!-- Task -->
					<col class="w-[40%]" />
					<!-- Info -->
					<col class="w-[20%]" />
					<!-- Assignment -->
				</colgroup>
				<thead>
					<tr>
						<th>User</th>
						<th>Status</th>
						<th>Role</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredUsers as user (user.user_id)}
						<ListUserEntry
							{users}
							{user}
							{enableUser}
							{requestDisable}
							busy={isBusy(user.user_id)}
						/>
					{/each}
				</tbody>
			</table>
			{#if users.value === null}
				<p class="text-center text-gray-500 italic">{loadingText}</p>
			{:else if users.value.length === 0}
				<p class="text-center text-lg text-gray-500 italic">No users found.</p>
			{/if}
		</div>
	</div>
</div>

<dialog class="modal" open={confirmDisableOpen}>
	<div class="modal-box">
		<h3 class="text-lg font-bold">
			Disable User <span class="font-extralight text-accent">{pendingDisableUser?.name}</span>
		</h3>

		<p class="py-2 opacity-80">The user will be disabled and will no longer be able to log in.</p>
		<p class="py-2 opacity-80">Enter a disable reason below.</p>

		<p class="py-2 opacity-80">
			<textarea
				class="textarea-bordered textarea w-full"
				placeholder="Reason for disabling this user..."
				bind:value={disableReasonInput}
				disabled={disabling}
			></textarea>
		</p>

		<div class="modal-action">
			<button type="button" class="btn" onclick={cancelDisable} disabled={disabling}>Cancel</button>
			<button
				type="button"
				class="btn btn-error"
				onclick={doDisableConfirmed}
				disabled={disableReasonInput.trim().length === 0 || disabling}
			>
				{disabling ? 'Disabling...' : 'Disable'}
			</button>
		</div>
	</div>

	<form
		method="dialog"
		class="modal-backdrop"
		onsubmit={(e) => {
			e.preventDefault();
			cancelDisable();
		}}
	>
		<button aria-label="Close"></button>
	</form>
</dialog>
