export type ToastKind = 'info' | 'success' | 'error';

export type ToastItem = {
	id: string;
	kind: ToastKind;
	message: string;
	detail?: string;
	durationMs: number;
	createdAt: Date;
};

function generateId(): string {
	return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

type ToastAPI = {
	items: ToastItem[];
	push: (
		kind: ToastKind,
		message: string,
		opts?: { detail?: string; durationMs?: number },
	) => string;
	dismiss: (id: string) => void;
	clear: () => void;

	success: (message: string, opts?: { detail?: string; durationMs?: number }) => string;
	error: (message: string, opts?: { detail?: string; durationMs?: number }) => string;
	info: (message: string, opts?: { detail?: string; durationMs?: number }) => string;
};

export function createToastStore(): ToastAPI {
	let items = $state<ToastItem[]>([]);

	function push(
		kind: ToastKind,
		message: string,
		opts?: { detail?: string; durationMs?: number },
	): string {
		const id = generateId();
		const durationMs = opts?.durationMs ?? (kind === 'error' ? 10_000 : 5_000);
		items.push({
			id,
			kind,
			message,
			detail: opts?.detail,
			durationMs,
			createdAt: new Date(),
		});

		window.setTimeout(() => {
			dismiss(id);
		}, durationMs);

		return id;
	}

	function dismiss(id: string) {
		items = items.filter((item) => item.id !== id);
	}

	function clear() {
		items = [];
	}

	return {
		get items() {
			return items;
		},
		push,
		dismiss,
		clear,
		success(message: string, opts?: { detail?: string; durationMs?: number }) {
			return push('success', message, opts);
		},
		error(message: string, opts?: { detail?: string; durationMs?: number }) {
			return push('error', message, opts);
		},
		info(message: string, opts?: { detail?: string; durationMs?: number }) {
			return push('info', message, opts);
		},
	};
}

export const toast = createToastStore();
