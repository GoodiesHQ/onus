<script lang="ts">
	import { auth, toast } from '$lib/stores';
	import { AuthRole, type UserWithRole } from '$lib/types';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import Page from '../../routes/+page.svelte';

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
	let roleDisabled = $state(false);

	async function setRole() {
		roleDisabled = true;
		// you can also route role-change through parent if you want
		try {
			const res = await fetch(`/api/admin/users/${user.user_id}/role`, {
				method: 'PATCH',
				credentials: 'same-origin',
				body: JSON.stringify({ role }),
			});
			if (!res.ok) {
				const detail = await res.text().catch(() => res.statusText);
				toast.error('Failed to update user role.', { detail: detail || res.statusText });
				// revert select
				role = user.role;
				return;
			}
			// Update user in list
			for (let u of users.value ?? []) {
				if (u.user_id === user.user_id) {
					u.role = role;
					break;
				}
			}
		} catch (e) {
			toast.error('Failed to update user role.', { detail: String(e) });
			// revert select
			role = user.role;
		} finally {
			roleDisabled = false;
		}
	}
</script>

<tr class="h-14 align-top">
	<td>
		<div class="truncate mask-ellipse font-bold">{user.name}</div>
		<div class="truncate mask-ellipse text-sm text-base-content/65">{user.email}</div>
	</td>

	<td class="align-middle">
		<div class="flex items-center justify-between gap-2">
			<div class="text-xs">
				<div class="h-4">
					{#if user.disabled_at === null}
						<span class="font-bold text-success">Active</span>
					{:else}
						<span class="font-bold text-warning">Disabled {new Date(user.disabled_at).toLocaleString()}</span>
					{/if}
				</div>
				<div class="h-3 text-base-content/75">
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
		</div>
	</td>

	<td>
		{#if user.role === AuthRole.Owner}
			<span class="">Owner</span>
		{:else}
			<select class="select w-32 select-sm" bind:value={role} onchange={setRole} disabled={busy || roleDisabled}>
				<option value={AuthRole.Member}>Member</option>
				<option value={AuthRole.Admin}>Admin</option>
			</select>
		{/if}
	</td>
</tr>
