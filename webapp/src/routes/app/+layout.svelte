<script lang="ts">
	import { auth, users } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import Icon from '@iconify/svelte';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);

	function isActive(href: string) {
		return currentPath === href || currentPath.startsWith(href + '/');
	}

	type NavItem = {
		label: string;
		href: string;
		icon?: string;
	};

	const navbarMembers: NavItem[] = [
		{ label: 'My Tasks', icon: 'mdi:calendar-task-outline', href: '/app/tasks/assigned' },
		{ label: 'My Requests', icon: 'mdi:subtasks', href: '/app/tasks/requested' },
		{ label: 'Profile', icon: 'mdi:account-settings-outline', href: '/app/profile' },
	];

	const navbarAdmin: NavItem[] = [
		{
			label: 'Organization Settings',
			icon: 'mdi:view-dashboard-outline',
			href: '/app/admin/organization',
		},
		{ label: 'User Management', icon: 'mdi:account-multiple-outline', href: '/app/admin/users' },
	];

	const navbar = $derived([...navbarMembers, ...(auth.hasAdminAccess ? navbarAdmin : [])]);

	onMount(async () => {
		users.startPolling(60000); // Poll list of organization users every 60 seconds

		// Wait for auth to finish initial fetch, this avoids false redirects on page reload
		await auth.wait();
		// If user is not authenticated, redirect to login
		if (!auth.isAuthenticated) {
			goto('/login');
		}
	});

	onDestroy(() => {
		users.stopPolling();
	});
</script>

<div class="drawer lg:drawer-open">
	<input id="app-drawer" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content flex flex-col">
		<!-- Navbar for mobile -->
		<div class="navbar bg-base-300 lg:hidden">
			<div class="flex-none">
				<label for="app-drawer" class="btn btn-square btn-ghost">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						class="inline-block h-6 w-6 stroke-current"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 12h16M4 18h16"
						></path>
					</svg>
				</label>
			</div>
			<div class="flex-1">
				<span class="text-xl font-bold">Onus</span>
			</div>
		</div>

		<!-- Page content -->
		<div class="p-4 lg:p-8">
			{@render children()}
		</div>
	</div>

	<div class="drawer-side">
		<label for="app-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
		<div class="menu flex min-h-full w-80 flex-col bg-base-200 p-4 text-base-content">
			<!-- Sidebar header -->
			<div class="border-b border-base-300 pb-4">
				<h1 class="px-4 pt-2 text-2xl tracking-widest uppercase font-stretch-125%">Onus</h1>
				<div class="card my-8 bg-base-300 shadow-xs">
					<div class="my-4 px-4 text-lg">
						<div class="text-md font-extralight text-base-content/75" class:invisible={!auth.self}>
							{auth.organizationName || '_'}
						</div>
					</div>
				</div>
			</div>

			<!-- Navigation items -->
			<ul class="flex-1">
				{#each navbar as nav (nav.href)}
					<li>
						<a
							href={nav.href}
							class="flex items-center gap-2 rounded-md py-2 transition-colors hover:bg-base-100"
							class:bg-base-300={isActive(nav.href)}
							class:font-semibold={isActive(nav.href)}
							aria-current={isActive(nav.href) ? 'page' : undefined}
						>
							{#if nav.icon}
								<Icon icon={nav.icon} class="h-5 w-5" />
							{/if}
							<span>{nav.label}</span>
						</a>
					</li>
				{/each}
			</ul>

			<!-- Bottom actions -->
			<div class="card my-8 bg-base-300 shadow-xs">
				<div class="my-4 px-4 text-lg">
					<div class="text-sm font-extralight text-base-content" class:invisible={!auth.self}>
						{auth.self?.name || '_'}
					</div>
					<div class="text-xs font-extralight text-base-content/70" class:invisible={!auth.self}>
						{auth.self?.email || '_'}
					</div>
				</div>
			</div>
			<ul class="border-t border-base-300">
				<li><a href="/auth/logout" class="text-warning">Logout</a></li>
			</ul>
		</div>
	</div>
</div>
