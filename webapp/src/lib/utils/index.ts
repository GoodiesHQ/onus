export type Optional<T> = T | null;

export function choose<T>(items: T[]): T {
	if (items.length === 0) {
		throw new RangeError('Cannot choose from an empty array');
	}

	const index = Math.floor(Math.random() * items.length);
	return items[index];
}

export function YMDtoLocalDate(ymd: string): Date {
	const [year, month, day] = ymd.split('-').map(Number);
	return new Date(year, month - 1, day, 0, 0, 0, 0);
}

export function dateToISO(date: Date | string): string {
	if (typeof date === 'string') {
		date = new Date(date);
	}
	return date.toISOString();
}

export function today(): string {
	const now = new Date();
	return dateToYMD(now);
}

export function dateToYMD(date: Date | string): string {
	if (typeof date === 'string') {
		date = new Date(date);
	}
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');
	return `${year}-${month}-${day}`;
}
