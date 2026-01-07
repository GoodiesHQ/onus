<script lang="ts">
	import { TaskListScope, type Task } from '$lib/types';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import { onMount, untrack } from 'svelte';
	import UserSelect from './userSelect.svelte';
	import Label from './label.svelte';

	type ListTaskFiltersProps = {
		disabled?: boolean;
		scope: TaskListScope;
		tasks: PollingResource<Task[]>;
	};

	let { disabled = $bindable(false), scope, tasks }: ListTaskFiltersProps = $props();

	function formatDateInput(date: Date) {
		return date.toISOString().slice(0, 10);
	}
	const defaultStartDate = formatDateInput(new Date(Date.now() - 14 * 24 * 60 * 60 * 1000));
	let filterStartDateISO = $derived.by(() => {
		return new Date(filterStartDate).toISOString();
	});

	// Filter state
	let filterStartDate = $state(defaultStartDate);
	let filterShowComplete = $state(false);
	let filterShowPastDue = $state(false);
	let filterPriorityMin = $state(1);
	let filterAssignmentId = $state('');
	let filterAssignmentName = $state('');

	function buildUrl() {
		let url = '/api/tasks?scope=' + scope;
		if (filterStartDate) {
			url += `&since=${filterStartDateISO}`;
		}
		if (filterAssignmentId) {
			if (scope === TaskListScope.Assigned) {
				url += `&assigner_id=${filterAssignmentId}`;
			}
			if (scope === TaskListScope.Requested) {
				url += `&assignee_id=${filterAssignmentId}`;
			}
		}
		url += `&include_complete=${filterShowComplete}`;
		url += `&past_due=${filterShowPastDue}`;
		url += `&priority_min=${filterPriorityMin}`;
		return url;
	}

	async function onChange() {
		disabled = true;
		try {
			tasks.stopPolling();
			tasks.updateUrl(buildUrl());
			tasks.startPolling();
			await tasks.wait();
		} finally {
			disabled = false;
		}
	}

	// Effects to trigger onChange when filters change
	$effect(() => {
		if (filterAssignmentId !== '') {
			untrack(() => onChange());
		}
	});

	// Reset assignment filter on scope change
	$effect(() => {
		if (scope) {
			untrack(() => {
				filterAssignmentId = '';
				filterAssignmentName = '';
				onChange();
			});
		}
	});

	onMount(async () => {
		onChange();
	});
</script>

<!-- Filters -->
<div class="card mt-4 rounded-md bg-base-200 p-4">
	<div class="grid grid-cols-1 gap-4 md:grid-cols-6">
		<!-- Date -->
		<div class="form-control w-full items-center">
			<Label target="start-date" text="Show tasks since" />
			<input
				id="start-date"
				type="date"
				class="input-bordered input input-sm w-full"
				{disabled}
				onchange={onChange}
				bind:value={filterStartDate}
			/>
		</div>

		<div class="form-control w-full items-center md:col-span-2">
			<div class="gap2 grid grid-cols-2">
				<div class="h-full w-full items-center">
					<Label target="show-complete" text="Show Complete" />
					<input
						id="show-complete"
						type="checkbox"
						class="toggle toggle-sm"
						class:toggle-success={filterShowComplete}
						onchange={onChange}
						{disabled}
						bind:checked={filterShowComplete}
					/>
				</div>
				<div>
					<Label target="past-due-only" text="Past Due Only" />
					<input
						id="past-due-only"
						type="checkbox"
						class="toggle toggle-sm"
						class:toggle-success={filterShowPastDue}
						onchange={onChange}
						{disabled}
						bind:checked={filterShowPastDue}
					/>
				</div>
			</div>
		</div>

		<!-- Priority -->
		<div class="form-control w-full items-center">
			<Label target="min-priority" text="Minimum Priority" />
			<select
				id="min-priority"
				class="select-bordered select w-full select-sm"
				bind:value={filterPriorityMin}
				onchange={onChange}
				{disabled}
			>
				<option value={1}>Low</option>
				<option value={2}>Medium</option>
				<option value={3}>High</option>
				<option value={4}>Urgent</option>
			</select>
		</div>
		<div class="form-control w-full items-center md:col-span-2">
			<Label target="assigned-by" text={`Assigned ${scope === TaskListScope.Assigned ? 'By' : 'To'}`} />
			<UserSelect
				bind:user_id={filterAssignmentId}
				bind:user_name={filterAssignmentName}
				{scope}
				{disabled}
				required={false}
				containerClass="w-full"
				small
			/>
		</div>
	</div>
</div>
