import type { Optional } from './utils';

// User information
export interface User {
	id: string;
	email: string;
	name: string;
	// created_at: string;
}

export interface UserWithRole {
	user_id: string;
	organization_id: string;
	name: string;
	email: string;
	role: AuthRole;
	disabled_at: Optional<string>;
	disabled_reason: Optional<string>;
}

export enum AuthRole {
	Member = 1,
	Admin = 2,
	Owner = 3,
}

export const AuthRoleNames: Record<AuthRole, string> = {
	[AuthRole.Member]: 'Member',
	[AuthRole.Admin]: 'Admin',
	[AuthRole.Owner]: 'Owner',
};

// Authentication snapshot
export interface AuthSnapshot {
	user_id: string;
	email: string;
	name: string;
	organization_id: string;
	organization_name: string;
	organization_domain: string;
	role: AuthRole;
	disabled: boolean;
}

// Organization information
export interface Organization {
	id: string;
	name: string;
	domain: string;
}

// Task status types
export enum TaskStatus {
	NotStarted = 0,
	InProgress = 1,
	Complete = 2,
}

// Human-readable task status names
export const TaskStatusNames: Record<TaskStatus, string> = {
	[TaskStatus.NotStarted]: 'Not Started',
	[TaskStatus.InProgress]: 'In Progress',
	[TaskStatus.Complete]: 'Complete',
};

// Task priority types
export enum TaskPriority {
	Low = 1,
	Medium = 2,
	High = 3,
	Urgent = 4,
}

// Human-readable task priority names
export const TaskPriorityNames: Record<TaskPriority, string> = {
	[TaskPriority.Low]: 'Low',
	[TaskPriority.Medium]: 'Medium',
	[TaskPriority.High]: 'High',
	[TaskPriority.Urgent]: 'Urgent',
};

// Task priority colors for UI representation
export const TaskPriorityColors: Record<TaskPriority, string> = {
	[TaskPriority.Low]: 'green',
	[TaskPriority.Medium]: 'blue',
	[TaskPriority.High]: 'yellow',
	[TaskPriority.Urgent]: 'red',
};

// Task list scope types (assigned to user or requested by user)
export enum TaskListScope {
	Assigned = 'assigned',
	Requested = 'requested',
}

// Task information
export interface Task {
	id: string; // ULID
	title: string; // short title
	description: string; // detailed description
	notes: string; // notes or comments by the assignee
	due_by: string; // ISO date string
	priority: TaskPriority; // 1 (low), 2 (medium), 3 (high), 4 (urgent)
	progress: number; // percentage 0-100
	status: TaskStatus; // 0 (not started), 1 (in progress), 2 (complete)
	assignee: string; // user_id assigned to
	assigner: string; // user_id assigned by
	organization_id: string; // organization the task belongs to
	created_at: string; // ISO date string
	updated_at: string; // ISO date string
}
