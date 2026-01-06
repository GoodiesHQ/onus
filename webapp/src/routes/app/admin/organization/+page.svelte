<script lang="ts">
	import Label from '$lib/components/label.svelte';
	import Title from '$lib/components/title.svelte';
	import { auth, toast } from '$lib/stores';
	import type { Organization } from '$lib/types';
	import { onMount } from 'svelte';

	let name = $state('');
	let domain = $state('');

	let saving = $state(false);
	let saveEnabled = $derived(
		!saving && name.trim().length > 0 && name !== auth.self?.organization_name,
	);

	onMount(async () => {
		await auth.wait();
		if (auth.isAuthenticated) {
			name = auth.self?.organization_name || '';
			domain = auth.self?.organization_domain || '';
		}
	});

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
