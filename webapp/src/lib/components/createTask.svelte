<script lang="ts">
	import { auth, toast, users } from '$lib/stores';
	import { TaskListScope, TaskPriority, type Task, type User } from '$lib/types';
	import { dateToISO, YMDtoLocalDate } from '$lib/utils';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import { onMount, tick } from 'svelte';
	import UserSelect from './userSelect.svelte';
	import Title from './title.svelte';

	type CreateTaskProps = {
		tasks: PollingResource<Task[]>;
		scope: TaskListScope;
	};

	let { tasks, scope }: CreateTaskProps = $props();

	// Task information
	let title = $state('');
	let assignee_id = $state(auth.self?.user_id ?? ''); // Default to self
	let assignee_name = $state(auth.self?.name ?? ''); // For display in the input
	let description = $state(''); // Optional longer text
	let priority = $state<TaskPriority>(2); // Default to Medium
	let due_by = $state(''); // YYYY-MM-DD convention
	let user_search_query = $state('');

	// Query and UI state for user selection
	let show_advanced = $state(false);

	let disable_create = $derived(!title.trim() || !assignee_id);

	function focusInputTaskTitle() {
		tick().then(() => {
			document.getElementById('task-title')?.focus();
		});
	}

	async function handleSubmit() {
		if (!title.trim() || !assignee_id) {
			return;
		}

		try {
			const res = await fetch('/api/tasks/new', {
				method: 'POST',
				credentials: 'same-origin',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					title,
					assignee_id,
					description,
					priority: priority.valueOf(),
					due_by: due_by ? dateToISO(YMDtoLocalDate(due_by)) : null,
				}),
			});

			if (res.ok) {
				// Reset form
				title = '';
				assignee_id = auth.self?.user_id ?? '';
				assignee_name = auth.self?.name ?? '';
				user_search_query = assignee_name;
				description = '';
				priority = 2;
				due_by = '';
				show_advanced = false;
				tasks.fetchNow(); // Refresh task list
				toast.success('Task created successfully');
			} else {
			}
		} catch (error) {
			console.error('Failed to create task:', error);
			toast.error('Failed to create task.', { detail: String(error) });
		} finally {
			focusInputTaskTitle();
		}
	}

	onMount(async () => {
		await Promise.all([users.startPolling(), auth.wait()]);

		assignee_id = auth.self?.user_id ?? '';
		assignee_name = auth.self?.name ?? '';

		focusInputTaskTitle();
	});
</script>

<Title text="Create a Task" />

<!-- New Task Form -->
<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-3"
>
	<!-- Title and Assignee Row -->
	<div class="flex items-center gap-3">
		<!-- Title Input -->
		<input
			id="task-title"
			type="text"
			bind:value={title}
			placeholder="Task title..."
			class="input-bordered input flex-1"
			required
		/>

		<UserSelect
			bind:user_id={assignee_id}
			bind:user_name={assignee_name}
			bind:query={user_search_query}
			{scope}
			containerClass="flex-[0_1_25%] w-full max-w-64"
			required
			defaultSelf={true}
		/>

		<!-- Optional Submit Button -->
		<button type="submit" disabled={disable_create} class="btn btn-sm btn-primary">Create</button>
	</div>

	<!-- Advanced Options Collapse -->
	<div class="collapse-arrow collapse bg-base-200">
		<input type="checkbox" bind:checked={show_advanced} />
		<div class="collapse-title text-sm font-medium">Advanced Options</div>
		<div class="collapse-content space-y-3">
			<!-- Description -->
			<textarea
				id="task-description"
				bind:value={description}
				placeholder="Detailed description (optional)..."
				class="textarea-bordered textarea h-20 w-full"
			></textarea>

			<!-- Priority and Date Row -->
			<div class="flex items-center gap-10">
				<div class="flex w-80 items-center gap-3">
					Priority:
					<select id="task-priority" bind:value={priority} class="select-bordered select flex-1">
						<option value={TaskPriority.Low}>Low</option>
						<option value={TaskPriority.Medium}>Medium </option>
						<option value={TaskPriority.High}>High</option>
						<option value={TaskPriority.Urgent}>Urgent</option>
					</select>
				</div>

				<div class="flex w-80 items-center gap-3">
					Due Date:
					<input
						id="task-date"
						type="date"
						bind:value={due_by}
						class="input-bordered input flex-1"
						placeholder="Due date"
					/>
				</div>
			</div>
		</div>
	</div>
</form>
