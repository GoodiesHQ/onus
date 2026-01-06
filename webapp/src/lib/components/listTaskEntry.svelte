<script lang="ts">
	import { toast, users } from '$lib/stores';
	import { TaskListScope, TaskPriorityNames, TaskStatus } from '$lib/types';
	import { TaskPriority, type Task } from '$lib/types';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import Status from './status.svelte';
	import Icon from '@iconify/svelte';

	type ListTaskEntryProps = {
		task: Task;
		tasks: PollingResource<Task[]>;
		scope: TaskListScope;
		openTask?: (task: Task) => void;
	};

	let { tasks = $bindable(), task = $bindable(), scope, openTask }: ListTaskEntryProps = $props();

	let disabled = $state(false);

	const taskDueBy = $derived.by(() => (task.due_by ? new Date(task.due_by) : null));

	let isPastDue = $derived.by(() => {
		if (!taskDueBy) {
			return false;
		}
		const now = new Date();
		return taskDueBy < now && task.status !== TaskStatus.Complete;
	});

	function setChecked(taskId: string, value: boolean) {
		disabled = true;
		const url = `/api/tasks/${taskId}`;
		fetch(url, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				progress: value ? 100 : 0,
			}),
		})
			.then((r) => {
				if (r.ok) {
					for (let t of tasks.value ?? []) {
						if (t.id === task.id) {
							t.status = value ? TaskStatus.Complete : TaskStatus.NotStarted;
							break;
						}
					}
					tasks.fetchNow();
				} else {
					console.error('Failed to update task status: ' + r.statusText);
					toast.error('Failed to update task status', { detail: r.statusText });
				}
			})
			.catch((error) => {
				console.error('Error updating task status:', error);
				toast.error('Failed to update task status', { detail: String(error) });
			})
			.finally(() => {
				disabled = false;
			});
	}

	const priorityColor = $derived.by(() => {
		switch (task.priority) {
			case TaskPriority.Low:
				return 'info';
			case TaskPriority.Medium:
				return 'success';
			case TaskPriority.High:
				return 'warning';
			case TaskPriority.Urgent:
				return 'error';
			default:
				return 'info';
		}
	});
</script>

<tr
	class={`h-14 align-top ${task.status === TaskStatus.Complete ? 'line-through opacity-50' : ''}`}
>
	<!-- Done -->
	<td class="w-12 align-middle">
		<input
			type="checkbox"
			class="checkbox checkbox-sm"
			checked={task.status === TaskStatus.Complete}
			onchange={(e) => {
				setChecked(task.id, (e.target as HTMLInputElement).checked);
			}}
			{disabled}
		/>
		<div class="mt-2 text-xs font-normal text-base-content/75">{task.progress ?? 0}%</div>
	</td>
	<td class="align-middle">
		<div class="min-w-0">
			<div class="truncate mask-ellipse font-bold">{task.title ?? ''}</div>
			<div class="truncate mask-ellipse text-sm text-base-content/65">{task.description ?? ''}</div>
		</div>
	</td>

	<td>
		<div class="text-xs">
			<span class={isPastDue ? 'font-bold text-error' : ''}>
				{#if task.due_by}
					{new Date(task.due_by).toLocaleDateString(undefined, {
						year: 'numeric',
						month: 'short',
						day: 'numeric',
					})}
				{:else}
					-
				{/if}
			</span>
		</div>
		<div class="text-xs">
			<span>
				<Status color={priorityColor} animate={isPastDue} />
			</span>
			<span>{TaskPriorityNames[task.priority]}</span>
		</div>
	</td>

	<td>
		<span class="text-xs font-bold">
			{#if scope === TaskListScope.Assigned}
				{users.getUserById(task.assigner)?.name}
			{:else if scope === TaskListScope.Requested}
				{users.getUserById(task.assignee)?.name}
			{/if}
		</span>
	</td>
	<td class="w-full">
		<button
			class="btn w-full p-0 btn-ghost btn-sm"
			{disabled}
			onclick={() => {
				if (openTask) {
					openTask(task);
				}
			}}
		>
			<Icon icon="mdi:gear" class="h-5 w-5 p-0" />
		</button>
	</td>
</tr>
