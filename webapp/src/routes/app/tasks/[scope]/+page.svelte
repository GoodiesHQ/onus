<script lang="ts">
	import CreateTask from '$lib/components/createTask.svelte';
	import FilterTasks from '$lib/components/filterTasks.svelte';
	import ListTasks from '$lib/components/listTasks.svelte';
	import TaskEditor from '$lib/components/taskEditor.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import type { PageData } from './$types';
	import { page } from '$app/state';
	import { TaskListScope, type Task } from '$lib/types';
	import { getTaskPoller } from '../poller.svelte';
	import { goto } from '$app/navigation';
	import type { Optional } from '$lib/utils';
	import { onMount } from 'svelte';

	const { data } = $props<{ data: PageData }>();
	const tasks = getTaskPoller();
	const scope = $derived(data.scope);
	const heading = $derived(scope === TaskListScope.Assigned ? 'assigned to' : 'assigned by');
	let taskEditing = $state<Optional<Task>>(null);
	let message = $state('');

	const selectedTaskId = $derived(page.url.searchParams.get('task'));
	const drawerOpen = $derived(!!selectedTaskId);

	function setTaskInUrl(task: Optional<Task>) {
		taskEditing = task;
		const url = new URL(page.url);
		if (task !== null) {
			url.searchParams.set('task', task.id);
		} else {
			url.searchParams.delete('task');
		}
		goto(url, { replaceState: false, keepFocus: true, noScroll: true });
	}

	function openTask(task: Task) {
		setTaskInUrl(task);
	}

	function closeTask() {
		setTaskInUrl(null);
		tasks.fetchNow();
	}

	onMount(async () => {
		await tasks.wait();

		// If a task ID is present in the URL, load that task for editing
		if (page.url.searchParams.get('task')) {
			const taskId = page.url.searchParams.get('task')!;
			const task = tasks.value?.find((t) => t.id === taskId) ?? null;
			if (task) {
				taskEditing = task;
			} else {
				// try to fetch the task from the server
				const res = await fetch(`/api/tasks/${taskId}`, { credentials: 'same-origin' });
				if (res.ok) {
					const fetchedTask = (await res.json()) as Task;
					taskEditing = fetchedTask;
				} else {
					if (res.status == 403) {
						message = `You do not have permission to view task with ID ${taskId}.`;
					} else if (res.status == 404) {
						message = `Task with ID ${taskId} not found.`;
					} else {
						message = `Failed to load task with ID ${taskId}: ${res.statusText}`;
					}
				}
			}
		} else {
			message = 'No task selected.';
		}
	});
</script>

<div class="drawer drawer-end">
	<!-- DaisyUI drawer toggle -->
	<input
		id="task-edit-drawer"
		type="checkbox"
		class="drawer-toggle"
		checked={drawerOpen}
		onchange={(e) => {
			const checked = (e.currentTarget as HTMLInputElement).checked;
			if (!checked) {
				closeTask();
			}
		}}
	/>

	<!-- Main page content -->
	<div class="drawer-content space-y-6">
		<CreateTask {scope} {tasks} />

		<div class="mt-2 mb-0">
			<p class="text-xl font-extralight">
				All tasks {heading} <span class="text-accent">{auth.self?.name}</span>:
			</p>

			<FilterTasks {scope} {tasks} />
		</div>

		<ListTasks {scope} {tasks} {openTask} />
	</div>

	<!-- Right-hand task editor panel -->
	<div class="drawer-side z-50">
		<!-- clicking the overlay closes drawer -->
		<button class="drawer-overlay" onclick={closeTask} title="Close task editor"></button>

		<aside class="flex h-full w-[92vw] max-w-5xl flex-col border-l border-base-300 bg-base-100 shadow-2xl">
			<div class="flex items-center justify-between border-b border-base-300 p-4">
				<div class="text-lg font-semibold">Edit Task</div>
				<button class="btn btn-ghost btn-sm" type="button" onclick={closeTask}>Close</button>
			</div>

			<div class="flex-1 overflow-y-auto p-4">
				{#if taskEditing !== null}
					<TaskEditor task={taskEditing} onClose={closeTask} />
				{:else}
					<div class="opacity-60">{message}</div>
				{/if}
			</div>
		</aside>
	</div>
</div>
