<script lang="ts">
	import { toast } from '$lib/stores';
	import { AuthRole, type UserWithRole } from '$lib/types';
	import type { PollingResource } from '$lib/utils/poller.svelte';

	type ListUserEntryProps = {
		user: UserWithRole;
		users: PollingResource<UserWithRole[]>;
		enableUser: (user: UserWithRole) => Promise<void>;
		requestDisable: (user: UserWithRole) => void;
		busy: boolean;
	};

	let { user, users = $bindable(), enableUser, requestDisable, busy }: ListUserEntryProps = $props();

	// svelte-ignore state_referenced_locally
	let role = $state(user.role);

	async function setRole() {
		// you can also route role-change through parent if you want
		toast.error('Failed to update user role.', { detail: 'Not implemented yet' });
		role = user.role;
	}
</script>

<tr class="h-14 align-top">
	<td>
		<div class="truncate mask-ellipse font-bold">{user.name}</div>
		<div class="truncate mask-ellipse text-sm text-base-content/65">{user.email}</div>
	</td>

	<td class="flex flex-row items-center justify-between">
		<div class="text-xs">
			{#if user.disabled_at === null}
				<span class="font-bold text-success">Active</span>
			{:else}
				<span class="font-bold text-warning">Disabled {new Date(user.disabled_at).toLocaleString()}</span>
			{/if}
			<div class="text-base-content/75" class:invisible={!user.disabled_at}>
				{user.disabled_reason ?? ''}
			</div>
		</div>

		<div>
			{#if user.disabled_at === null}
				<button class="btn w-16 btn-xs btn-error" disabled={busy} onclick={() => requestDisable(user)}>
					Disable
				</button>
			{:else}
				<button class="btn w-16 btn-xs btn-success" disabled={busy} onclick={() => enableUser(user)}> Enable </button>
			{/if}
		</div>
	</td>

	<td>
		<select class="select w-32 select-sm" bind:value={role} onchange={setRole} disabled={busy}>
			<option value={AuthRole.Member}>Member</option>
			<option value={AuthRole.Admin}>Admin</option>
			<option value={AuthRole.Owner}>Owner</option>
		</select>
	</td>
</tr>
