import { env } from "$env/dynamic/public";
import { browser } from "$app/environment";
import { authStore } from "./stores/auth";
import type { components } from "$lib/api/generated/users-api";

export const TOKEN_KEY = "wili_jwt";
export const USER_KEY = "wili_user";
export const JUST_LOGGED_IN_KEY = "wili_just_logged_in";

const AUTH_API_BASE_URL =
  env.PUBLIC_AUTH_API_BASE_URL ??
  env.PUBLIC_API_BASE_URL ??
  (import.meta as any).env?.VITE_AUTH_API_BASE_URL ??
  "https://api.wili.me";

type AuthResponse = components["schemas"]["AuthResponse"];

let storageListenerSet = false;

export function initApi() {
  const token = typeof localStorage !== "undefined" ? localStorage.getItem(TOKEN_KEY) : null;
  const user = typeof localStorage !== "undefined" ? localStorage.getItem(USER_KEY) : null;
  const justLoggedIn =
    typeof localStorage !== "undefined" ? localStorage.getItem(JUST_LOGGED_IN_KEY) : null;

  authStore.set({
    token: token ?? undefined,
    user: user ? (JSON.parse(user) as components["schemas"]["User"]) : null,
    isLoading: false,
    justLoggedIn: justLoggedIn === "true",
  });

  if (browser && !storageListenerSet) {
    window.addEventListener("storage", async (e) => {
      if (e.key === TOKEN_KEY) {
        authStore.update((state) => ({ ...state, token: e.newValue ?? undefined }));
      }
      if (e.key === USER_KEY) {
        authStore.update((state) => ({
          ...state,
          user: JSON.parse(e.newValue ?? "{}") as components["schemas"]["User"],
        }));
      }
      if (e.key === JUST_LOGGED_IN_KEY) {
        authStore.update((state) => ({ ...state, justLoggedIn: e.newValue === "true" }));
      }
    });
    storageListenerSet = true;
  }
}

export function redirectToYandex() {
  if (!browser) return;
  const clientId = import.meta.env.VITE_YANDEX_CLIENT_ID as string;
  const redirectUri = `${window.location.origin}/auth/callback`;
  const url = new URL("https://oauth.yandex.ru/authorize");
  url.searchParams.set("response_type", "code");
  url.searchParams.set("client_id", clientId);
  url.searchParams.set("redirect_uri", redirectUri);
  url.searchParams.set("scope", "login:email login:info");
  window.location.href = url.toString();
}

export const TELEGRAM_AUTH_STATE_KEY = "wili_tg_auth_state";

export function redirectToTelegramBot(botUsername: string) {
  if (!browser) return;
  if (!botUsername) {
    console.warn("Telegram bot username not provided");
    return;
  }
  const state = crypto.randomUUID().replace(/-/g, "").slice(0, 16);
  sessionStorage.setItem(TELEGRAM_AUTH_STATE_KEY, state);
  window.location.href = `https://t.me/${botUsername}?start=webauth_${state}`;
}

export function handleTelegramCallback(token: string, state: string): boolean {
  if (!browser) return false;
  const savedState = sessionStorage.getItem(TELEGRAM_AUTH_STATE_KEY);
  if (!savedState || savedState !== state) {
    console.warn("Telegram auth state mismatch");
    return false;
  }
  sessionStorage.removeItem(TELEGRAM_AUTH_STATE_KEY);

  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(JUST_LOGGED_IN_KEY, "true");

  fetch(`${AUTH_API_BASE_URL}/users/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
    .then((res) => (res.ok ? res.json() : Promise.reject()))
    .then((user) => {
      localStorage.setItem(USER_KEY, JSON.stringify(user));
      authStore.set({ token, user, isLoading: false, justLoggedIn: true });
    })
    .catch(() => {
      authStore.set({ token, user: null, isLoading: false, justLoggedIn: true });
    });

  return true;
}

export async function exchangeCode(code: string) {
  const redirectUri = browser ? `${window.location.origin}/auth/callback` : "";
  const body = { code, redirectUri };
  const res = await fetch(`${AUTH_API_BASE_URL}/auth/yandex`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error("Auth failed");
  const resp = (await res.json()) as AuthResponse;

  const token = resp.accessToken;
  if (browser) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(resp.user));
    localStorage.setItem(JUST_LOGGED_IN_KEY, "true");
    authStore.set({
      token,
      user: resp.user,
      isLoading: false,
      justLoggedIn: true,
    });
  }
}

export async function exchangeTelegramInitData(initData: string) {
  const body = { initData };
  const res = await fetch(`${AUTH_API_BASE_URL}/auth/telegram`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error("Telegram auth failed");
  const resp = (await res.json()) as AuthResponse;

  const token = resp.accessToken;
  if (browser) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(resp.user));
    localStorage.setItem(JUST_LOGGED_IN_KEY, "true");
    authStore.set({
      token,
      user: resp.user,
      isLoading: false,
      justLoggedIn: true,
    });
  }
}

export function logout() {
  if (browser) {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem(JUST_LOGGED_IN_KEY);
  }
  authStore.set({ token: undefined, user: null, isLoading: false, justLoggedIn: false });
}
