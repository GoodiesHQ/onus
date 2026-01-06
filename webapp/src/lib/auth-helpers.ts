import { auth } from '$lib/stores';
import type { AuthSnapshot } from './types';
import type { Optional } from './utils';

// check if user is authenticated
export function isUserAuthenticated(): boolean {
	return auth.isAuthenticated;
}

// get current user
export function getUser(): Optional<AuthSnapshot> {
	return auth.self;
}

// Check if user has at least a specific role
export function hasRole(minRole: number): boolean {
	return auth.self ? auth.self.role >= minRole : false;
}

// Check if user has admin access
export function isAdmin(): boolean {
	return auth.hasAdminAccess;
}

// Get user's organization name
export function getOrganization(): string {
	return auth.organizationName ?? '';
}

// Get user's email
export function getUserEmail(): string {
	return auth.self?.email ?? '';
}

// Get user's name
export function getUserName(): string {
	return auth.self?.name ?? '';
}

// Check if user is disabled
export function isUserDisabled(): boolean {
	return auth.self?.disabled ?? false;
}
