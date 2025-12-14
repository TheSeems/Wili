<script lang="ts">
  import { browser } from "$app/environment";
  import { onMount } from "svelte";
  import { env } from "$env/dynamic/public";
  import { exchangeTelegramInitData, redirectToTelegramBot, JUST_LOGGED_IN_KEY } from "$lib/auth";
  import { authStore } from "$lib/stores/auth";
  import { makeAlert } from "$lib/stores/alerts";
  import CheckCircle2Icon from "@lucide/svelte/icons/check-circle-2";
  import { Button } from "$lib/components/ui/button";
  import { ListIcon } from "@lucide/svelte";
  import { _, isLoading as i18nLoading } from "svelte-i18n";
  import T from "$lib/components/T.svelte";
  import WiliLogo from "$lib/components/WiliLogo.svelte";

  $: ({ token, user, isLoading, justLoggedIn } = $authStore);
  let telegramLoginAvailable = false;
  let telegramInitData: string = "";
  let telegramBotLoginAvailable = false;

  $: if (justLoggedIn && user) {
    makeAlert({
      title: $i18nLoading ? "Welcome!" : $_("home.welcomeTitle") || "Welcome!",
      description: $i18nLoading
        ? `You have successfully logged in to Wili, ${user.displayName}!`
        : $_("home.welcomeDescription", { values: { name: user.displayName } }) ||
          `You have successfully logged in to Wili, ${user.displayName}!`,
      icon: CheckCircle2Icon,
      duration: 5000,
    });
    authStore.update((state) => ({ ...state, justLoggedIn: false }));
    localStorage.removeItem(JUST_LOGGED_IN_KEY);
  }

  function redirectToYandexAuth() {
    if (!browser) return;

    const clientId = import.meta.env.VITE_YANDEX_CLIENT_ID;
    const redirectUri = `${window.location.origin}/auth/callback`;

    const oauthUrl = new URL("https://oauth.yandex.ru/authorize");
    oauthUrl.searchParams.set("response_type", "code");
    oauthUrl.searchParams.set("client_id", clientId);
    oauthUrl.searchParams.set("redirect_uri", redirectUri);
    oauthUrl.searchParams.set("scope", "login:email login:info");

    window.location.href = oauthUrl.toString();
  }

  async function loginWithTelegram() {
    if (!telegramInitData) return;
    await exchangeTelegramInitData(telegramInitData);
  }

  onMount(() => {
    if (!browser) return;
    const tg = (window as any)?.Telegram?.WebApp;
    const fromSdk = (tg?.initData as string | undefined) || "";
    if (fromSdk) {
      telegramInitData = fromSdk;
    } else {
      const hash = window.location.hash?.replace(/^#/, "");
      if (hash) {
        const params = new URLSearchParams(hash);
        telegramInitData = params.get("tgWebAppData") || "";
      }
    }
    telegramLoginAvailable = Boolean(telegramInitData);
    telegramBotLoginAvailable = Boolean(env.PUBLIC_TELEGRAM_BOT_USERNAME);
  });
</script>

<section class="flex h-[80vh] flex-col items-center justify-center px-4 py-10">
  {#if isLoading}
    <div class="flex flex-col items-center gap-4">
      <div class="h-12 w-12 animate-spin rounded-full border-b-2 border-gray-900"></div>
      <p class="text-muted-foreground"><T key="common.loading" fallback="Loading..." /></p>
    </div>
  {:else}
    <h1 class="sr-only"><T key="home.title" fallback="Welcome to Wili" /></h1>
    <WiliLogo className="h-20 md:h-28 lg:h-32 mb-6" />
    <p class="text-muted-foreground max-w-md text-center">
      <T
        key="home.description"
        fallback="Create and share wish-lists with friends, family, or anyone else."
      />
      {#if !token}
        <T key="home.loginPrompt" fallback=" Log in to start building yours." />
      {/if}
    </p>

    {#if token && user}
      <div class="mt-8 flex flex-col items-center gap-4">
        <Button href="/wishlists" class="flex items-center gap-2">
          <ListIcon class="h-4 w-4" />
          <T key="nav.wishlists" fallback="My Wishlists" />
        </Button>
      </div>
    {:else}
      {#if telegramLoginAvailable}
        <Button
          onclick={loginWithTelegram}
          class="mt-8 w-full border border-white/10 bg-black text-white hover:bg-black/90 sm:w-1/3"
          aria-label="Login with Telegram"
        >
          <div class="flex items-center justify-center gap-2">
            <T key="auth.loginWithTelegram" fallback="Login with Telegram" />
          </div>
        </Button>
      {:else if telegramBotLoginAvailable}
        <Button
          onclick={() => redirectToTelegramBot(env.PUBLIC_TELEGRAM_BOT_USERNAME || "")}
          class="mt-8 w-full border border-white/10 bg-black text-white hover:bg-black/90 sm:w-1/3"
          aria-label="Login with Telegram"
        >
          <div class="flex items-center justify-center gap-2">
            <svg viewBox="0 0 24 24" class="h-5 w-5 fill-current"
              ><path
                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .37z"
              /></svg
            >
            <T key="auth.loginWithTelegram" fallback="Login with Telegram" />
          </div>
        </Button>
      {/if}
      <Button
        onclick={redirectToYandexAuth}
        class="mt-4 w-full border border-white/10 bg-black text-white hover:bg-black/90 sm:w-1/3"
        aria-label="Login with Yandex"
      >
        <div class="flex items-center justify-center gap-2">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="40"
            height="40"
            fill="none"
            viewBox="0 0 26 26"
            ><path
              fill="#F8604A"
              d="M26 13c0-7.18-5.82-13-13-13S0 5.82 0 13s5.82 13 13 13 13-5.82 13-13Z"
            ></path><path
              fill="#fff"
              d="M17.55 20.822h-2.622V7.28h-1.321c-2.254 0-3.38 1.127-3.38 2.817 0 1.885.758 2.816 2.448 3.943l1.322.932-3.749 5.828H7.237l3.575-5.265c-2.059-1.495-3.185-2.817-3.185-5.265 0-3.012 2.058-5.07 6.023-5.07h3.9v15.622Z"
            ></path></svg
          >
          <T key="auth.loginWithYandex" fallback="Login with Yandex" />
        </div>
      </Button>
    {/if}
  {/if}
</section>

<!-- Manual OAuth - no external scripts needed -->
