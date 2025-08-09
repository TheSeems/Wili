import { browser } from "$app/environment";
import { env } from "$env/dynamic/public";
import { logout } from "$lib/auth";
import type { components } from "$lib/api/generated/users-api";

type User = components["schemas"]["User"];
type YandexAuthRequest = components["schemas"]["YandexAuthRequest"];
type AuthResponse = components["schemas"]["AuthResponse"];
type UpdateUserRequest = components["schemas"]["UpdateUserRequest"];

// Resolve base URL from SvelteKit public env or Vite env, with localhost fallback
const PUBLIC_USER_API_BASE_URL =
  env.PUBLIC_USER_API_BASE_URL ??
  (import.meta as any).env?.VITE_USER_API_BASE_URL ??
  "http://localhost:8080";

export class UserApiClient {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    // Use provided baseUrl, or get from environment, or fallback to localhost
    this.baseUrl = baseUrl || "http://localhost:8080";
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    token?: string
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      if (response.status === 401 && browser) {
        logout();
      }
      throw new Error(`API request failed: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  // Auth operations
  async authenticateWithYandex(data: YandexAuthRequest): Promise<AuthResponse> {
    return this.request("/auth/yandex", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // User operations
  async getCurrentUser(token: string): Promise<User> {
    return this.request("/users/me", { method: "GET" }, token);
  }

  async updateCurrentUser(data: UpdateUserRequest, token: string): Promise<User> {
    return this.request(
      "/users/me",
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
      token
    );
  }

  async getUser(id: string, token?: string): Promise<User> {
    return this.request(`/users/${id}`, { method: "GET" }, token);
  }
}

// Default client instance
export const userApi = new UserApiClient(PUBLIC_USER_API_BASE_URL);
