<script lang="ts">
	import Label from '$lib/components/label.svelte';
	import Title from '$lib/components/title.svelte';
	import { auth, toast } from '$lib/stores';
	import { onMount } from 'svelte';

	let name = $state('');
	let email = $state('');
	let organization = $state('');

	let saving = $state(false);
	let saveEnabled = $derived(!saving && name.trim().length > 0 && name !== auth.self?.name);

	onMount(async () => {
		await auth.wait();
		if (auth.isAuthenticated) {
			name = auth.self?.name || '';
			email = auth.self?.email || '';
			organization = auth.self?.organization_name || '';
		}
	});

	async function saveProfile(e: Event) {
		e.preventDefault();
		saving = true;
		try {
			const res = await fetch('/api/me', {
				method: 'PATCH',
				credentials: 'same-origin',
				body: JSON.stringify({
					name: name.trim(),
				}),
			});
			if (!res.ok) {
				console.log('Failed to save profile:', res.status, await res.text());
				toast.error('Failed to save profile changes.', { detail: `${res.statusText}` });
			} else {
				const updatedUser = await res.json();
				if (auth.self) {
					auth.self.name = updatedUser.name;
				}
				toast.success('Profile updated successfully.');
			}
		} catch (err) {
			console.error('Error saving profile:', err);
			toast.error('Failed to save profile changes.', { detail: String(err) });
		} finally {
			saving = false;
		}
	}
</script>

<Title text="User Profile" />
<form class="grid grid-cols-2 items-center gap-4 lg:grid-cols-3" onsubmit={saveProfile}>
	<div class="col-span-full lg:col-span-1">
		<Label target="name" text="Display Name" />
		<input id="name" type="text" class="input-bordered input w-full" bind:value={name} />
	</div>
	<div class="col-span-full lg:col-span-1">
		<Label target="email" text="Email Address" />
		<div class="input m-0 w-full input-ghost p-0">{email}</div>
	</div>
	<div class="col-span-full lg:col-span-1">
		<Label target="organization" text="Organization" />
		<input type="text" class="input w-full input-ghost p-0" value={organization} readonly />
	</div>
	<div class="col-span-full mt-8 flex justify-around">
		<button type="submit" class="btn btn-primary" disabled={!saveEnabled}>Save Changes</button>
	</div>
</form>
