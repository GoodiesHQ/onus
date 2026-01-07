<script lang="ts">
	import { auth, toast, users } from '$lib/stores';
	import { AuthRole, TaskPriority, type Task } from '$lib/types';
	import { dateToISO, dateToYMD, YMDtoLocalDate } from '$lib/utils';
	import Icon from '@iconify/svelte';
	import UserSelect from './userSelect.svelte';
	import Label from './label.svelte';

	type TaskEditorProps = {
		task: Task;
		onClose?: () => void;
	};

	let { task, onClose }: TaskEditorProps = $props();

	// Intentionally using only the initial task values to populate state

	// svelte-ignore state_referenced_locally
	let title = $state(task.title ?? '');
	const title_valid = $derived(title.trim() !== '');

	// svelte-ignore state_referenced_locally
	let description = $state(task.description ?? '');
	// svelte-ignore state_referenced_locally
	let progress = $state(Math.floor(task.progress / 10) * 10);

	// svelte-ignore state_referenced_locally
	let due_by = $state(task.due_by ? dateToYMD(new Date(task.due_by)) : '');
	let clear_due_by = $derived(due_by === '');
	let due_by_valid = $derived.by(() => {
		console.log('Validating due date:', due_by);
		if (due_by === '') {
			return true;
		}

		try {
			const [y, m, d] = due_by.split('-').map(Number);
			if (isNaN(y) || isNaN(m) || isNaN(d)) {
				return false;
			}
			return true;
		} catch (_) {
			return false;
		}
	});

	// svelte-ignore state_referenced_locally
	let assignee_id = $state(task.assignee ?? '');
	// svelte-ignore state_referenced_locally
	let assignee_name = $state(users.getUserById(assignee_id)?.name ?? 'Unknown User');

	// svelte-ignore state_referenced_locally
	let assigner_id = $state(task.assigner ?? '');
	// svelte-ignore state_referenced_locally
	let assigner_name = $state(users.getUserById(assigner_id)?.name ?? 'Unknown User');

	// svelte-ignore state_referenced_locally
	let priority = $state<TaskPriority>(task.priority ?? TaskPriority.Medium);

	const anyInvalid = $derived(!due_by_valid || !title_valid);

	const isAssigner = $derived(auth.self?.user_id === task.assigner);
	const isAssignee = $derived(auth.self?.user_id === task.assignee);
	const isAssignerOrAssignee = $derived(isAssigner || isAssignee);

	let confirmDeleteOpen = $state(false);
	let confirmDeleteChecked = $state(false);
	let deleting = $state(false);

	function save() {
		try {
			const url = `/api/tasks/${task.id}`;
			const body: any = {
				title: title.trim(),
				description: description.trim(),
				priority: priority.valueOf(),
				progress: progress,
				assignee_user_id: assignee_id,
				clear_due_by: clear_due_by,
				due_by: due_by ? dateToISO(YMDtoLocalDate(due_by)) : null,
			};

			fetch(url, {
				method: 'PATCH',
				headers: {
					'Content-Type': 'application/json',
				},
				credentials: 'same-origin',
				body: JSON.stringify(body),
			}).then((r) => {
				if (r.ok) {
					console.log('Task saved successfully');
					if (onClose) {
						onClose();
					}
				} else {
					console.error('Error saving task:', r.statusText);
					toast.error('Error saving task', { detail: r.statusText });
				}
			});
		} catch (error) {
			console.error('Error saving task:', error);
			toast.error('Error saving task', { detail: String(error) });
		}
	}

	function requestDelete() {
		confirmDeleteOpen = true;
		confirmDeleteChecked = false;
	}

	function cancelDelete() {
		confirmDeleteOpen = false;
		confirmDeleteChecked = false;
		deleting = false;
	}

	function doDelete() {
		if (!confirmDeleteChecked || deleting) {
			return;
		}

		deleting = true;
		fetch(`/api/tasks/${task.id}`, {
			method: 'DELETE',
			credentials: 'same-origin',
		})
			.then((r) => {
				if (r.ok) {
					if (onClose) onClose();
				} else {
					console.error('Error deleting task:', r.statusText);
					toast.error('Error deleting task', { detail: r.statusText });
				}
			})
			.catch((e) => {
				console.error('Error deleting task:', e);
				toast.error('Error deleting task', { detail: String(e) });
			})
			.finally(() => {
				deleting = false;
			});
	}
</script>

