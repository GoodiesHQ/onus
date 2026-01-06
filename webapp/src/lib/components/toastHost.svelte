<script lang="ts">
	import { toast, type ToastKind } from '$lib/stores';

	function kindClass(kind: ToastKind) {
		switch (kind) {
			case 'success':
				return 'alert-success';
			case 'error':
				return 'alert-error';
			case 'info':
				return 'alert-info';
		}
		return 'alert-info';
	}
</script>

<div class="toast toast-center toast-bottom z-9999">
	{#each toast.items as t (t.id)}
		<div class={'alert max-w-md shadow-lg ' + kindClass(t.kind)}>
			<div class="flex flex-col gap-1">
				<div class="font-semibold">{t.message}</div>
				{#if t.detail}
					<div class="text-xs wrap-break-word opacity-80">{t.detail}</div>
				{/if}
			</div>

			<button
				type="button"
				class="btn btn-ghost btn-xs"
				aria-label="Dismiss toast"
				onclick={() => toast.dismiss(t.id)}
			>
				✕
			</button>
		</div>
	{/each}
</div>
