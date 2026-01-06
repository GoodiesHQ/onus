<script lang="ts">
	import { TaskListScope, type Task } from '$lib/types';
	import { choose } from '$lib/utils';
	import type { PollingResource } from '$lib/utils/poller.svelte';
	import ListTaskEntry from './listTaskEntry.svelte';

	type ListTasksProps = {
		tasks: PollingResource<Task[]>;
		scope: TaskListScope;
		openTask: (task: Task) => void;
	};

	let { tasks, scope, openTask }: ListTasksProps = $props();

	const loadingText = '';

	// Fun motivational messages for when there are no tasks
	export const emptyTaskMessages = [
		'All caught up!',
		'Nothing on your plate.',
		'Zero tasks. Nice.',
		'You’re officially done.',
		'Mission accomplished.',
		'Zen achieved.',
		'Enjoy the calm.',
		'Free as a bird.',
		'No tasks. No stress.',
		'Time to breathe.',
		'Now go touch grass.',
		'Take a victory lap.',
		'You’ve earned a nap.',
		'No pending side quests.',
		'That’s a wrap.',
		'All quiet here.',
		'Silence is golden.',
		'Productivity: complete.',
		'Everything’s handled.',
		'Nothing left to do.',
		'You’re unstoppable.',
		'Go be mysterious.',
		'Time for snacks.',
		'Coffee break approved.',
		'Nap time authorized.',
		'Go forth and relax.',
		'Enjoy your freedom.',
		'Not a task in sight.',
		'You did all the things.',
		'Blank slate achieved.',
		'Nothing pressing.',
		'Schedule: clear.',
		'Your work here is done.',
		'World peace achieved.',
	];
</script>

<div class="card mt-0 bg-base-100 shadow-md">
	<div class="card-body p-0">
		<div class="overflow-x-auto">
			<table class="table w-full table-fixed table-sm">
				<colgroup>
					<col class="w-12" />
					<col class="w-[60%]" />
					<!-- Task -->
					<col class="w-[18%]" />
					<!-- Info -->
					<col class="w-[22%]" />
					<!-- Assignment -->
					<col class="w-14" />
					<!-- Edit -->
				</colgroup>
				<thead>
					<tr>
						<th class="w-12 text-center">✓</th>
						<th>Task</th>
						<th>Due/Priority</th>
						<th>Assigned {scope === TaskListScope.Assigned ? 'By' : 'To'}</th>
						<th>Edit</th>
					</tr>
				</thead>
				<tbody>
					{#each tasks.value ?? [] as task (task.id)}
						<ListTaskEntry {tasks} {scope} {task} {openTask} />
					{/each}
				</tbody>
			</table>
			{#if tasks.value === null}
				<p class="text-center text-gray-500 italic">{loadingText}</p>
			{:else if tasks.value.length === 0}
				<p class="text-center text-lg text-gray-500 italic">{choose(emptyTaskMessages)}</p>
			{/if}
		</div>
	</div>
</div>
