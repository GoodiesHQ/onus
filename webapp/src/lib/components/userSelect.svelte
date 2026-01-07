<script lang="ts">
	import { auth, users } from '$lib/stores';
	import { TaskStatus, type TaskListScope, type User } from '$lib/types';
	import { onMount, untrack } from 'svelte';

	type CreateTaskProps = {
		user_id: string;
		user_name: string;
		query?: string;
		required?: boolean;
		small?: boolean;
		containerClass: string;
		disabled?: boolean;
		readonly?: boolean;
		scope?: TaskListScope;
		defaultSelf?: boolean;
	};

	let {
		user_id = $bindable(),
		user_name = $bindable(),
		query = $bindable(''),
		required = false,
		small = false,
		containerClass,
		disabled = false,
		readonly = false,
		defaultSelf = false,
		scope = undefined,
	}: CreateTaskProps = $props();

	const provided_user_id = $state(user_id);
	const provided_user_name = $state(user_name);

	if (user_id && user_name) {
		query = user_name;
	}

	let show_user_dropdown = $state(false);

	let filtered_users = $derived(
		users.all?.filter(
			(user) =>
				user.name.toLowerCase().includes(query.toLowerCase()) || user.email.toLowerCase().includes(query.toLowerCase()),
		) ?? [],
	);

	function selectUser(user: User) {
		user_id = user.id;
		user_name = user.name;
		query = user.name;
		show_user_dropdown = false;
	}

	function handleUserInputFocus() {
		if (readonly || disabled) {
			return;
		}
		show_user_dropdown = true;
		query = '';
	}

	function handleUserInputBlur() {
		if (readonly || disabled) {
			return;
		}
		// Delay to allow click on dropdown item
		setTimeout(() => {
			show_user_dropdown = false;
			if (!user_id) {
				query = '';
				user_id = '';
				user_name = '';
			} else {
				query = user_name;
			}
		}, 200);
	}

	$effect(() => {
		scope;
		untrack(() => {
			if (defaultSelf) {
				user_id = auth.self?.user_id ?? '';
				user_name = auth.self?.name ?? '';
				query = auth.self?.name ?? '';
			} else {
				user_id = provided_user_id;
				user_name = provided_user_name;
				query = provided_user_name;
			}
		});
	});

	onMount(async () => {
		if (defaultSelf) {
			await auth.wait();
			query = auth.self?.name ?? '';
			user_id = auth.self?.user_id ?? '';
			user_name = auth.self?.name ?? '';
		}
	});
</script>

<!-- Assignee Selection -->
<div class="relative {containerClass}">
	<input
		id="task-assignee"
		type="text"
		bind:value={query}
		onfocus={handleUserInputFocus}
		onblur={handleUserInputBlur}
		placeholder=""
		class="input {small ? 'input-sm' : ''} input-bordered w-full"
		autocomplete="off"
		{disabled}
		{readonly}
		required
	/>

	{#if !readonly && show_user_dropdown && filtered_users.length > 0}
		<ul class="menu absolute top-full z-10 mt-1 max-h-60 w-full overflow-y-auto rounded-box bg-base-100 shadow-lg">
			{#if required === false}
				<li>
					<button
						type="button"
						onclick={() => {
							user_id = '';
							user_name = '';
							query = '';
						}}
						class="flex flex-col items-start"
					>
						<span class="font-semibold">-</span>
					</button>
				</li>
			{/if}
			{#each filtered_users as user}
				<li>
					<button {disabled} type="button" onclick={() => selectUser(user)} class="flex flex-col items-start">
						<span class="font-semibold">{user.name}</span>
						<span class="text-xs text-base-content/70">{user.email}</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
