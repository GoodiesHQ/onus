<script lang="ts">
	import Label from '$lib/components/label.svelte';
	import Title from '$lib/components/title.svelte';
	import UserSelect from '$lib/components/userSelect.svelte';
	import { auth, toast } from '$lib/stores';
	import { AuthRole, TaskListScope, type Organization } from '$lib/types';
	import { onMount } from 'svelte';

	let name = $state('');
	let domain = $state('');

	let saving = $state(false);
	let saveEnabled = $derived(!saving && name.trim().length > 0 && name !== auth.self?.organization_name);

	let confirmTransferOpen = $state(false);
	let confirmTransferChecked = $state(false);
	let transferring = $state(false);
	let newOwnerID = $state<string>('');
	let newOwnerName = $state<string>('');

	onMount(async () => {
		await auth.wait();
		if (auth.isAuthenticated) {
			name = auth.self?.organization_name || '';
			domain = auth.self?.organization_domain || '';
		}
		if (auth.userRole === AuthRole.Owner) {
			newOwnerID = auth.self?.user_id || '';
			newOwnerName = auth.self?.name || '';
		}
	});

	function doTransfer() {
		transferring = true;
		fetch('/api/owner/transfer', {
			method: 'POST',
			credentials: 'same-origin',
			body: JSON.stringify({
				new_owner_user_id: newOwnerID,
			}),
		})
			.then(async (res) => {
				if (!res.ok) {
					const detail = await res.text().catch(() => res.statusText);
					toast.error('Failed to transfer organization ownership.', { detail: detail || res.statusText });
					return;
				}
				toast.success('Organization ownership transferred successfully.');
				// Update auth store
				await auth.fetchState();
				// Reset transfer form
				newOwnerID = '';
				newOwnerName = '';
				confirmTransferChecked = false;
				confirmTransferOpen = false;
			})
			.catch((err) => {
				console.error('Error transferring organization ownership:', err);
				toast.error('Failed to transfer organization ownership.', { detail: String(err) });
			})
			.finally(() => {
				transferring = false;
			});
	}

	function cancelTransfer() {
		confirmTransferOpen = false;
		confirmTransferChecked = false;
	}

	async function saveSettings(e: Event) {
		e.preventDefault();
		saving = true;
		try {
			const res = await fetch('/api/admin/organization', {
				method: 'PATCH',
				credentials: 'same-origin',
				body: JSON.stringify({
					name: name.trim(),
				}),
			});
			if (!res.ok) {
				console.log('Failed to save organization settings:', res.status, await res.text());
				toast.error('Failed to save organization settings.', { detail: `${res.statusText}` });
			} else {
				const updatedOrg = (await res.json()) as Organization;
				if (auth.self) {
					auth.self.organization_name = updatedOrg.name;
				}
				toast.success('Organization settings updated successfully.');
			}
		} catch (err) {
			console.error('Error saving organization settings:', err);
			toast.error('Failed to save organization settings.', { detail: String(err) });
		} finally {
			saving = false;
		}
	}
</script>

<Title text="Organization Settings" />
<form class="grid grid-cols-2 items-center gap-4" onsubmit={saveSettings}>
	<div class="col-span-full lg:col-span-1">
		<Label target="name" text="Display Name" />
		<input id="name" type="text" class="input-bordered input w-full" bind:value={name} />
	</div>
	<div class="col-span-full lg:col-span-1">
		<Label target="email" text="Domain" />
		<div class="input m-0 w-full input-ghost p-0">@{domain}</div>
	</div>
	<div class="col-span-full mt-8 flex justify-around">
		<button type="submit" class="btn btn-primary" disabled={!saveEnabled}>Save Changes</button>
	</div>
</form>

{#if auth.userRole === AuthRole.Owner}
	<div class="mt-12">
		<Label target="transfer-form" text="Transfer Organization Ownership" />
		<form id="transfer-form" class="flex w-full flex-row items-center gap-4">
			<UserSelect
				bind:user_id={newOwnerID}
				bind:user_name={newOwnerName}
				defaultSelf={true}
				required={true}
				containerClass="w-full max-w-lg min-w-xs"
			/>
			<button
				type="button"
				class="btn w-28 text-xs btn-outline btn-warning"
				disabled={newOwnerID === auth.self?.user_id}
				onclick={() => (confirmTransferOpen = true)}
			>
				Transfer
			</button>
		</form>
	</div>
{/if}

<dialog class="modal" open={confirmTransferOpen}>
	<div class="modal-box">
		<h3 class="text-lg font-bold">Confirm deletion</h3>
		<p class="py-2 opacity-80">
			This will permanently transfer ownership of <span class="text-accent">{auth.self?.organization_name}</span> to
			<span class="text-accent">{newOwnerName}</span>. This action cannot be undone.
		</p>

		<label class="label cursor-pointer justify-start gap-1 md:gap-3">
			<input type="checkbox" class="checkbox checkbox-sm" bind:checked={confirmTransferChecked} />
			<span class="label-text">I understand.</span>
		</label>

		<div class="modal-action">
			<button type="button" class="btn" onclick={cancelTransfer}>Cancel</button>
			<button
				type="button"
				class="btn btn-error"
				disabled={!confirmTransferChecked || transferring}
				onclick={doTransfer}
			>
				Transfer
			</button>
		</div>
	</div>

	<!-- click outside to close -->
	<form
		method="dialog"
		class="modal-backdrop"
		onsubmit={(e) => {
			e.preventDefault();
			cancelTransfer();
		}}
	>
		<button aria-label="Close"></button>
	</form>
</dialog>
