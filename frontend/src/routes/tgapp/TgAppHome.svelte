<script lang="ts">
  import type { components } from "$lib/api/generated/wishlist-api";
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import WiliLogo from "$lib/components/WiliLogo.svelte";
  import { LinkIcon } from "@lucide/svelte";

  type Wishlist = components["schemas"]["Wishlist"];

  interface Props {
    isLoggedIn: boolean;
    userName: string;
    telegramLoginAvailable: boolean;
    telegramLoginLoading: boolean;
    creatingWishlist: boolean;
    wishlists: Wishlist[];
    wishlistsLoading: boolean;
    wishlistsError: string | null;
    onLogin: () => void;
    onCreateWishlist: () => void;
    onSelectWishlist: (id: string) => void;
  }

  let {
    isLoggedIn,
    userName,
    telegramLoginAvailable,
    telegramLoginLoading,
    creatingWishlist,
    wishlists,
    wishlistsLoading,
    wishlistsError,
    onLogin,
    onCreateWishlist,
    onSelectWishlist,
  }: Props = $props();
</script>

<div class="flex flex-col items-center gap-6 py-10 text-center">
  <div class="flex flex-col items-center gap-3">
    <WiliLogo className="h-14 w-auto" />
    <p class="text-muted-foreground max-w-sm text-sm leading-relaxed">
      {#if isLoggedIn}
        {$_("tgapp.homeCreatePrompt", { values: { name: userName } })}
      {:else if telegramLoginAvailable}
        {$_("tgapp.homeLoginPrompt")}
      {:else}
        {$_("tgapp.openFromChat")}
      {/if}
    </p>
  </div>

  <div class="w-full max-w-sm space-y-2">
    {#if !isLoggedIn && telegramLoginAvailable}
      <Button disabled={telegramLoginLoading} onclick={onLogin} class="w-full">
        {telegramLoginLoading ? $_("common.loading") : $_("auth.loginWithTelegram")}
      </Button>
    {/if}
    {#if isLoggedIn}
      <Button disabled={creatingWishlist} onclick={onCreateWishlist} class="w-full">
        {creatingWishlist ? $_("common.loading") : $_("tgapp.createWishlist")}
      </Button>
    {/if}
  </div>

  {#if isLoggedIn}
    <div class="w-full max-w-sm space-y-2 text-left">
      {#if wishlistsLoading}
        <div class="text-muted-foreground text-center text-sm">{$_("common.loading")}</div>
      {:else if wishlistsError}
        <div class="text-muted-foreground text-center text-sm">{wishlistsError}</div>
      {:else if wishlists.length > 0}
        {#each wishlists.slice(0, 8) as w}
          <button
            type="button"
            class="border-border hover:bg-muted flex w-full items-center justify-between rounded-lg border px-4 py-3 text-left"
            onclick={() => onSelectWishlist(w.id)}
          >
            <span class="truncate font-medium">{w.title}</span>
            <span class="text-muted-foreground ml-3 text-sm">{w.items?.length || 0}</span>
          </button>
        {/each}
      {:else}
        <div class="text-muted-foreground text-center text-sm">
          {$_("wishlists.noWishlists")}
        </div>
      {/if}
    </div>
  {/if}

  <a
    class="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 text-sm"
    href="https://wili.me"
    target="_blank"
    rel="noreferrer"
  >
    <LinkIcon class="h-4 w-4" />
    {$_("tgapp.openInBrowser")}
  </a>
</div>
