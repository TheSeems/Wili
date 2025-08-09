import { env } from "$env/dynamic/public";
import { logout } from "$lib/auth";
import type { components } from "$lib/api/generated/wishlist-api";

type Wishlist = components["schemas"]["Wishlist"];
type WishlistItem = components["schemas"]["WishlistItem"];
type CreateWishlistRequest = components["schemas"]["CreateWishlistRequest"];
type UpdateWishlistRequest = components["schemas"]["UpdateWishlistRequest"];
type CreateWishlistItemRequest = components["schemas"]["CreateWishlistItemRequest"];
type UpdateWishlistItemRequest = components["schemas"]["UpdateWishlistItemRequest"];

// Resolve base URL from SvelteKit public env or Vite env, with localhost fallback
const PUBLIC_WISHLIST_API_BASE_URL =
  env.PUBLIC_WISHLIST_API_BASE_URL ??
  (import.meta as any).env?.VITE_WISHLIST_API_BASE_URL ??
  "http://localhost:8081";

export class WishlistApiClient {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    // Use provided baseUrl, or get from environment, or fallback to localhost
    this.baseUrl = baseUrl || "http://localhost:8081";
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    token?: string,
    expectJson: boolean = true
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
      if (response.status === 401) {
        logout();
      }
      throw new Error(`API request failed: ${response.status} ${response.statusText}`, {
        cause: response,
      });
    }

    if (!expectJson) {
      return undefined as T;
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text) {
      return undefined as T;
    }

    return JSON.parse(text);
  }

  // Wishlist operations
  async getWishlists(token: string): Promise<{ wishlists: Wishlist[] }> {
    return this.request("", { method: "GET" }, token);
  }

  async getWishlist(id: string, token?: string): Promise<Wishlist> {
    return this.request(`/${id}`, { method: "GET" }, token);
  }

  async createWishlist(data: CreateWishlistRequest, token: string): Promise<Wishlist> {
    return this.request(
      "",
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    );
  }

  async updateWishlist(id: string, data: UpdateWishlistRequest, token: string): Promise<Wishlist> {
    return this.request(
      `/${id}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
      token
    );
  }

  async deleteWishlist(id: string, token: string): Promise<void> {
    await this.request(`/${id}`, { method: "DELETE" }, token, false);
  }

  // Wishlist item operations
  async addWishlistItem(
    wishlistId: string,
    data: CreateWishlistItemRequest,
    token: string
  ): Promise<WishlistItem> {
    return this.request(
      `/${wishlistId}/items`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    );
  }

  async updateWishlistItem(
    wishlistId: string,
    itemId: string,
    data: UpdateWishlistItemRequest,
    token: string
  ): Promise<WishlistItem> {
    return this.request(
      `/${wishlistId}/items/${itemId}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
      token
    );
  }

  async deleteWishlistItem(wishlistId: string, itemId: string, token: string): Promise<void> {
    await this.request(`/${wishlistId}/items/${itemId}`, { method: "DELETE" }, token, false);
  }
}

// Default client instance
export const wishlistApi = new WishlistApiClient(PUBLIC_WISHLIST_API_BASE_URL);
