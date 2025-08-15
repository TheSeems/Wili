<script lang="ts">
  import { onMount } from "svelte";
  import { authStore } from "$lib/stores/auth";
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "$lib/components/ui/card";
  import { PlusIcon, ListIcon, CalendarIcon, PlusCircleIcon } from "@lucide/svelte";
  import { wishlistApi } from "$lib/api/wishlist-client";
  import type { components } from "$lib/api/generated/wishlist-api";
  import { browser } from "$app/environment";
  import ExpandableText from "$lib/components/ExpandableText.svelte";
  import T from "$lib/components/T.svelte";
  import I18nText from "$lib/components/I18nText.svelte";
  import { _ } from "svelte-i18n";
  import { validation } from "$lib/api/validation-constants";
  type Wishlist = components["schemas"]["Wishlist"];

  let wishlists: Wishlist[] = [];
  let loading = true;
  let error: string | null = null;

  $: if (!$authStore.token && !$authStore.isLoading) {
    if (browser) {
      goto("/");
    }
  }

  async function loadWishlists() {
    if (!$authStore.token) return;

    try {
      loading = true;
      const data = await wishlistApi.getWishlists($authStore.token);
      wishlists = data.wishlists || [];
    } catch (err) {
      error = err instanceof Error ? err.message : $_("wishlists.failedToLoad");
      console.error("Error loading wishlists:", err);
    } finally {
      loading = false;
    }
  }

  async function createWishlist() {
    if (!$authStore.token) return;

    const title = $_("wishlists.newWishlist");
    const description = $_("wishlists.newWishlistDescription");

    // Basic validation (safety net for i18n values)
    const titleErrorKey = validation.getWishlistTitleErrorKey(title);
    if (titleErrorKey) {
      error = $_(titleErrorKey);
      return;
    }

    const descriptionErrorKey = validation.getWishlistDescriptionErrorKey(description);
    if (descriptionErrorKey) {
      error = $_(descriptionErrorKey);
      return;
    }

    try {
      const newWishlist = await wishlistApi.createWishlist(
        {
          title,
          description,
        },
        $authStore.token
      );
      goto(`/wishlists/${newWishlist.id}`);
    } catch (err) {
      error = err instanceof Error ? err.message : $_("wishlists.failedToCreate");
      console.error("Error creating wishlist:", err);
    }
  }

  function formatDate(dateString: string) {
    return new Date(dateString).toLocaleDateString();
  }

  onMount(() => {
    if ($authStore.token) {
      loadWishlists();
    }
  });
</script>

<svelte:head>
  <title>{$_("wishlists.title")} - Wili</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
  <div class="mb-8 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
    <div>
      <I18nText key="wishlists.title" fallback="Wishlists" tag="h1" class="text-3xl font-bold" />
      <p class="text-muted-foreground mt-2">
        <T
          key="wishlists.manageDescription"
          fallback="Manage your wishlists and track your desired items"
        />
      </p>
    </div>
    <Button onclick={createWishlist} class="gap-2 self-start md:self-auto">
      <PlusIcon class="h-4 w-4" />
      <T key="wishlists.createWishlist" fallback="Create Wishlist" />
    </Button>
  </div>

  {#if loading}
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {#each Array(6) as _}
        <Card>
          <CardHeader>
            <div class="h-4 animate-pulse rounded bg-gray-200 dark:bg-black"></div>
            <div class="mt-2 h-3 animate-pulse rounded bg-gray-100 dark:bg-black"></div>
          </CardHeader>
          <CardContent>
            <div class="h-3 animate-pulse rounded bg-gray-100 dark:bg-black"></div>
          </CardContent>
        </Card>
      {/each}
    </div>
  {:else if error}
    <Card class="border-destructive">
      <CardContent class="pt-6">
        <p class="text-destructive">{error}</p>
        <Button onclick={loadWishlists} variant="outline" class="mt-4">
          <T key="common.tryAgain" fallback="Try Again" />
        </Button>
      </CardContent>
    </Card>
  {:else if wishlists.length === 0}
    <div class="py-12 text-center">
      <PlusIcon class="text-muted-foreground mx-auto mb-2 h-12 w-12" />
      <h3 class="text-muted-foreground text-lg font-semibold">
        <T key="wishlists.noWishlists" fallback="No Wishlists" />
      </h3>
      <p class="text-muted-foreground">
        <T key="wishlists.createFirstWishlist" fallback="Create your first wishlist" />
      </p>
    </div>
  {:else}
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {#each wishlists as wishlist}
        <Card
          class="cursor-pointer transition-shadow hover:shadow-md"
          onclick={() => {
            if (browser) {
              goto(`/wishlists/${wishlist.id}`);
            }
          }}
        >
          <CardHeader>
            <CardTitle class="line-clamp-1">{wishlist.title}</CardTitle>
            {#if wishlist.description}
              <CardDescription>
                <ExpandableText
                  content={wishlist.description}
                  maxHeight={400}
                  className="text-muted-foreground"
                  allowYandexMarket={false}
                />
              </CardDescription>
            {/if}
          </CardHeader>
          <CardContent>
            <div class="text-muted-foreground flex items-center justify-between text-sm">
              <div class="flex items-center gap-1">
                <ListIcon class="h-4 w-4" />
                <span
                  >{wishlist.items?.length || 0} <T key="wishlists.items" fallback="items" /></span
                >
              </div>
              <div class="flex items-center gap-1">
                <CalendarIcon class="h-4 w-4" />
                <span>{formatDate(wishlist.createdAt)}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      {/each}
    </div>
  {/if}
</div>
