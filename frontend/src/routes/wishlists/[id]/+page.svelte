<script lang="ts">
  import { onMount } from "svelte";
  import { env } from "$env/dynamic/public";
  import { page } from "$app/state";
  import { authStore } from "$lib/stores/auth";
  import { goto } from "$app/navigation";
  import { Button, buttonVariants } from "$lib/components/ui/button";
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Separator } from "$lib/components/ui/separator";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
  import {
    ArrowLeftIcon,
    EditIcon,
    TrashIcon,
    PlusIcon,
    SaveIcon,
    XIcon,
    MoreVerticalIcon,
    ShareIcon,
    SendIcon,
    BookOpenIcon,
    UserIcon,
    EyeIcon,
  } from "@lucide/svelte";
  import { wishlistApi } from "$lib/api/wishlist-client";
  import type { components } from "$lib/api/generated/wishlist-api";
  import { saveBookingToken, getBookingToken, removeBookingToken, hasBookingToken } from "$lib/utils/booking-storage";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import { showSuccessAlert, showInfoAlert } from "$lib/utils/alerts";
  import { _ } from "svelte-i18n";
  import T from "$lib/components/T.svelte";
  import {
    ITEM_NAME_MAX_LENGTH,
    ITEM_DESCRIPTION_MAX_LENGTH,
    WISHLIST_TITLE_MAX_LENGTH,
    WISHLIST_DESCRIPTION_MAX_LENGTH,
    validation,
  } from "$lib/api/validation-constants";
  type Wishlist = components["schemas"]["Wishlist"];
  type WishlistItem = components["schemas"]["WishlistItem"];
  type ItemBooking = components["schemas"]["ItemBooking"];

  let wishlist: Wishlist | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let editing = $state(false);
  let editForm = $state({
    title: "",
    description: "",
  });

  // New item form
  let addingItem = $state(false);
  let newItem = $state({
    name: "",
    description: "",
    type: "general" as const,
  });

  // Item editing
  let editingItemId = $state<string | null>(null);
  let editItemForm = $state({
    name: "",
    description: "",
  });

  // Booking
  let bookingItemId = $state<string | null>(null);
  let bookingForm = $state({
    bookerName: "",
    message: "",
  });

  // Revealed bookings for owner (to preserve surprise)
  type RevealLevel = "none" | "status" | "details";
  let revealedBookings = $state<Map<string, RevealLevel>>(new Map());

  const wishlistId = $derived(page.params.id);

  // SEO/OG derived values
  const defaultOgTitle = "Wishlist — Wili";
  const defaultOgDescription = "View this wishlist on Wili.";
  function summarize(text: string): string {
    const s = (text || "").toString().replace(/\s+/g, " ").trim();
    return s.length > 160 ? s.slice(0, 160) + "…" : s;
  }
  const ogTitle = $derived(
    wishlist && wishlist.title ? `${wishlist.title} — Wili` : defaultOgTitle
  );
  const ogDescription = $derived(
    wishlist && wishlist.description ? summarize(wishlist.description) : defaultOgDescription
  );

  async function loadWishlist() {
    if (!wishlistId) return;

    try {
      loading = true;
      wishlist = await wishlistApi.getWishlist(wishlistId, $authStore.token);
      editForm = {
        title: wishlist.title,
        description: wishlist.description || "",
      };
    } catch (err) {
      if (err instanceof Error && err.message.includes("404")) {
        error = $_("wishlists.wishlistNotFound");
      } else {
        error = err instanceof Error ? err.message : $_("wishlists.failedToLoad");
      }
      console.error("Error loading wishlist:", err);
    } finally {
      loading = false;
    }
  }

  async function saveWishlist() {
    if (!$authStore.token || !wishlistId || !wishlist) return;

    // Client-side validation using generated constants
    const titleErrorKey = validation.getWishlistTitleErrorKey(editForm.title);
    if (titleErrorKey) {
      error = $_(titleErrorKey);
      return;
    }

    const descriptionErrorKey = validation.getWishlistDescriptionErrorKey(editForm.description);
    if (descriptionErrorKey) {
      error = $_(descriptionErrorKey);
      return;
    }

    try {
      wishlist = await wishlistApi.updateWishlist(wishlistId, editForm, $authStore.token);
      editing = false;
    } catch (err) {
      error = err instanceof Error ? err.message : $_("wishlists.failedToUpdate");
      console.error("Error updating wishlist:", err);
    }
  }

  async function deleteWishlist() {
    if (!$authStore.token || !wishlistId) return;
    if (!confirm($_("wishlists.confirmDelete"))) return;

    try {
      await wishlistApi.deleteWishlist(wishlistId, $authStore.token);
      goto("/wishlists");
    } catch (err) {
      error = err instanceof Error ? err.message : $_("wishlists.failedToDelete");
      console.error("Error deleting wishlist:", err);
    }
  }

  async function addItem() {
    if (!$authStore.token || !wishlistId) return;

    // Client-side validation using generated constants
    const nameErrorKey = validation.getItemNameErrorKey(newItem.name);
    if (nameErrorKey) {
      error = $_(nameErrorKey);
      return;
    }

    const descriptionErrorKey = validation.getItemDescriptionErrorKey(newItem.description);
    if (descriptionErrorKey) {
      error = $_(descriptionErrorKey);
      return;
    }

    try {
      const itemData: any = {
        type: newItem.type,
        data: {
          name: newItem.name.trim(),
        },
      };

      if (newItem.description && newItem.description.trim())
        itemData.data.description = newItem.description.trim();

      await wishlistApi.addWishlistItem(wishlistId, itemData, $authStore.token);

      // Reload wishlist to get updated items
      await loadWishlist();
      addingItem = false;
      newItem = { name: "", description: "", type: "general" };
    } catch (err) {
      error = err instanceof Error ? err.message : $_("items.failedToAdd");
      console.error("Error adding item:", err);
    }
  }

  async function deleteItem(itemId: string) {
    if (!$authStore.token || !wishlistId) return;
    if (!confirm($_("items.confirmDelete"))) return;

    try {
      await wishlistApi.deleteWishlistItem(wishlistId, itemId, $authStore.token);
      // Reload wishlist to get updated items
      await loadWishlist();
    } catch (err) {
      error = err instanceof Error ? err.message : $_("items.failedToDelete");
      console.error("Error deleting item:", err);
    }
  }

  function startEditingItem(item: WishlistItem) {
    editingItemId = item.id;
    editItemForm = {
      name: item.data?.name || "",
      description: item.data?.description || "",
    };
  }

  function cancelEditingItem() {
    editingItemId = null;
    editItemForm = {
      name: "",
      description: "",
    };
  }

  async function saveItemEdit(itemId: string, currentItem: WishlistItem) {
    if (!$authStore.token || !wishlistId) return;

    // Client-side validation using generated constants
    const nameErrorKey = validation.getItemNameErrorKey(editItemForm.name);
    if (nameErrorKey) {
      error = $_(nameErrorKey);
      return;
    }

    const descriptionErrorKey = validation.getItemDescriptionErrorKey(editItemForm.description);
    if (descriptionErrorKey) {
      error = $_(descriptionErrorKey);
      return;
    }

    try {
      const updateData: any = {
        type: currentItem.type,
        data: {
          name: editItemForm.name.trim(),
        },
      };

      if (editItemForm.description && editItemForm.description.trim()) {
        updateData.data.description = editItemForm.description.trim();
      }

      await wishlistApi.updateWishlistItem(wishlistId, itemId, updateData, $authStore.token);

      // Reload wishlist to get updated items
      await loadWishlist();
      cancelEditingItem();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to update item";
      console.error("Error updating item:", err);
    }
  }

  const telegramBotUsername = env.PUBLIC_TELEGRAM_BOT_USERNAME;
  const telegramWebAppUrl = env.PUBLIC_TELEGRAM_WEBAPP_URL;

  async function shareWishlist() {
    if (!wishlist) return;

    const shareUrl = `${window.location.origin}/wishlists/${wishlist.id}`;

    try {
      await navigator.clipboard.writeText(shareUrl);
      showSuccessAlert($_("wishlists.shareLinkCopied"), undefined, "top-right");
    } catch (err) {
      // Fallback for older browsers or when clipboard API fails
      console.error("Failed to copy to clipboard:", err);
      showInfoAlert($_("wishlists.shareLink"), shareUrl, "top-right");
    }
  }

  function shareWishlistToTelegram() {
    if (!wishlist) return;
    const startParam = `list_${wishlist.id}`;

    if (telegramBotUsername) {
      const link = `https://t.me/${telegramBotUsername}?start=${startParam}`;
      window.open(link, "_blank", "noopener");
      return;
    }

    // Bot not configured: log and noop
    console.warn("Telegram bot username is not configured. Set PUBLIC_TELEGRAM_BOT_USERNAME.");
  }

  // Booking functions
  function startBooking(itemId: string) {
    bookingItemId = itemId;
    bookingForm = {
      bookerName: "",
      message: "",
    };
  }

  function cancelBooking() {
    bookingItemId = null;
    bookingForm = {
      bookerName: "",
      message: "",
    };
  }

  async function bookItem() {
    if (!bookingItemId || !wishlistId) return;

    try {
      const bookingData = {
        bookerName: bookingForm.bookerName.trim() || undefined,
        message: bookingForm.message.trim() || undefined,
      };

      const response = await wishlistApi.bookItem(wishlistId, bookingItemId, bookingData);
      saveBookingToken(wishlistId, bookingItemId, response.cancellationToken);
      await loadWishlist();
      cancelBooking();
      
      showSuccessAlert(
        $_("items.bookedSuccessfully"),
        undefined,
        "top-right"
      );
    } catch (err) {
      if (err instanceof Error && err.message.includes("already booked")) {
        error = $_("items.alreadyBooked");
      } else {
        error = err instanceof Error ? err.message : $_("items.failedToBook");
      }
      console.error("Error booking item:", err);
    }
  }

  async function unbookItemByOwner(item: WishlistItem) {
    if (!item.booking || !wishlistId || !$authStore.token) return;
    if (!confirm($_("items.confirmUnbook"))) return;

    try {
      await wishlistApi.unbookItemByOwner(wishlistId, item.id, item.booking.bookingId, $authStore.token);
      await loadWishlist();
      revealedBookings.delete(item.id);
      revealedBookings = new Map(revealedBookings);
      
      showSuccessAlert($_("items.unbookedSuccessfully"), undefined, "top-right");
    } catch (err) {
      error = err instanceof Error ? err.message : $_("items.failedToUnbook");
      console.error("Error unbooking item:", err);
    }
  }

  async function unbookItemByToken(item: WishlistItem) {
    if (!wishlistId) return;
    
    const cancellationToken = getBookingToken(wishlistId, item.id);
    if (!cancellationToken) {
      error = $_("items.noCancellationToken");
      return;
    }

    if (!confirm($_("items.confirmCancelBooking"))) return;

    try {
      await wishlistApi.unbookItemByToken(wishlistId, item.id, cancellationToken);
      removeBookingToken(wishlistId, item.id);
      await loadWishlist();
      
      showSuccessAlert($_("items.bookingCancelled"), undefined, "top-right");
    } catch (err) {
      error = err instanceof Error ? err.message : $_("items.failedToCancelBooking");
      console.error("Error cancelling booking:", err);
    }
  }

  function revealBookingStatus(itemId: string) {
    revealedBookings.set(itemId, "status");
    revealedBookings = new Map(revealedBookings);
  }

  function revealBookingDetails(itemId: string) {
    revealedBookings.set(itemId, "details");
    revealedBookings = new Map(revealedBookings);
  }

  function getRevealLevel(itemId: string): RevealLevel {
    return revealedBookings.get(itemId) || "none";
  }

  function revealAllBookings(level: RevealLevel) {
    if (!wishlist?.items) return;
    wishlist.items.forEach(item => {
      revealedBookings.set(item.id, level);
    });
    revealedBookings = new Map(revealedBookings);
  }

  onMount(() => {
    loadWishlist();
  });