<form class="grid w-full grid-cols-2 gap-2 md:grid-cols-3 md:gap-4">
	<!-- Title Input -->
	<div class="col-span-full">
		<!-- Edit the title -->
		<Label target="task-title" text="Title" />
		<input
			id="task-title"
			type="text"
			bind:value={title}
			placeholder="Task title..."
			class="input-bordered input w-full flex-1"
			readonly={!isAssigner}
			disabled={!isAssigner}
			required
		/>
	</div>
	<div class="col-span-full">
		<!-- Edit the description -->
		<Label target="task-description" text="Description" />
		<textarea
			id="task-description"
			bind:value={description}
			placeholder="Detailed description (optional)..."
			class="textarea-bordered textarea h-20 w-full"
			readonly={!isAssigner}
			disabled={!isAssigner}
		>
		</textarea>
	</div>
	<div class="col-span-2 md:col-span-1">
		<!-- Edit the progress -->
		<div class="w-full">
			<Label target="task-progress" text={`Progress (${progress}%)`} />
			<input
				type="range"
				min="0"
				max="100"
				bind:value={progress}
				class="range w-full rounded-lg range-accent range-xs"
				step="10"
				disabled={!isAssignerOrAssignee}
			/>
			<div class="mt-2 flex w-full justify-between px-2.5 text-xs">
				{#each Array(11) as _, i (i)}
					<span>|</span>
				{/each}
			</div>
			<div class="mt-2 flex w-full justify-between px-2.5 text-xs">
				{#each Array(11) as _, i (i)}
					{#if i % 5 === 0}
						<span>{i * 10}</span>
					{/if}
				{/each}
			</div>
		</div>
	</div>
	<div class="col-span-1">
		<Label target="task-due-date" text="Due Date" />
		<input
			id="task-due-date"
			type="date"
			bind:value={due_by}
			class="input-bordered input flex-1"
			placeholder="Due date"
			disabled={!isAssigner}
		/>
	</div>
	<div class="col-span-1">
		<Label target="task-priority" text="Priority" />
		<select id="task-priority" bind:value={priority} class="select-bordered select flex-1" disabled={!isAssigner}>
			<option value={TaskPriority.Low}>Low</option>
			<option value={TaskPriority.Medium}>Medium </option>
			<option value={TaskPriority.High}>High</option>
			<option value={TaskPriority.Urgent}>Urgent</option>
		</select>
	</div>
	<div class="col-span-full grid grid-cols-2 gap-4">
		<div class="col-span-2 md:col-span-1">
			<Label target="task-assignee" text="Assigned To" />
			<UserSelect
				bind:user_id={assignee_id}
				bind:user_name={assignee_name}
				containerClass="w-full"
				required
				disabled={!isAssigner}
				defaultSelf={false}
			/>
		</div>
		<div class="col-span-2 md:col-span-1">
			<Label target="task-assigner" text="Assigned By" />
			<UserSelect
				bind:user_id={assigner_id}
				bind:user_name={assigner_name}
				containerClass="w-full"
				defaultSelf={false}
				required
				disabled
				readonly
			/>
		</div>
	</div>
	<div class="col-span-full mt-20 flex items-center justify-between">
		<div class:invisible={!isAssigner && !auth.isAdmin}>
			<button type="button" class="btn btn-sm btn-error" onclick={requestDelete} aria-label="Delete task">
				<Icon icon="mdi:delete" class="inline-block h-5 w-5 p-0" />
			</button>
		</div>
		<div class="flex items-center gap-3">
			<button type="button" class="btn w-32 btn-ghost btn-sm" onclick={onClose}> Cancel </button>

			<button type="button" class="btn w-32 btn-sm btn-primary" disabled={anyInvalid} onclick={save}> Save </button>
		</div>
	</div>
</form>

<dialog class="modal" open={confirmDeleteOpen}>
	<div class="modal-box">
		<h3 class="text-lg font-bold">Confirm deletion</h3>
		<p class="py-2 opacity-80">
			This will permanently delete the task from the Onus database. This action cannot be undone.
		</p>

		<label class="label cursor-pointer justify-start gap-1 md:gap-3">
			<input type="checkbox" class="checkbox checkbox-sm" bind:checked={confirmDeleteChecked} />
			<span class="label-text">I understand. Delete this task.</span>
		</label>

		<div class="modal-action">
			<button type="button" class="btn" onclick={cancelDelete}>Cancel</button>
			<button type="button" class="btn btn-error" disabled={!confirmDeleteChecked || deleting} onclick={doDelete}>
				{deleting ? 'Deleting...' : 'Delete'}
			</button>
		</div>
	</div>

	<!-- click outside to close -->
	<form
		method="dialog"
		class="modal-backdrop"
		onsubmit={(e) => {
			e.preventDefault();
			cancelDelete();
		}}
	>
		<button aria-label="Close"></button>
	</form>
</dialog>
