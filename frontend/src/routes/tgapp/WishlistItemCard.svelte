<script lang="ts">
  import type { components } from "$lib/api/generated/wishlist-api";
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import {
    CheckIcon,
    XIcon,
    ShieldOffIcon,
    EditIcon,
    SaveIcon,
    TrashIcon,
    EyeIcon,
  } from "@lucide/svelte";

  type WishlistItem = components["schemas"]["WishlistItem"];

  interface Props {
    item: WishlistItem;
    listId: string;
    isOwner: boolean;
    isBookingRevealed: boolean;
    hasBookingToken: boolean;
    defaultName: string | null;
    deleting: boolean;
    onRevealBooking: () => void;
    onBook: (anonymous: boolean, message: string) => Promise<void>;
    onUnbook: () => Promise<void>;
    onSave: (name: string, description: string) => Promise<void>;
    onDelete: () => Promise<void>;
  }

  let {
    item,
    listId,
    isOwner,
    isBookingRevealed,
    hasBookingToken,
    defaultName,
    deleting,
    onRevealBooking,
    onBook,
    onUnbook,
    onSave,
    onDelete,
  }: Props = $props();

  let editing = $state(false);
  let editName = $state("");
  let editDescription = $state("");
  let saving = $state(false);

  let booking = $state(false);
  let anonymous = $state(false);
  let message = $state("");

  function startEdit() {
    editName = item.data?.name || "";
    editDescription = item.data?.description || "";
    editing = true;
  }

  function cancelEdit() {
    editing = false;
    editName = "";
    editDescription = "";
  }

  async function save() {
    const name = editName.trim();
    if (!name) return;
    saving = true;
    try {
      await onSave(name, editDescription.trim());
      editing = false;
      editName = "";
      editDescription = "";
    } catch {
      // error handled by parent, keep editing open
    } finally {
      saving = false;
    }
  }

  function startBooking() {
    anonymous = defaultName ? false : true;
    message = "";
    booking = true;
  }

  function cancelBooking() {
    booking = false;
    anonymous = false;
    message = "";
  }

  async function confirmBook() {
    try {
      await onBook(anonymous, message.trim());
      cancelBooking();
    } catch {
      // error handled by parent, keep booking form open
    }
  }
</script>

<Card class="h-full">
  <CardHeader class="relative">
    <CardTitle class="line-clamp-2">{item.data?.name || $_("wishlists.untitledItem")}</CardTitle>
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
    {#if isOwner}
      <div class="absolute right-4 top-4 flex items-center gap-1">
        <Button
          variant="ghost"
          size="sm"
          class="h-8 w-8 p-0"
          disabled={deleting}
          onclick={startEdit}
        >
          <EditIcon class="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="text-destructive h-8 w-8 p-0"
          disabled={deleting}
          onclick={onDelete}
        >
          <TrashIcon class="h-4 w-4" />
        </Button>
      </div>
    {/if}
  </CardHeader>
  <CardContent class="flex flex-col gap-3">
    {#if editing && isOwner}
      <div class="space-y-3">
        <Input placeholder={$_("items.namePlaceholder")} bind:value={editName} />
        <Textarea placeholder={$_("items.descriptionPlaceholder")} bind:value={editDescription} />
        <div class="flex gap-2">
          <Button class="gap-2" disabled={saving || !editName.trim()} onclick={save}>
            <SaveIcon class="h-4 w-4" />
            {saving ? $_("common.loading") : $_("common.save")}
          </Button>
          <Button variant="outline" onclick={cancelEdit}>
            <XIcon class="h-4 w-4" />
            {$_("common.cancel")}
          </Button>
        </div>
      </div>
    {:else if isOwner}
      {#if isBookingRevealed}
        {#if item.booking}
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
          </div>
        {:else}
          <div class="text-muted-foreground px-1 py-2 text-center text-sm">
            {$_("items.itemNotBooked")}
          </div>
        {/if}
      {:else}
        <button
          type="button"
          onclick={onRevealBooking}
          class="text-muted-foreground hover:text-foreground flex w-full items-center justify-center gap-2 rounded-md border border-dashed py-2 text-sm"
        >
          <EyeIcon class="h-4 w-4" />
          {$_("items.revealBooking")}
        </button>
      {/if}
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
        {#if hasBookingToken}
          <Button size="sm" variant="outline" onclick={onUnbook}>
            <XIcon class="mr-2 h-4 w-4" />
            {$_("common.cancel")}
          </Button>
        {/if}
      </div>
    {:else if booking}
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
          <Button class="gap-2" onclick={confirmBook}>
            <CheckIcon class="h-4 w-4" />
            {$_("items.bookItem")}
          </Button>
          <Button variant="outline" onclick={cancelBooking}>
            <XIcon class="h-4 w-4" />
            {$_("common.cancel")}
          </Button>
        </div>
      </div>
    {:else}
      <Button class="w-full gap-2" onclick={startBooking}>
        <ShieldOffIcon class="h-4 w-4" />
        {$_("items.bookItem")}
      </Button>
    {/if}
  </CardContent>
</Card>