</script>

<svelte:head>
  <title>{ogTitle}</title>
  <meta name="description" content={ogDescription} />
  <meta property="og:type" content="website" />
  <meta property="og:title" content={ogTitle} />
  <meta property="og:description" content={ogDescription} />
  <meta property="og:url" content={page.url.href} />
  <meta property="og:image" content="/favicon.ico" />
  <meta name="twitter:card" content="summary" />
  <meta name="twitter:title" content={ogTitle} />
  <meta name="twitter:description" content={ogDescription} />
  <meta name="twitter:image" content="/favicon.ico" />
</svelte:head>

<div class="container mx-auto px-4 py-8">
  <!-- Back button -->
  {#if $authStore.token}
    <Button variant="ghost" onclick={() => goto("/wishlists")} class="mb-6 gap-2">
      <ArrowLeftIcon class="h-4 w-4" />
      <T key="wishlists.backToWishlists" fallback="Back to Wishlists" />
    </Button>
  {/if}

  {#if loading}
    <div class="space-y-4">
      <div class="h-8 w-1/3 animate-pulse rounded bg-gray-200 dark:bg-black"></div>
      <div class="h-4 w-1/2 animate-pulse rounded bg-gray-100 dark:bg-black"></div>
      <div class="h-32 animate-pulse rounded bg-gray-100 dark:bg-black"></div>
    </div>
  {:else if error}
    <Card class="border-destructive">
      <CardContent class="pt-6">
        <p class="text-destructive">{error}</p>
        <Button onclick={loadWishlist} variant="outline" class="mt-4">
          <T key="common.tryAgain" fallback="Try Again" />
        </Button>
      </CardContent>
    </Card>
  {:else if wishlist}
    <!-- Wishlist Header -->
    <div class="mb-8 flex items-start justify-between">
      <div class="flex-1">
        <!-- Show owner info for anonymous viewers -->
        {#if !$authStore.token}
          <div class="bg-muted mb-6 rounded-lg p-3">
            <p class="text-muted-foreground text-sm">
              <T key="auth.anonymousViewing" fallback="Anonymous viewing" />
              <a href="/" class="text-primary hover:underline"
                ><T key="nav.login" fallback="Login" /></a
              >
              <T key="auth.loginToCreateWishlists" fallback="Login to create wishlists" />
            </p>
          </div>
        {/if}
        {#if editing}
          <div class="space-y-4">
            <Input
              bind:value={editForm.title}
              placeholder={$_("wishlists.titlePlaceholder")}
              class="text-2xl font-bold {!validation.isValidWishlistTitle(editForm.title) &&
              editForm.title.length > 0
                ? 'border-destructive'
                : ''}"
              required
              maxlength={WISHLIST_TITLE_MAX_LENGTH}
            />
            {#if validation.getWishlistTitleErrorKey(editForm.title)}
              <p class="text-destructive text-sm">
                {$_(validation.getWishlistTitleErrorKey(editForm.title))}
              </p>
            {/if}
            <Textarea
              bind:value={editForm.description}
              placeholder={$_("wishlists.descriptionPlaceholder")}
              maxlength={WISHLIST_DESCRIPTION_MAX_LENGTH}
            />
            <p class="text-muted-foreground text-sm">
              {editForm.description.length || 0} / {WISHLIST_DESCRIPTION_MAX_LENGTH}
            </p>
            {#if validation.getWishlistDescriptionErrorKey(editForm.description)}
              <p class="text-destructive text-sm">
                {$_(validation.getWishlistDescriptionErrorKey(editForm.description))}
              </p>
            {/if}
            <div class="flex items-center gap-3 pt-2">
              <Button
                onclick={saveWishlist}
                class="gap-2"
                disabled={!validation.isValidWishlistTitle(editForm.title) ||
                  !validation.isValidWishlistDescription(editForm.description)}
              >
                <SaveIcon class="h-4 w-4" />
                {$_("common.save")}
              </Button>
              <Button variant="outline" onclick={() => (editing = false)} class="gap-2">
                <XIcon class="h-4 w-4" />
                {$_("common.cancel")}
              </Button>
            </div>
          </div>
        {:else}
          <h1 class="mb-2 text-3xl font-bold">{wishlist.title}</h1>
          {#if wishlist.description}
            <div class="text-muted-foreground mb-4">
              <ExpandableText
                content={wishlist.description}
                allowYandexMarket={false}
                maxHeight={200}
                className="text-muted-foreground"
              />
            </div>
          {/if}
          {#if wishlist.items?.length !== 0}
            <div class="flex items-center gap-2">
              <span class="text-muted-foreground text-sm">
                {wishlist.items?.length || 0}
                <T key="wishlists.items" fallback="items" />
              </span>
            </div>
          {/if}
        {/if}
      </div>

      <!-- Only show edit controls if user is authenticated and owns the wishlist -->
      {#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id && !editing}
        <div class="flex items-center gap-2">
          <Button variant="outline" onclick={() => (editing = true)} class="gap-2">
            <EditIcon class="h-4 w-4" />
            <T key="common.edit" fallback="Edit" />
          </Button>
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <Button variant="ghost" size="sm" class="h-9 w-9 p-0">
                <MoreVerticalIcon class="h-4 w-4" />
              </Button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              {#if telegramBotUsername}
                <DropdownMenu.Item onclick={shareWishlistToTelegram}>
                  <SendIcon class="mr-2 h-4 w-4" />
                  <T key="wishlists.shareToTelegram" fallback="Share to Telegram" />
                </DropdownMenu.Item>
              {/if}
              <DropdownMenu.Item onclick={shareWishlist}>
                <ShareIcon class="mr-2 h-4 w-4" />
                <T key="wishlists.shareLink" fallback="Share Link" />
              </DropdownMenu.Item>
              <DropdownMenu.Separator />
              <DropdownMenu.Item onclick={() => revealAllBookings("status")}>
                <EyeIcon class="mr-2 h-4 w-4" />
                <T key="items.revealAllStatus" fallback="Reveal all: status" />
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={() => revealAllBookings("details")}>
                <EyeIcon class="mr-2 h-4 w-4" />
                <T key="items.revealAllDetails" fallback="Reveal all: details" />
              </DropdownMenu.Item>
              <DropdownMenu.Separator />
              <DropdownMenu.Item
                onclick={deleteWishlist}
                class="text-destructive focus:text-destructive"
              >
                <TrashIcon class="mr-2 h-4 w-4" />
                <T key="wishlists.deleteWishlist" fallback="Delete Wishlist" />
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      {/if}
    </div>

    <Separator class="mb-8" />

    <!-- Add Item Section - Only show if user owns the wishlist -->
    {#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id}
      {#if addingItem}
        <Card class="mb-6">
          <CardHeader>
            <CardTitle><T key="wishlists.addNewItem" fallback="Add New Item" /></CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <Input
              bind:value={newItem.name}
              placeholder={$_("wishlists.itemNamePlaceholder")}
              required
              maxlength={ITEM_NAME_MAX_LENGTH}
              class={!validation.isValidItemName(newItem.name) && newItem.name.length > 0
                ? "border-destructive"
                : ""}
            />
            {#if validation.getItemNameErrorKey(newItem.name)}
              <p class="text-destructive text-sm">
                {$_(validation.getItemNameErrorKey(newItem.name))}
              </p>
            {/if}
            <Textarea
              bind:value={newItem.description}
              placeholder={$_("wishlists.itemDescriptionPlaceholder")}
              maxlength={ITEM_DESCRIPTION_MAX_LENGTH}
              rows={5}
              class="min-h-[100px]"
            />
            <p class="text-muted-foreground text-sm">
              {newItem.description.length || 0} / {ITEM_DESCRIPTION_MAX_LENGTH}
            </p>
            {#if validation.getItemDescriptionErrorKey(newItem.description)}
              <p class="text-destructive text-sm">
                {$_(validation.getItemDescriptionErrorKey(newItem.description))}
              </p>
            {/if}
            <div class="flex gap-2">
              <Button
                onclick={addItem}
                class="gap-2"
                disabled={!validation.isValidItemName(newItem.name) ||
                  !validation.isValidItemDescription(newItem.description)}
              >
                <SaveIcon class="h-4 w-4" />
                <T key="wishlists.addItem" fallback="Add Item" />
              </Button>
              <Button variant="outline" onclick={() => (addingItem = false)} class="gap-2">
                <XIcon class="h-4 w-4" />
                <T key="common.cancel" fallback="Cancel" />
              </Button>
            </div>
          </CardContent>
        </Card>
      {:else}
        <Button onclick={() => (addingItem = true)} class="mb-6 gap-2">
          <PlusIcon class="h-4 w-4" />
          <T key="wishlists.addItem" fallback="Add Item" />
        </Button>
      {/if}
    {/if}

    <!-- Items List -->
    {#if wishlist.items && wishlist.items.length > 0}
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {#each wishlist.items as item}
          <Card class="min-h-[100px] transition-shadow hover:shadow-md">
            {#if editingItemId === item.id}
              <!-- Edit mode -->
              <CardHeader>
                <CardTitle><T key="wishlists.editItem" fallback="Edit Item" /></CardTitle>
              </CardHeader>
              <CardContent class="space-y-4">
                <Input
                  bind:value={editItemForm.name}
                  placeholder="Item name"
                  required
                  maxlength={ITEM_NAME_MAX_LENGTH}
                  class={!validation.isValidItemName(editItemForm.name) &&
                  editItemForm.name.length > 0
                    ? "border-destructive"
                    : ""}
                />
                {#if validation.getItemNameErrorKey(editItemForm.name)}
                  <p class="text-destructive text-sm">
                    {$_(validation.getItemNameErrorKey(editItemForm.name))}
                  </p>
                {/if}
                <Textarea
                  bind:value={editItemForm.description}
                  placeholder="Description (optional) - Markdown supported. URLs will become badges!"
                  maxlength={ITEM_DESCRIPTION_MAX_LENGTH}
                  rows={10}
                  class="min-h-[100px]"
                />
                <p class="text-muted-foreground text-sm">
                  {editItemForm.description.length || 0} / {ITEM_DESCRIPTION_MAX_LENGTH}
                </p>
                {#if validation.getItemDescriptionErrorKey(editItemForm.description)}
                  <p class="text-destructive text-sm">
                    {$_(validation.getItemDescriptionErrorKey(editItemForm.description))}
                  </p>
                {/if}
                <div class="flex gap-2">
                  <Button
                    onclick={() => saveItemEdit(item.id, item)}
                    class="gap-2"
                    size="sm"
                    disabled={!validation.isValidItemName(editItemForm.name) ||
                      !validation.isValidItemDescription(editItemForm.description)}
                  >
                    <SaveIcon class="h-4 w-4" />
                    <T key="common.save" fallback="Save" />
                  </Button>
                  <Button variant="outline" onclick={cancelEditingItem} class="gap-2" size="sm">
                    <XIcon class="h-4 w-4" />
                    <T key="common.cancel" fallback="Cancel" />
                  </Button>
                </div>
              </CardContent>
            {:else}
              <!-- Display mode -->
              <CardHeader class="relative">
                <CardTitle class="line-clamp-2 pr-10"
                  >{item.data?.name || $_("wishlists.untitledItem")}</CardTitle
                >
                {#if item.data?.description}
                  <CardDescription>
                    <ExpandableText
                      content={item.data.description}
                      className="mt-2 text-muted-foreground"
                      mobileFraction={0.7}
                      desktopFraction={0.55}
                      minMaxHeight={260}
                      maxMaxHeight={640}
                    />
                  </CardDescription>
                {/if}
                <!-- Actions menu in top-right corner -->
                {#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id}
                  <div class="absolute top-4 right-4">
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                          <MoreVerticalIcon class="h-4 w-4" />
                        </Button>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        <DropdownMenu.Item onclick={() => startEditingItem(item)}>
                          <EditIcon class="mr-2 h-4 w-4" />
                          <T key="common.edit" fallback="Edit" />
                        </DropdownMenu.Item>
                        {#if item.booking && getRevealLevel(item.id) !== "none"}
                          <DropdownMenu.Item
                            onclick={() => unbookItemByOwner(item)}
                            class="text-orange-600 focus:text-orange-600"
                          >
                            <BookOpenIcon class="mr-2 h-4 w-4" />
                            <T key="items.unbookItem" fallback="Unbook Item" />
                          </DropdownMenu.Item>
                        {/if}
                        <DropdownMenu.Item
                          onclick={() => deleteItem(item.id)}
                          class="text-destructive focus:text-destructive"
                        >
                          <TrashIcon class="mr-2 h-4 w-4" />
                          <T key="common.delete" fallback="Delete" />
                        </DropdownMenu.Item>
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  </div>
                {/if}
              </CardHeader>
              
              <!-- Booking information and actions -->
              <CardContent class="pt-0">
                {#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id}
                  {@const revealLevel = getRevealLevel(item.id)}
                  {#if revealLevel === "none"}
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger class={buttonVariants({ variant: "outline", size: "sm", class: "w-full gap-2" })}>
                        <EyeIcon class="h-4 w-4" />
                        <T key="items.revealBooking" fallback="Reveal booking info" />
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="center">
                        <DropdownMenu.Item onclick={() => revealBookingStatus(item.id)}>
                          <T key="items.showStatus" fallback="Show booking status" />
                        </DropdownMenu.Item>
                        <DropdownMenu.Item onclick={() => revealBookingDetails(item.id)}>
                          <T key="items.showDetails" fallback="Show full details" />
                        </DropdownMenu.Item>
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  {:else if revealLevel === "status"}
                    {#if item.booking}
                      <div class="rounded-lg bg-green-50 p-3 dark:bg-green-900/20">
                        <div class="flex items-center justify-between">
                          <div class="flex items-center gap-2 text-green-700 dark:text-green-300">
                            <BookOpenIcon class="h-4 w-4" />
                            <span class="text-sm font-medium">
                              <T key="items.itemBooked" fallback="Item is booked" />
                            </span>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            onclick={() => revealBookingDetails(item.id)}
                            class="h-7 gap-1 text-xs"
                          >
                            <EyeIcon class="h-3 w-3" />
                            <T key="items.showMore" fallback="Show more" />
                          </Button>
                        </div>
                      </div>
                    {:else}
                      <div class="rounded-lg bg-gray-50 p-3 dark:bg-gray-900/20">
                        <div class="flex items-center gap-2 text-gray-700 dark:text-gray-300">
                          <BookOpenIcon class="h-4 w-4" />
                          <span class="text-sm font-medium">
                            <T key="items.itemNotBooked" fallback="Not booked yet" />
                          </span>
                        </div>
                      </div>
                    {/if}
                  {:else if revealLevel === "details"}
                    {#if item.booking}
                      <div class="rounded-lg bg-green-50 p-3 dark:bg-green-900/20">
                        <div class="flex items-center gap-2 text-green-700 dark:text-green-300">
                          <BookOpenIcon class="h-4 w-4" />
                          <span class="text-sm font-medium">
                            <T key="items.bookedBy" fallback="Booked by" />
                            {item.booking.bookerName || $_("items.anonymous")}
                          </span>
                        </div>
                        {#if item.booking.message}
                          <p class="mt-1 text-sm text-green-600 dark:text-green-400">
                            "{item.booking.message}"
                          </p>
                        {/if}
                        <p class="mt-1 text-xs text-green-600 dark:text-green-400">
                          <T key="items.bookedOn" fallback="Booked on" />
                          {new Date(item.booking.bookedAt).toLocaleDateString()}
                        </p>
                      </div>
                    {:else}
                      <div class="rounded-lg bg-gray-50 p-3 dark:bg-gray-900/20">
                        <div class="flex items-center gap-2 text-gray-700 dark:text-gray-300">
                          <BookOpenIcon class="h-4 w-4" />
                          <span class="text-sm font-medium">
                            <T key="items.itemNotBooked" fallback="Not booked yet" />
                          </span>
                        </div>
                      </div>
                    {/if}
                  {/if}
                {:else}
                  {#if item.booking}
                    {@const hasToken = hasBookingToken(wishlistId, item.id)}
                    <div class="rounded-lg bg-green-50 p-3 dark:bg-green-900/20">
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2 text-green-700 dark:text-green-300">
                          <BookOpenIcon class="h-4 w-4" />
                          <span class="text-sm font-medium">
                            {#if hasToken}
                              <T key="items.youBookedThis" fallback="You booked this" />
                            {:else}
                              <T key="items.bookedBy" fallback="Booked by" />
                              {item.booking.bookerName || $_("items.anonymous")}
                            {/if}
                          </span>
                        </div>
                        {#if hasToken}
                          <Button
                            variant="ghost"
                            size="sm"
                            onclick={() => unbookItemByToken(item)}
                            class="h-7 gap-1 text-xs text-orange-600 hover:text-orange-700"
                          >
                            <XIcon class="h-3 w-3" />
                            <T key="items.cancelMyBooking" fallback="Cancel" />
                          </Button>
                        {/if}
                      </div>
                      {#if item.booking.message}
                        <p class="mt-1 text-sm text-green-600 dark:text-green-400">
                          "{item.booking.message}"
                        </p>
                      {/if}
                      <p class="mt-1 text-xs text-green-600 dark:text-green-400">
                        <T key="items.bookedOn" fallback="Booked on" />
                        {new Date(item.booking.bookedAt).toLocaleDateString()}
                      </p>
                    </div>
                  {:else}
                    <div class="flex gap-2">
                      <Button
                        onclick={() => startBooking(item.id)}
                        variant="outline"
                        size="sm"
                        class="flex-1 gap-2"
                      >
                        <BookOpenIcon class="h-4 w-4" />
                        <T key="items.bookItem" fallback="Book Item" />
                      </Button>
                    </div>
                  {/if}
                {/if}
              </CardContent>
            {/if}
          </Card>
        {/each}
      </div>
    {:else}
      <div class="py-12 text-center">
        <PlusIcon class="text-muted-foreground mx-auto mb-2 h-12 w-12" />
        <h3 class="text-muted-foreground text-lg font-semibold">
          <T key="wishlists.noItemsDescription" fallback="No items yet" />
        </h3>
        <p class="text-muted-foreground">
          <T key="wishlists.addFirstItem" fallback="Add your first item to this wishlist!" />
        </p>
      </div>
    {/if}
    <Separator class="my-6" />
    <p class="text-muted-foreground text-xs">
      <T
        key="wishlists.disclaimer"
        fallback="All links and other information presented on this page are provided by the user and do not reflect Wili's views or imply any affiliation, endorsement, or partnership."
      />
    </p>
  {/if}
</div>

<!-- Booking Modal -->
{#if bookingItemId}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <Card class="w-full max-w-md mx-4">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <BookOpenIcon class="h-5 w-5" />
          <T key="items.bookItem" fallback="Book Item" />
        </CardTitle>
        <CardDescription>
          <T key="items.bookItemDescription" fallback="Reserve this item for yourself" />
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div>
          <label class="text-sm font-medium">
            <T key="items.yourName" fallback="Your Name (Optional)" />
          </label>
          <Input
            bind:value={bookingForm.bookerName}
            placeholder={$_("items.yourNamePlaceholder")}
            class="mt-1"
          />
          <p class="text-muted-foreground text-xs mt-1">
            <T key="items.nameOptional" fallback="Leave empty for anonymous booking" />
          </p>
        </div>
        
        <div>
          <label class="text-sm font-medium">
            <T key="items.message" fallback="Message (Optional)" />
          </label>
          <Textarea
            bind:value={bookingForm.message}
            placeholder={$_("items.messagePlaceholder")}
            rows={3}
            class="mt-1"
          />
          <p class="text-muted-foreground text-xs mt-1">
            <T key="items.messageOptional" fallback="Add a message for the wishlist owner" />
          </p>
        </div>
        
        <div class="flex gap-2 pt-2">
          <Button
            onclick={bookItem}
            class="flex-1 gap-2"
          >
            <BookOpenIcon class="h-4 w-4" />
            <T key="items.bookItem" fallback="Book Item" />
          </Button>
          <Button variant="outline" onclick={cancelBooking} class="gap-2">
            <XIcon class="h-4 w-4" />
            <T key="common.cancel" fallback="Cancel" />
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
{/if}
