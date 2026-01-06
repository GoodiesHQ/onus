import type { Optional } from './index';

type PollHandle = ReturnType<typeof setInterval>;

export type PollingResource<T> = {
	value: Optional<T>;
	loading: boolean;
	error: Optional<string>;
	lastFetch: Optional<number>;

	fetchNow(): Promise<Optional<T>>;
	startPolling(interval?: number): Promise<Optional<T>>;
	stopPolling(): void;
	wait(timeout?: number): Promise<boolean>;
	updateUrl(url: string): void;
};

type CreatePollingResourceOptions<T> = {
	interval?: number;
	credentials?: RequestCredentials;
	unauthorizedValue?: null;
};

export function createPollingResource<T>(
	url: string,
	opts: CreatePollingResourceOptions<T> = {},
): PollingResource<T> {
	const credentials = opts.credentials ?? 'same-origin';
	const defaultInterval = opts.interval ?? 30000;

	let pollInterval: Optional<PollHandle> = null;

	let value = $state<Optional<T>>(null);
	let loading = $state(true);
	let error = $state<Optional<string>>(null);
	let lastFetch = $state<Optional<number>>(null);
	let urlCurrent = $state<string>(url);

	let fetchSeq = 0;
	let abort: Optional<AbortController> = null;

	async function fetchNow(): Promise<Optional<T>> {
		loading = true;

		const seq = ++fetchSeq;
		abort?.abort();
		abort = new AbortController();

		try {
			const res = await fetch(urlCurrent, { credentials, signal: abort.signal });

			if (seq !== fetchSeq) {
				return value;
			}

			if (!res.ok) {
				if (res.status === 401 || res.status === 403) {
					value = null;
					loading = false;
					error = null;
					lastFetch = Date.now();
					return null;
				}
				throw new Error(`Failed to fetch resource: ${res.status} ${res.statusText}`);
			}
			const parsed = (await res.json()) as T;
			value = parsed;
			loading = false;
			error = null;
			lastFetch = Date.now();
			console.log('Fetched polling resource:', urlCurrent);
			return parsed;
		} catch (err) {
			if (err instanceof DOMException && err.name === 'AbortError') {
				return value;
			}

			const message = err instanceof Error ? err.message : 'Unknown error';
			loading = false;
			error = message;
			lastFetch = Date.now();
			return null;
		}
	}

	function updateUrl(newUrl: string) {
		urlCurrent = newUrl;
	}

	function startPolling(interval: number = defaultInterval): Promise<Optional<T>> {
		const val = fetchNow();

		if (pollInterval) {
			clearInterval(pollInterval);
		}
		pollInterval = setInterval(() => {
			fetchNow();
		}, interval);

		return val;
	}

	function stopPolling() {
		if (pollInterval) {
			clearInterval(pollInterval);
			pollInterval = null;
		}
		abort?.abort();
		abort = null;
	}

	async function wait(timeoutMs = 5000) {
		if (lastFetch) {
			return true;
		}

		return new Promise<boolean>((resolve) => {
			const start = Date.now();
			const interval = setInterval(() => {
				if (lastFetch) {
					clearInterval(interval);
					resolve(true);
				} else if (Date.now() - start > timeoutMs) {
					clearInterval(interval);
					resolve(false);
				}
			}, 50);
		});
	}

	return {
		get value(): Optional<T> {
			return value;
		},
		set value(v: Optional<T>) {
			value = v;
		},

		get loading(): boolean {
			return loading;
		},
		set loading(v: boolean) {
			loading = v;
		},

		get error(): Optional<string> {
			return error;
		},
		set error(v: Optional<string>) {
			error = v;
		},

		get lastFetch(): Optional<number> {
			return lastFetch;
		},
		set lastFetch(v: Optional<number>) {
			lastFetch = v;
		},

		updateUrl,
		fetchNow,
		startPolling,
		stopPolling,
		wait,
	};
}
