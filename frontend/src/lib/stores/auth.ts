import { writable } from "svelte/store";
import type { components } from "$lib/api/generated/users-api";

type User = components["schemas"]["User"];

export interface AuthState {
  token?: string;
  user: User | null;
  isLoading: boolean;
  justLoggedIn: boolean;
}

export const authStore = writable<AuthState>({ user: null, isLoading: true, justLoggedIn: false });
