<script lang="ts">
  import { authStore } from "$lib/stores/auth";
  import { invalidateAll, goto } from "$app/navigation";
  import { browser } from "$app/environment";
  import SunIcon from "@lucide/svelte/icons/sun";
  import MoonIcon from "@lucide/svelte/icons/moon";
  import LogOutIcon from "@lucide/svelte/icons/log-out";
  import ListIcon from "@lucide/svelte/icons/list";
  import { toggleMode } from "mode-watcher";
  import { Button } from "$lib/components/ui/button/index.js";
  import { TOKEN_KEY, USER_KEY } from "$lib/auth";
  import { Avatar, AvatarImage, AvatarFallback } from "$lib/components/ui/avatar";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
  import LanguageSwitcher from "$lib/components/LanguageSwitcher.svelte";
  import T from "$lib/components/T.svelte";
  import { _, isLoading as i18nLoading } from "svelte-i18n";
  import WiliLogo from "$lib/components/WiliLogo.svelte";
  let user: {
    displayName: string;
    email?: string | null;
    avatarUrl?: string | null;
  } | null = null;
  let token: string | undefined;
  let isLoading: boolean;

  $: {
    const authState = $authStore;
    token = authState.token;
    user = authState.user;
    isLoading = authState.isLoading;
  }

  async function logout() {
    const confirmMessage = $i18nLoading
      ? "Are you sure you want to log out?"
      : $_("nav.confirmLogout") || "Are you sure you want to log out?";
    if (!confirm(confirmMessage)) return;

    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    authStore.update((state) => ({
      ...state,
      token: undefined,
      user: null,
    }));

    invalidateAll();
    if (browser) {
      window.location.href = "/";
    }
  }
</script>

<nav
  class="sticky top-0 z-10 flex w-full items-center justify-between px-4 py-2"
  style="backdrop-filter: blur(8px);"
>
  <div class="flex items-center gap-6">
    <a href="/" class="flex items-center" aria-label="Go to home">
      <WiliLogo className="h-7 -translate-y-0.5" />
    </a>
    {#if !isLoading && (user || token)}
      <nav class="flex items-center gap-4">
        <a
          href="/wishlists"
          class="hover:text-primary flex h-7 items-center text-sm leading-none font-medium transition-colors"
        >
          <T key="nav.wishlists" fallback="Wishlists" />
        </a>
      </nav>
    {/if}
  </div>

  <div class="flex items-center gap-3">
    <LanguageSwitcher />
    <Button onclick={toggleMode} variant="ghost" size="icon" class="h-8 w-8">
      <SunIcon class="h-4 w-4 scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90" />
      <MoonIcon
        class="absolute h-4 w-4 scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"
      />
      <span class="sr-only"><T key="nav.toggleTheme" fallback="Toggle theme" /></span>
    </Button>

    {#if !isLoading && (user || token)}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          <button class="flex h-8 w-8 items-center justify-center rounded-full">
            <Avatar class="h-8 w-8">
              <AvatarImage src={user?.avatarUrl} alt={user?.displayName} />
              <AvatarFallback class="text-xs">{user?.displayName?.charAt(0)}</AvatarFallback>
            </Avatar>
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content class="w-48" align="start">
          <DropdownMenu.Label class="text-xs font-medium"
            >{user?.displayName || "User"}</DropdownMenu.Label
          >
          {#if user?.email}
            <div class="text-muted-foreground px-2 py-1 text-xs">
              {user.email}
            </div>
          {/if}
          <DropdownMenu.Separator />
          <DropdownMenu.Item onclick={logout} class="text-xs">
            <LogOutIcon class="mr-2 h-3.5 w-3.5" />
            <T key="nav.logout" fallback="Log out" />
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onclick={() => {
              goto("/wishlists");
            }}
            class="text-xs"
          >
            <ListIcon class="mr-2 h-3.5 w-3.5" />
            <T key="nav.viewMyWishlists" fallback="View my wishlists" />
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {:else if isLoading}
      <div class="h-8 w-8 animate-pulse rounded-full bg-gray-300"></div>
    {/if}
  </div>
</nav>
