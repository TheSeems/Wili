<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { env } from "$env/dynamic/public";
  import type { components } from "$lib/api/generated/wishlist-api";
  import { wishlistApi } from "$lib/api/wishlist-client";
  import { _, locale } from "svelte-i18n";
  import { Card, CardContent } from "$lib/components/ui/card";
  import {
    saveBookingToken,
    getBookingToken,
    removeBookingToken,
  } from "$lib/utils/booking-storage";
  import { showSuccessAlert, showInfoAlert } from "$lib/utils/alerts";
  import { authStore } from "$lib/stores/auth";
  import { exchangeTelegramInitData } from "$lib/auth";
  import { Loader2Icon, LinkIcon } from "@lucide/svelte";
  import TgAppHome from "./TgAppHome.svelte";
  import WishlistHeader from "./WishlistHeader.svelte";
  import WishlistItemCard from "./WishlistItemCard.svelte";
  import AddItemForm from "./AddItemForm.svelte";

  type Wishlist = components["schemas"]["Wishlist"];
  type WishlistItem = components["schemas"]["WishlistItem"];

  let wishlist: Wishlist | null = $state(null);
  let listId: string | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  let defaultName: string | null = $state(null);
  let telegramAuthAttempted = $state(false);
  let telegramLoginLoading = $state(false);
  let telegramLoginAvailable = $state(false);
  let telegramInitData = $state("");
  let creatingWishlist = $state(false);
  let deletingWishlist = $state(false);
  let deletingItemId = $state<string | null>(null);
  let revealedBookings = $state<Set<string>>(new Set());

  const telegramBotUsername = env.PUBLIC_TELEGRAM_BOT_USERNAME;
  let myWishlists: Wishlist[] = $state([]);
  let myWishlistsLoading = $state(false);
  let myWishlistsError = $state<string | null>(null);
  let myWishlistsToken = $state<string | null>(null);

  function parseListId(): string | null {
    if (typeof window === "undefined") return null;

    const fromStart = (window as any)?.Telegram?.WebApp?.initDataUnsafe?.start_param as
      | string
      | undefined;
    const fromHashStart = (() => {
      try {
        const hash = window.location.hash?.replace(/^#/, "");
        if (!hash) return undefined;
        const params = new URLSearchParams(hash);
        const raw = params.get("tgWebAppData");
        if (!raw) return undefined;
        const inner = new URLSearchParams(raw);
        return (inner.get("start_param") || inner.get("startapp") || undefined) as
          | string
          | undefined;
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
      wishlist = await wishlistApi.getWishlist(listId, $authStore.token || undefined);
    } catch (err) {
      console.error("Failed to load wishlist", err);
      error = $_("tgapp.loadError");
    } finally {
      loading = false;
    }
  }

  async function loadMyWishlists() {
    if (!$authStore.token) return;
    if (myWishlistsLoading) return;
    if (myWishlistsToken === $authStore.token) return;
    myWishlistsLoading = true;
    myWishlistsError = null;
    try {
      const data = await wishlistApi.getWishlists($authStore.token);
      myWishlists = data.wishlists || [];
      myWishlistsToken = $authStore.token;
    } catch (e) {
      console.warn("failed to load wishlists", e);
      myWishlistsError = $_("wishlists.failedToLoad");
    } finally {
      myWishlistsLoading = false;
    }
  }

  function extractUserFromHash(): {
    first_name?: string;
    last_name?: string;
    username?: string;
    language_code?: string;
  } | null {
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
    }
  }

  function setLocaleFromTelegram(langCode: string | undefined) {
    if (!langCode) return;
    const code = langCode.toLowerCase();
    if (code.startsWith("ru")) {
      locale.set("ru");
    } else if (code.startsWith("en")) {
      locale.set("en");
    }
  }

  async function loginWithTelegram(showFeedback: boolean) {
    if (!telegramInitData) return;
    telegramLoginLoading = true;
    try {
      await exchangeTelegramInitData(telegramInitData);
      if (listId) await loadWishlist();
    } catch (e) {
      console.warn("telegram auth failed", e);
      if (showFeedback) {
        showInfoAlert($_("tgapp.loadError"), undefined, "bottom-center");
      }
    } finally {
      telegramLoginLoading = false;
    }
  }

  onMount(() => {
    listId = parseListId();
    const tg = (window as any)?.Telegram?.WebApp;
    if (tg) {
      tg.ready?.();
      tg.expand?.();
      tg.setHeaderColor?.("bg_color");
      tg.setBackgroundColor?.("bg_color");
      tg.setBottomBarColor?.("bg_color");
      const user = tg.initDataUnsafe?.user;
      setDefaultNameFromUser(user);
      setLocaleFromTelegram(user?.language_code);

      const fromSdk = (tg.initData as string | undefined) || "";
      if (fromSdk) {
        telegramInitData = fromSdk;
      }
    } else {
      const hash = window.location.hash?.replace(/^#/, "");
      if (hash) {
        const params = new URLSearchParams(hash);
        telegramInitData = params.get("tgWebAppData") || "";
      }
    }

    telegramLoginAvailable = Boolean(telegramInitData);

    if (!defaultName) {
      const hashUser = extractUserFromHash();
      setDefaultNameFromUser(hashUser);
      setLocaleFromTelegram(hashUser?.language_code);
    }

    if (!telegramAuthAttempted && !$authStore.token && telegramInitData) {
      telegramAuthAttempted = true;
      void loginWithTelegram(false);
    }

    if (!listId) {
      const sp = new URLSearchParams(window.location.search);
      const hasTgParams =
        telegramLoginAvailable ||
        sp.has("tgWebAppVersion") ||
        sp.has("tgWebAppPlatform") ||
        sp.has("tgWebAppStartParam") ||
        Boolean(window.location.hash?.includes("tgWebAppData="));

      if (!hasTgParams) {
        window.location.replace("https://wili.me");
        return;
      }

      loading = false;
      error = null;
      if ($authStore.token) void loadMyWishlists();
      return;
    }

    loadWishlist();
  });

  $effect(() => {
    if (!listId && $authStore.token) {
      void loadMyWishlists();
    }
  });

  function isOwner(): boolean {
    if (!wishlist) return false;
    return Boolean($authStore.token && $authStore.user && wishlist.userId === $authStore.user.id);
  }

  function revealBooking(itemId: string) {
    if (!confirm($_("items.confirmReveal"))) return;
    revealedBookings.add(itemId);
    revealedBookings = new Set(revealedBookings);
  }

  function isBookingRevealed(itemId: string): boolean {
    return revealedBookings.has(itemId);
  }

  function shareWishlistToTelegram() {
    if (!wishlist) return;
    const tg = (window as any)?.Telegram?.WebApp;
    if (tg?.switchInlineQuery) {
      try {
        tg.switchInlineQuery(`wishlist:${wishlist.id}`, [
          "users",
          "groups",
          "supergroups",
          "channels",
        ]);
        return;
      } catch (e) {
        console.warn("Telegram switchInlineQuery failed", e);
      }
    }

    if (telegramBotUsername) {
      const link = `https://t.me/${telegramBotUsername}?start=share_${wishlist.id}`;
      if (tg?.openTelegramLink) {
        tg.openTelegramLink(link);
        return;
      }
      window.open(link, "_blank", "noopener");
    }
  }

  async function deleteWishlist() {
    if (!wishlist || !$authStore.token) return;
    if (!confirm($_("wishlists.confirmDelete"))) return;
    deletingWishlist = true;
    try {
      await wishlistApi.deleteWishlist(wishlist.id, $authStore.token);
      wishlist = null;
      listId = null;
      myWishlistsToken = null;
      void loadMyWishlists();
      showSuccessAlert($_("tgapp.wishlistDeleted"), undefined, "bottom-center");
    } catch (e) {
      console.warn("wishlist delete failed", e);
      showInfoAlert($_("wishlists.failedToDelete"), undefined, "bottom-center");
    } finally {
      deletingWishlist = false;
    }
  }

  async function saveWishlistEdits(title: string, description: string) {
    if (!wishlist || !$authStore.token) return;
    try {
      wishlist = await wishlistApi.updateWishlist(wishlist.id, { title, description }, $authStore.token);
      showSuccessAlert($_("common.saved"), undefined, "bottom-center");
    } catch (e) {
      console.warn("wishlist update failed", e);
      showInfoAlert($_("wishlists.failedToUpdate"), undefined, "bottom-center");
    }
  }

  async function addItem(name: string, description: string) {
    if (!wishlist || !$authStore.token) return;
    try {
      await wishlistApi.addWishlistItem(
        wishlist.id,
        {
          type: "general",
          data: { name, ...(description ? { description } : {}) },
        } as any,
        $authStore.token
      );
      await loadWishlist();
      showSuccessAlert($_("items.addedSuccessfully"), undefined, "bottom-center");
    } catch (e) {
      console.warn("add item failed", e);
      showInfoAlert($_("items.failedToAdd"), undefined, "bottom-center");
    }
  }

  async function saveItem(item: WishlistItem, name: string, description: string) {
    if (!wishlist || !$authStore.token) return;
    try {
      await wishlistApi.updateWishlistItem(
        wishlist.id,
        item.id,
        {
          type: item.type,
          data: { name, ...(description ? { description } : {}) },
        } as any,
        $authStore.token
      );
      await loadWishlist();
      showSuccessAlert($_("common.saved"), undefined, "bottom-center");
    } catch (e) {
      console.warn("update item failed", e);
      showInfoAlert($_("items.failedToUpdate"), undefined, "bottom-center");
      throw e;
    }
  }

  async function deleteItem(item: WishlistItem) {
    if (!wishlist || !$authStore.token) return;
    deletingItemId = item.id;
    try {
      await wishlistApi.deleteWishlistItem(wishlist.id, item.id, $authStore.token);
      await loadWishlist();
      showSuccessAlert($_("common.deleted"), undefined, "bottom-center");
    } catch (e) {
      console.warn("delete item failed", e);
      showInfoAlert($_("items.failedToDelete"), undefined, "bottom-center");
    } finally {
      deletingItemId = null;
    }
  }

  async function bookItem(item: WishlistItem, anonymous: boolean, message: string) {
    if (!listId) return;
    try {
      const bookerName = anonymous ? undefined : defaultName || undefined;
      const resp = await wishlistApi.bookItem(listId, item.id, {
        bookerName,
        message: message || undefined,
      });
      saveBookingToken(listId, item.id, resp.cancellationToken);
      await loadWishlist();
      showSuccessAlert($_("items.bookedSuccessfully"), undefined, "bottom-center");
    } catch (e) {
      console.warn("booking failed", e);
      showInfoAlert($_("tgapp.bookFailed"), undefined, "bottom-center");
      throw e;
    }
  }

  async function unbookItem(item: WishlistItem) {
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
    } catch (e) {
      console.warn("unbook failed", e);
      showInfoAlert($_("tgapp.unbookFailed"), undefined, "bottom-center");
      throw e;
    }
  }

  async function createMyWishlist() {
    if (!$authStore.token) return;
    creatingWishlist = true;
    try {
      const title = $_("wishlists.newWishlist");
      const description = $_("wishlists.newWishlistDescription");
      const wl = await wishlistApi.createWishlist({ title, description }, $authStore.token);
      showSuccessAlert($_("tgapp.wishlistCreated"), undefined, "bottom-center");
      wishlist = wl as Wishlist;
      listId = wl.id;
      error = null;
    } catch (e) {
      console.warn("create wishlist failed", e);
      showInfoAlert($_("wishlists.failedToCreate"), undefined, "bottom-center");
    } finally {
      creatingWishlist = false;
    }
  }

  function goBackToList() {
    listId = null;
    wishlist = null;
    myWishlistsToken = null;
    void loadMyWishlists();
  }

  function selectWishlist(id: string) {
    listId = id;
    wishlist = null;
    void loadWishlist();
  }
</script>

<svelte:head>
  <title>{$_("tgapp.pageTitle")}</title>
</svelte:head>

<div class="mx-auto flex max-w-5xl flex-col gap-4 px-4 py-6">
  {#if loading}
    <div class="text-muted-foreground flex items-center gap-3">
      <Loader2Icon class="h-5 w-5 animate-spin" />
      <span>{$_("tgapp.loadingWishlist")}</span>
    </div>
  {:else if error}
    <Card class="border-destructive">
      <CardContent class="text-destructive pt-6">
        {error}
      </CardContent>
    </Card>
  {:else if !listId}
    <TgAppHome
      isLoggedIn={Boolean($authStore.token)}
      userName={$authStore.user?.displayName || ""}
      {telegramLoginAvailable}
      {telegramLoginLoading}
      {creatingWishlist}
      wishlists={myWishlists}
      wishlistsLoading={myWishlistsLoading}
      wishlistsError={myWishlistsError}
      onLogin={() => loginWithTelegram(true)}
      onCreateWishlist={createMyWishlist}
      onSelectWishlist={selectWishlist}
    />
  {:else if wishlist}
    <WishlistHeader
      {wishlist}
      isOwner={isOwner()}
      {deletingWishlist}
      {telegramLoginAvailable}
      {telegramLoginLoading}
      isLoggedIn={Boolean($authStore.token)}
      onGoBack={goBackToList}
      onLogin={() => loginWithTelegram(true)}
      onShare={shareWishlistToTelegram}
      onDelete={deleteWishlist}
      onSave={saveWishlistEdits}
    />

    <div class="grid gap-4 md:grid-cols-2">
      {#each wishlist.items || [] as item}
        <WishlistItemCard
          {item}
          listId={listId}
          isOwner={isOwner()}
          isBookingRevealed={isBookingRevealed(item.id)}
          hasBookingToken={Boolean(getBookingToken(listId, item.id))}
          {defaultName}
          deleting={deletingItemId === item.id}
          onRevealBooking={() => revealBooking(item.id)}
          onBook={(anon, msg) => bookItem(item, anon, msg)}
          onUnbook={() => unbookItem(item)}
          onSave={(name, desc) => saveItem(item, name, desc)}
          onDelete={() => deleteItem(item)}
        />
      {/each}
    </div>

    {#if isOwner()}
      <AddItemForm onAdd={addItem} />
    {/if}

    <div class="text-muted-foreground mt-2 flex justify-center text-sm">
      <a
        class="hover:text-foreground inline-flex items-center gap-2"
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
