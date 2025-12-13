<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { env } from "$env/dynamic/public";
  import type { components } from "$lib/api/generated/wishlist-api";
  import { wishlistApi } from "$lib/api/wishlist-client";
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "$lib/components/ui/card";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "$lib/components/ui/dropdown-menu";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Input } from "$lib/components/ui/input";
  import {
    saveBookingToken,
    getBookingToken,
    removeBookingToken,
  } from "$lib/utils/booking-storage";
  import { showSuccessAlert, showInfoAlert } from "$lib/utils/alerts";
  import { authStore } from "$lib/stores/auth";
  import { exchangeTelegramInitData } from "$lib/auth";
  import WiliLogo from "$lib/components/WiliLogo.svelte";
  import {
    Loader2Icon,
    LinkIcon,
    CheckIcon,
    XIcon,
    ShieldOffIcon,
    EditIcon,
    SaveIcon,
    PlusIcon,
    SendIcon,
    TrashIcon,
    ArrowLeftIcon,
    EllipsisVerticalIcon,
  } from "@lucide/svelte";

  type Wishlist = components["schemas"]["Wishlist"];
  type WishlistItem = components["schemas"]["WishlistItem"];

  let wishlist: Wishlist | null = $state(null);
  let listId: string | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  let bookingItemId: string | null = $state(null);
  let anonymous = $state(false);
  let message = $state("");
  let defaultName: string | null = $state(null);
  let telegramAuthAttempted = $state(false);
  let telegramLoginLoading = $state(false);
  let telegramLoginAvailable = $state(false);
  let telegramInitData = $state("");
  let creatingWishlist = $state(false);
  let editingWishlist = $state(false);
  let editTitle = $state("");
  let editDescription = $state("");
  let addingItem = $state(false);
  let newItemName = $state("");
  let newItemDescription = $state("");
  let editingItemId = $state<string | null>(null);
  let editItemName = $state("");
  let editItemDescription = $state("");
  let savingItem = $state(false);
  let deletingItemId = $state<string | null>(null);
  let deletingWishlist = $state(false);
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
      editTitle = wishlist?.title || "";
      editDescription = wishlist?.description || "";
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

  function extractUserFromHash(): {
    first_name?: string;
    last_name?: string;
    username?: string;
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
      anonymous = false;
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
    }
    if (!defaultName) {
      anonymous = true;
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

  function startEditingItem(item: WishlistItem) {
    editingItemId = item.id;
    editItemName = item.data?.name || "";
    editItemDescription = item.data?.description || "";
  }

  function cancelEditingItem() {
    editingItemId = null;
    editItemName = "";
    editItemDescription = "";
  }

  async function saveItem(item: WishlistItem) {
    if (!wishlist || !$authStore.token) return;
    const name = editItemName.trim();
    const description = editItemDescription.trim();
    if (!name) return;

    savingItem = true;
    try {
      await wishlistApi.updateWishlistItem(
        wishlist.id,
        item.id,
        {
          type: item.type,
          data: {
            name,
            ...(description ? { description } : {}),
          },
        } as any,
        $authStore.token
      );
      await loadWishlist();
      cancelEditingItem();
      showSuccessAlert($_("common.save"), undefined, "bottom-center");
    } catch (e) {
      console.warn("update item failed", e);
      showInfoAlert($_("items.failedToUpdate"), undefined, "bottom-center");
    } finally {
      savingItem = false;
    }
  }

  async function deleteItem(item: WishlistItem) {
    if (!wishlist || !$authStore.token) return;
    deletingItemId = item.id;
    try {
      await wishlistApi.deleteWishlistItem(wishlist.id, item.id, $authStore.token);
      await loadWishlist();
      if (editingItemId === item.id) cancelEditingItem();
      showSuccessAlert($_("common.delete"), undefined, "bottom-center");
    } catch (e) {
      console.warn("delete item failed", e);
      showInfoAlert($_("items.failedToDelete"), undefined, "bottom-center");
    } finally {
      deletingItemId = null;
    }
  }

  async function saveWishlistEdits() {
    if (!wishlist || !$authStore.token) return;
    try {
      wishlist = await wishlistApi.updateWishlist(
        wishlist.id,
        { title: editTitle, description: editDescription },
        $authStore.token
      );
      editingWishlist = false;
      showSuccessAlert($_("common.save"), undefined, "bottom-center");
    } catch (e) {
      console.warn("wishlist update failed", e);
      showInfoAlert($_("wishlists.failedToUpdate"), undefined, "bottom-center");
    }
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
      editingWishlist = false;
      addingItem = false;
      editingItemId = null;
      bookingItemId = null;
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

  async function addItem() {
    if (!wishlist || !$authStore.token) return;
    const name = newItemName.trim();
    const description = newItemDescription.trim();
    if (!name) return;
    try {
      await wishlistApi.addWishlistItem(
        wishlist.id,
        {
          type: "general",
          data: {
            name,
            ...(description ? { description } : {}),
          },
        } as any,
        $authStore.token
      );
      newItemName = "";
      newItemDescription = "";
      addingItem = false;
      await loadWishlist();
      showSuccessAlert($_("items.addedSuccessfully"), undefined, "bottom-center");
    } catch (e) {
      console.warn("add item failed", e);
      showInfoAlert($_("items.failedToAdd"), undefined, "bottom-center");
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
      editTitle = wl.title;
      editDescription = wl.description || "";
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
    editingWishlist = false;
    addingItem = false;
    editingItemId = null;
    bookingItemId = null;
    myWishlistsToken = null;
    void loadMyWishlists();
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
    <div class="flex flex-col items-center gap-6 py-10 text-center">
      <div class="flex flex-col items-center gap-3">
        <WiliLogo className="h-14 w-auto" />
        <p class="text-muted-foreground max-w-sm text-sm leading-relaxed">
          {#if $authStore.token}
            {$_("tgapp.homeCreatePrompt", { values: { name: $authStore.user?.displayName || "" } })}
          {:else if telegramLoginAvailable}
            {$_("tgapp.homeLoginPrompt")}
          {:else}
            {$_("tgapp.openFromChat")}
          {/if}
        </p>
      </div>

      <div class="w-full max-w-sm space-y-2">
        {#if !$authStore.token && telegramLoginAvailable}
          <Button
            disabled={telegramLoginLoading}
            onclick={() => loginWithTelegram(true)}
            class="w-full"
          >
            {telegramLoginLoading ? $_("common.loading") : $_("auth.loginWithTelegram")}
          </Button>
        {/if}
        {#if $authStore.token}
          <Button disabled={creatingWishlist} onclick={createMyWishlist} class="w-full">
            {creatingWishlist ? $_("common.loading") : $_("tgapp.createWishlist")}
          </Button>
        {/if}
      </div>

      {#if $authStore.token}
        <div class="w-full max-w-sm space-y-2 text-left">
          {#if myWishlistsLoading}
            <div class="text-muted-foreground text-center text-sm">{$_("common.loading")}</div>
          {:else if myWishlistsError}
            <div class="text-muted-foreground text-center text-sm">{myWishlistsError}</div>
          {:else if myWishlists.length > 0}
            {#each myWishlists.slice(0, 8) as w}
              <button
                type="button"
                class="border-border hover:bg-muted flex w-full items-center justify-between rounded-lg border px-4 py-3 text-left"
                onclick={() => {
                  listId = w.id;
                  wishlist = null;
                  void loadWishlist();
                }}
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
  {:else if wishlist}
    <div class="flex items-center justify-between gap-2">
      <Button variant="ghost" onclick={goBackToList} class="gap-2">
        <ArrowLeftIcon class="h-4 w-4" />
        {$_("common.back")}
      </Button>
      {#if telegramLoginAvailable && !$authStore.token}
        <Button
          variant="outline"
          disabled={telegramLoginLoading}
          onclick={() => loginWithTelegram(true)}
        >
          {telegramLoginLoading ? $_("common.loading") : $_("auth.loginWithTelegram")}
        </Button>
      {/if}
    </div>
    <div class="flex w-full items-center justify-center gap-2">
      <Button variant="outline" onclick={shareWishlistToTelegram} class="gap-2">
        <SendIcon class="h-4 w-4" />
        {$_("wishlists.shareToTelegram")}
      </Button>
      {#if isOwner()}
        <Button variant="outline" onclick={() => (addingItem = !addingItem)} class="gap-2">
          <PlusIcon class="h-4 w-4" />
          {$_("wishlists.addItem")}
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger>
            <Button variant="outline" size="icon">
              <EllipsisVerticalIcon class="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onclick={() => (editingWishlist = true)}
              disabled={editingWishlist}
              class="gap-2"
            >
              <EditIcon class="h-4 w-4" />
              {$_("common.edit")}
            </DropdownMenuItem>
            <DropdownMenuItem
              onclick={deleteWishlist}
              disabled={deletingWishlist}
              class="text-destructive gap-2"
            >
              <TrashIcon class="h-4 w-4" />
              {$_("wishlists.deleteWishlist")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      {/if}
    </div>
    <div class="p-0">
      <div class="flex items-start justify-between gap-3">
        <div class="space-y-2">
          {#if editingWishlist && isOwner()}
            <div class="space-y-3">
              <Input bind:value={editTitle} class="text-lg font-semibold" />
              <Textarea bind:value={editDescription} />
              <div class="flex gap-2">
                <Button onclick={saveWishlistEdits} class="gap-2">
                  <SaveIcon class="h-4 w-4" />
                  {$_("common.save")}
                </Button>
                <Button
                  variant="outline"
                  onclick={() => {
                    editingWishlist = false;
                    editTitle = wishlist?.title || "";
                    editDescription = wishlist?.description || "";
                  }}
                >
                  {$_("common.cancel")}
                </Button>
              </div>
            </div>
          {:else}
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
          {/if}
        </div>
      </div>
    </div>

    {#if isOwner() && addingItem}
      <Card>
        <CardHeader>
          <CardTitle>{$_("wishlists.addNewItem")}</CardTitle>
        </CardHeader>
        <CardContent class="flex flex-col gap-3">
          <Input placeholder={$_("items.namePlaceholder")} bind:value={newItemName} />
          <Textarea
            placeholder={$_("items.descriptionPlaceholder")}
            bind:value={newItemDescription}
          />
          <div class="flex gap-2">
            <Button onclick={addItem} disabled={!newItemName.trim()} class="gap-2">
              <PlusIcon class="h-4 w-4" />
              {$_("common.add")}
            </Button>
            <Button variant="outline" onclick={() => (addingItem = false)}>
              {$_("common.cancel")}
            </Button>
          </div>
        </CardContent>
      </Card>
    {/if}

    <div class="grid gap-4 md:grid-cols-2">
      {#each wishlist.items || [] as item}
        <Card class="h-full">
          <CardHeader class="relative">
            <CardTitle class="line-clamp-2"
              >{item.data?.name || $_("wishlists.untitledItem")}</CardTitle
            >
            {#if item.data?.description}
              <CardDescription class="text-muted-foreground text-sm">
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
            {#if isOwner()}
              <div class="absolute right-4 top-4 flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 w-8 p-0"
                  disabled={deletingItemId === item.id}
                  onclick={() => startEditingItem(item)}
                >
                  <EditIcon class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  class="text-destructive h-8 w-8 p-0"
                  disabled={deletingItemId === item.id}
                  onclick={() => deleteItem(item)}
                >
                  <TrashIcon class="h-4 w-4" />
                </Button>
              </div>
            {/if}
          </CardHeader>
          <CardContent class="flex flex-col gap-3">
            {#if editingItemId === item.id && isOwner()}
              <div class="space-y-3">
                <Input placeholder={$_("items.namePlaceholder")} bind:value={editItemName} />
                <Textarea
                  placeholder={$_("items.descriptionPlaceholder")}
                  bind:value={editItemDescription}
                />
                <div class="flex gap-2">
                  <Button
                    class="gap-2"
                    disabled={savingItem || !editItemName.trim()}
                    onclick={() => saveItem(item)}
                  >
                    <SaveIcon class="h-4 w-4" />
                    {savingItem ? $_("common.loading") : $_("common.save")}
                  </Button>
                  <Button variant="outline" onclick={cancelEditingItem}>
                    <XIcon class="h-4 w-4" />
                    {$_("common.cancel")}
                  </Button>
                </div>
              </div>
            {:else if item.booking}
              <div
                class="flex flex-wrap items-center justify-between gap-2 px-1 py-1 text-sm text-green-700 dark:text-green-300"
              >
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-2">
                    <CheckIcon class="h-4 w-4 shrink-0" />
                    <span>{$_("tgapp.alreadyBooked")}</span>
                  </div>
                  {#if item.booking.bookerName}
                    <div class="text-sm text-green-700/80 dark:text-green-200/80">
                      {item.booking.bookerName}
                    </div>
                  {/if}
                </div>
                {#if !isOwner() && listId && getBookingToken(listId, item.id)}
                  <Button size="sm" variant="outline" onclick={() => unbook(item)}>
                    <XIcon class="mr-2 h-4 w-4" />
                    {$_("common.cancel")}
                  </Button>
                {/if}
              </div>
            {:else if !isOwner() && bookingItemId === item.id}
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
            {:else if !isOwner()}
              <Button class="w-full gap-2" onclick={() => startBooking(item.id)}>
                <ShieldOffIcon class="h-4 w-4" />
                {$_("items.bookItem")}
              </Button>
            {/if}
          </CardContent>
        </Card>
      {/each}
    </div>

    <div class="text-muted-foreground mt-6 flex justify-center text-sm">
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
