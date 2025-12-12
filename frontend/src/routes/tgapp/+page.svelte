<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import type { components } from "$lib/api/generated/wishlist-api";
  import { wishlistApi } from "$lib/api/wishlist-client";
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "$lib/components/ui/card";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Input } from "$lib/components/ui/input";
  import { saveBookingToken, getBookingToken, removeBookingToken } from "$lib/utils/booking-storage";
  import { showSuccessAlert, showInfoAlert } from "$lib/utils/alerts";
  import { authStore } from "$lib/stores/auth";
  import { exchangeTelegramInitData } from "$lib/auth";
  import {
    Loader2Icon,
    LinkIcon,
    CheckIcon,
    XIcon,
    ShieldOffIcon,
    SendIcon,
  } from "@lucide/svelte";

  type Wishlist = components["schemas"]["Wishlist"];
  type WishlistItem = components["schemas"]["WishlistItem"];

  let wishlist: Wishlist | null = $state(null);
  let listId: string | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Booking form state
  let bookingItemId: string | null = $state(null);
  let anonymous = $state(false);
  let message = $state("");
  let defaultName: string | null = $state(null);
  let telegramAuthAttempted = $state(false);

  function parseListId(): string | null {
    if (typeof window === "undefined") return null;

    const fromStart = (window as any)?.Telegram?.WebApp?.initDataUnsafe?.start_param as string | undefined;
    const fromHashStart = (() => {
      try {
        const hash = window.location.hash?.replace(/^#/, "");
        if (!hash) return undefined;
        const params = new URLSearchParams(hash);
        const raw = params.get("tgWebAppData");
        if (!raw) return undefined;
        const inner = new URLSearchParams(raw);
        return (inner.get("start_param") || inner.get("startapp") || undefined) as string | undefined;
      } catch {
        return undefined;
      }
    })();
    const fromQuery =
      page?.url?.searchParams?.get("start") ||
      page?.url?.searchParams?.get("list") ||
      page?.url?.searchParams?.get("startapp") ||
      page?.url?.searchParams?.get("tgWebAppStartParam");

    const candidate = fromStart || fromHashStart || fromQuery || "";
    const match = candidate.match(/^list_([0-9a-fA-F-]{36})$/);
    return match ? match[1] : null;
  }

  async function loadWishlist() {
    if (!listId) return;
    try {
      loading = true;
      error = null;
      wishlist = await wishlistApi.getWishlist(listId);
    } catch (err) {
      console.error("Failed to load wishlist", err);
      error = $_("tgapp.loadError");
    } finally {
      loading = false;
    }
  }

  function startBooking(itemId: string) {
    bookingItemId = itemId;
    anonymous = defaultName ? false : true;
    message = "";
  }

  function closeBooking() {
    bookingItemId = null;
    anonymous = false;
    message = "";
  }

  function resolveBookerName(): string | undefined {
    if (anonymous) return undefined;
    return defaultName || undefined;
  }

  async function book(item: WishlistItem) {
    if (!listId) return;
    try {
      const payload = {
        bookerName: resolveBookerName(),
        message: message.trim() || undefined,
      };
      const resp = await wishlistApi.bookItem(listId, item.id, payload);
      saveBookingToken(listId, item.id, resp.cancellationToken);
      await loadWishlist();
      closeBooking();
      showSuccessAlert($_("items.bookedSuccessfully"), undefined, "bottom-center");
    } catch (err) {
      console.error("Booking failed", err);
      error = err instanceof Error ? err.message : $_("tgapp.bookFailed");
    }
  }

  async function unbook(item: WishlistItem) {
    if (!listId) return;
    const token = getBookingToken(listId, item.id);
    if (!token) {
      error = $_("items.noCancellationToken");
      return;
    }
    try {
      await wishlistApi.unbookItemByToken(listId, item.id, token);
      removeBookingToken(listId, item.id);
      await loadWishlist();
      showInfoAlert($_("items.bookingCancelled"), undefined, "bottom-center");
    } catch (err) {
      console.error("Unbook failed", err);
      error = err instanceof Error ? err.message : $_("tgapp.unbookFailed");
    }
  }

  function extractUserFromHash(): { first_name?: string; last_name?: string; username?: string } | null {
    try {
      const hash = window.location.hash?.replace(/^#/, "");
      if (!hash) return null;
      const params = new URLSearchParams(hash);
      const raw = params.get("tgWebAppData");
      if (!raw) return null;
      const inner = new URLSearchParams(raw);
      const userRaw = inner.get("user");
      if (!userRaw) return null;
      return JSON.parse(userRaw);
    } catch (e) {
      console.warn("failed to parse tgWebAppData", e);
      return null;
    }
  }

  function setDefaultNameFromUser(user: any) {
    if (!user) return;
    const nameParts = [user.first_name, user.last_name].filter(Boolean);
    const usernameTag = user.username ? `@${user.username}` : "";
    const fullName = nameParts.join(" ").trim();
    const resolved =
      (fullName && usernameTag ? `${fullName} ${usernameTag}` : fullName || usernameTag) || null;
    if (resolved) {
      defaultName = resolved;
      anonymous = false;
    }
  }

  onMount(() => {
    listId = parseListId();
    if (!listId) {
      error = $_("tgapp.badLink");
      loading = false;
      return;
    }

    const tg = (window as any)?.Telegram?.WebApp;
    if (tg) {
      tg.ready?.();
      tg.expand?.();
      tg.setHeaderColor?.("bg_color");
      tg.setBackgroundColor?.("bg_color");
      tg.setBottomBarColor?.("bg_color");
      const user = tg.initDataUnsafe?.user;
      setDefaultNameFromUser(user);

      const initData = (tg.initData as string | undefined) || "";
      if (!telegramAuthAttempted && !$authStore.token && initData) {
        telegramAuthAttempted = true;
        exchangeTelegramInitData(initData).catch((e) =>
          console.warn("telegram auth failed", e)
        );
      }
    }
    if (!defaultName) {
      const hashUser = extractUserFromHash();
      setDefaultNameFromUser(hashUser);
    }
    if (!defaultName) {
      anonymous = true;
    }

    loadWishlist();
  });

  function openExternal(url: string) {
    if (!url) return;
    if (confirm($_("tgapp.openLinkConfirm"))) {
      window.open(url, "_blank", "noreferrer");
    }
  }

  function shortUrl(url: string): string {
    return url.replace(/^https?:\/\//, "");
  }

  function shareToTelegram() {
    if (!wishlist || !listId) return;

    const tg = (window as any)?.Telegram?.WebApp;
    if (tg?.switchInlineQuery) {
      try {
        tg.switchInlineQuery(`wishlist:${listId}`, [
          "users",
          "groups",
          "supergroups",
          "channels",
        ]);
        return;
      } catch (e) {
        console.warn("Telegram switchInlineQuery failed, falling back", e);
      }
    }

    const shareUrl = `${window.location.origin}/wishlists/${wishlist.id}`;
    const text = `Wishlist: ${wishlist.title || ""}`.trim();
    const tgShare = `https://t.me/share/url?url=${encodeURIComponent(shareUrl)}&text=${encodeURIComponent(text)}`;
    window.open(tgShare, "_blank", "noopener");
  }
</script>

<svelte:head>
  <title>{$_("tgapp.pageTitle")}</title>
</svelte:head>

<div class="mx-auto flex max-w-5xl flex-col gap-4 px-4 py-6">
  {#if loading}
    <div class="flex items-center gap-3 text-muted-foreground">
      <Loader2Icon class="h-5 w-5 animate-spin" />
      <span>{$_("tgapp.loadingWishlist")}</span>
    </div>
  {:else if error}
    <Card class="border-destructive">
      <CardContent class="pt-6 text-destructive">
        {error}
      </CardContent>
    </Card>
  {:else if wishlist}
    <div class="p-0">
      <div class="flex items-start justify-between gap-3">
        <div class="space-y-2">
          <p class="text-xl font-semibold">{wishlist.title}</p>
          {#if wishlist.description}
            <ExpandableText
              content={wishlist.description}
              className="text-sm text-muted-foreground"
              maxHeight={200}
              useResponsive={false}
              allowYandexMarket={false}
              smallOverflowThreshold={0}
            />
          {/if}
        </div>
        <Button variant="outline" class="shrink-0 gap-2" onclick={shareToTelegram}>
          <SendIcon class="h-4 w-4" />
          {$_("wishlists.shareToTelegram")}
        </Button>
        {#if wishlist.description}
          <!-- description moved above -->
        {/if}
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      {#each wishlist.items || [] as item}
        <Card class="h-full">
          <CardHeader>
            <CardTitle class="line-clamp-2">{item.data?.name || $_("wishlists.untitledItem")}</CardTitle>
            {#if item.data?.description}
              <CardDescription class="text-sm text-muted-foreground">
                <ExpandableText
                  content={item.data.description}
                  className="text-sm text-muted-foreground"
                  maxHeight={180}
                  useResponsive={false}
                  allowYandexMarket={false}
                  smallOverflowThreshold={0}
                />
              </CardDescription>
            {/if}
          </CardHeader>
          <CardContent class="flex flex-col gap-3">
            {#if item.booking}
              <div class="flex flex-wrap items-center justify-between gap-2 rounded border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-800 dark:border-green-900/40 dark:bg-green-900/20 dark:text-green-200">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-2">
                    <CheckIcon class="h-4 w-4 shrink-0" />
                    <span>{$_("tgapp.alreadyBooked")}</span>
                  </div>
                  {#if item.booking.bookerName}
                    <div class="text-green-700/80 dark:text-green-200/80 text-sm">{item.booking.bookerName}</div>
                  {/if}
                </div>
                {#if listId && getBookingToken(listId, item.id)}
                  <Button size="sm" variant="outline" onclick={() => unbook(item)}>
                    <XIcon class="mr-2 h-4 w-4" />
                    {$_("common.cancel")}
                  </Button>
                {/if}
              </div>
            {:else}
              {#if bookingItemId === item.id}
                <div class="space-y-3 px-1 py-1">
                  <label class="flex items-center gap-2 text-sm">
                    {#if defaultName}
                      <input type="checkbox" bind:checked={anonymous} />
                      {$_("tgapp.bookAnonymously")}
                    {:else}
                      <span class="text-muted-foreground">
                        {$_("tgapp.willBeAnonymous")}
                      </span>
                    {/if}
                  </label>
                  {#if !anonymous && defaultName}
                    <div class="text-muted-foreground text-sm">
                      {$_("tgapp.nameLabel")}: {defaultName}
                    </div>
                  {/if}
                  <div class="space-y-2">
                    <Input
                      placeholder={$_("tgapp.messagePlaceholder")}
                      bind:value={message}
                      class="w-full text-sm"
                    />
                  </div>
                  <div class="mt-3 flex gap-3">
                    <Button class="gap-2" onclick={() => book(item)}>
                      <CheckIcon class="h-4 w-4" />
                      {$_("items.bookItem")}
                    </Button>
                    <Button variant="outline" onclick={closeBooking}>
                      <XIcon class="h-4 w-4" />
                      {$_("common.cancel")}
                    </Button>
                  </div>
                </div>
              {:else}
                <Button class="w-full gap-2" onclick={() => startBooking(item.id)}>
                  <ShieldOffIcon class="h-4 w-4" />
                  {$_("items.bookItem")}
                </Button>
              {/if}
            {/if}
          </CardContent>
        </Card>
      {/each}
    </div>

    <div class="mt-6 text-center text-sm text-muted-foreground">
      <a
        class="inline-flex items-center gap-2 rounded-full border border-border px-3 py-1 text-primary underline-offset-4 hover:text-primary"
        href={wishlist ? `${window.location.origin}/wishlists/${wishlist.id}` : "#"}
        target="_blank"
        rel="noreferrer"
      >
        <LinkIcon class="h-4 w-4" />
        {$_("tgapp.openInBrowser")}
      </a>
    </div>
  {/if}
</div>

