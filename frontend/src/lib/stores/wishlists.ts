import { writable } from "svelte/store";
import type { components } from "$lib/api/generated/wishlist-api";

type Wishlist = components["schemas"]["Wishlist"];

interface WishlistStore {
  wishlists: Wishlist[];
  currentWishlist: Wishlist | null;
  loading: boolean;
  error: string | null;
}

const initialState: WishlistStore = {
  wishlists: [],
  currentWishlist: null,
  loading: false,
  error: null,
};

export const wishlistStore = writable<WishlistStore>(initialState);

export const wishlistActions = {
  setLoading: (loading: boolean) => {
    wishlistStore.update((state) => ({ ...state, loading }));
  },

  setError: (error: string | null) => {
    wishlistStore.update((state) => ({ ...state, error }));
  },

  setWishlists: (wishlists: Wishlist[]) => {
    wishlistStore.update((state) => ({ ...state, wishlists, error: null }));
  },

  setCurrentWishlist: (wishlist: Wishlist | null) => {
    wishlistStore.update((state) => ({ ...state, currentWishlist: wishlist }));
  },

  addWishlist: (wishlist: Wishlist) => {
    wishlistStore.update((state) => ({
      ...state,
      wishlists: [...state.wishlists, wishlist],
    }));
  },

  updateWishlist: (updatedWishlist: Wishlist) => {
    wishlistStore.update((state) => ({
      ...state,
      wishlists: state.wishlists.map((w) => (w.id === updatedWishlist.id ? updatedWishlist : w)),
      currentWishlist:
        state.currentWishlist?.id === updatedWishlist.id ? updatedWishlist : state.currentWishlist,
    }));
  },

  removeWishlist: (wishlistId: string) => {
    wishlistStore.update((state) => ({
      ...state,
      wishlists: state.wishlists.filter((w) => w.id !== wishlistId),
      currentWishlist: state.currentWishlist?.id === wishlistId ? null : state.currentWishlist,
    }));
  },

  reset: () => {
    wishlistStore.set(initialState);
  },
};
