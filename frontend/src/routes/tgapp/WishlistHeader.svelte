<script lang="ts">
  import type { components } from "$lib/api/generated/wishlist-api";
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "$lib/components/ui/dropdown-menu";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import {
    EditIcon,
    SaveIcon,
    SendIcon,
    TrashIcon,
    EllipsisVerticalIcon,
    ArrowLeftIcon,
  } from "@lucide/svelte";

  type Wishlist = components["schemas"]["Wishlist"];

  interface Props {
    wishlist: Wishlist;
    isOwner: boolean;
    deletingWishlist: boolean;
    telegramLoginAvailable: boolean;
    telegramLoginLoading: boolean;
    isLoggedIn: boolean;
    onGoBack: () => void;
    onLogin: () => void;
    onShare: () => void;
    onDelete: () => void;
    onSave: (title: string, description: string) => Promise<void>;
  }

  let {
    wishlist,
    isOwner,
    deletingWishlist,
    telegramLoginAvailable,
    telegramLoginLoading,
    isLoggedIn,
    onGoBack,
    onLogin,
    onShare,
    onDelete,
    onSave,
  }: Props = $props();

  let editing = $state(false);
  let editTitle = $state("");
  let editDescription = $state("");

  function startEdit() {
    editTitle = wishlist.title;
    editDescription = wishlist.description || "";
    editing = true;
  }

  function cancelEdit() {
    editing = false;
    editTitle = wishlist.title;
    editDescription = wishlist.description || "";
  }

  async function save() {
    await onSave(editTitle, editDescription);
    editing = false;
  }
</script>

<div class="flex items-center justify-between gap-2">
  <Button variant="ghost" onclick={onGoBack} class="gap-2">
    <ArrowLeftIcon class="h-4 w-4" />
    {$_("common.back")}
  </Button>
  {#if telegramLoginAvailable && !isLoggedIn}
    <Button variant="outline" disabled={telegramLoginLoading} onclick={onLogin}>
      {telegramLoginLoading ? $_("common.loading") : $_("auth.loginWithTelegram")}
    </Button>
  {/if}
</div>

<div class="space-y-2">
  {#if editing && isOwner}
    <div class="space-y-3">
      <input
        type="text"
        bind:value={editTitle}
        class="border-input bg-background w-full rounded-md border px-3 py-2 text-lg font-semibold outline-none focus:ring-2 focus:ring-ring"
      />
      <textarea
        bind:value={editDescription}
        rows="3"
        class="border-input bg-background w-full rounded-md border px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
      ></textarea>
      <div class="flex gap-2">
        <Button onclick={save} class="gap-2">
          <SaveIcon class="h-4 w-4" />
          {$_("common.save")}
        </Button>
        <Button variant="outline" onclick={cancelEdit}>
          {$_("common.cancel")}
        </Button>
      </div>
    </div>
  {:else}
    <div class="flex items-start justify-between gap-2">
      <p class="text-xl font-semibold">{wishlist.title}</p>
      <div class="flex shrink-0 items-center">
        {#if isOwner}
          <button
            type="button"
            onclick={startEdit}
            class="text-muted-foreground hover:text-foreground flex h-8 w-8 items-center justify-center rounded"
          >
            <EditIcon class="h-4 w-4" />
          </button>
        {/if}
        <button
          type="button"
          onclick={onShare}
          class="text-muted-foreground hover:text-foreground flex h-8 w-8 items-center justify-center rounded"
        >
          <SendIcon class="h-4 w-4" />
        </button>
        {#if isOwner}
          <DropdownMenu>
            <DropdownMenuTrigger>
              <button
                type="button"
                class="text-muted-foreground hover:text-foreground flex h-8 w-8 items-center justify-center rounded"
              >
                <EllipsisVerticalIcon class="h-4 w-4" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onclick={onDelete}
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
    </div>
    {#if wishlist.description}
      <ExpandableText
        content={wishlist.description}
        className="text-sm text-muted-foreground"
        maxHeight={200}
        useResponsive={false}
        allowYandexMarket={false}
        smallOverflowThreshold={0}
      />
    {:else if isOwner}
      <button
        type="button"
        onclick={startEdit}
        class="text-muted-foreground hover:text-foreground text-sm italic"
      >
        {$_("wishlists.descriptionPlaceholder")}
      </button>
    {/if}
  {/if}
</div>
