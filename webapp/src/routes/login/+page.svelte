<script lang="ts">
	import type { PageData } from './$types';
	import logo from '$lib/assets/onus_logo.png';

	let { data }: { data: PageData } = $props();
</script>

<div class="mb-4">
	<img src={logo} alt="Onus Logo" class="mx-auto mt-6 w-28" />
	<h1 class="mt-2 justify-center text-center text-3xl">
		<span class="font-bold">Onus </span>
		<span class="font-extralight">| Task Manager</span>
	</h1>
</div>
<div class="card bg-base-100 py-0 shadow-xl">
	<div class="card-body">
		<h2 class="mb-8 card-title justify-center text-center text-xl">Sign In</h2>

		{#if data.authProviders && data.authProviders.length > 0}
			<div class="flex flex-col gap-3">
				{#each data.authProviders as provider}
					<a href="/auth/{provider}/login" class="btn uppercase btn-lg btn-primary">
						Log in with {provider}
					</a>
				{/each}
			</div>
		{:else}
			<div class="alert alert-warning">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-6 w-6 shrink-0 stroke-current"
					fill="none"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
					/>
				</svg>
				<span>No authentication providers available</span>
			</div>
		{/if}
	</div>
</div>
