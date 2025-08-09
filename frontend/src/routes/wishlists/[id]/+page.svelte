<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/state";
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
	} from "@lucide/svelte";
	import { wishlistApi } from "$lib/api/wishlist-client";
	import type { components } from "$lib/api/generated/wishlist-api";
	import ExpandableText from "$lib/components/ExpandableText.svelte";
	import { showSuccessAlert, showInfoAlert } from "$lib/utils/alerts";
	import { _ } from 'svelte-i18n';
	import T from "$lib/components/T.svelte";
	import { 
		ITEM_NAME_MAX_LENGTH, 
		ITEM_DESCRIPTION_MAX_LENGTH,
		WISHLIST_TITLE_MAX_LENGTH,
		WISHLIST_DESCRIPTION_MAX_LENGTH, 
		validation 
	} from "$lib/api/validation-constants";
	type Wishlist = components["schemas"]["Wishlist"];
	type WishlistItem = components["schemas"]["WishlistItem"];

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

	const wishlistId = $derived(page.params.id);

	// SEO/OG derived values
	const defaultOgTitle = 'Wishlist — Wili';
	const defaultOgDescription = 'View this wishlist on Wili.';
	function summarize(text: string): string {
		const s = (text || '').toString().replace(/\s+/g, ' ').trim();
		return s.length > 160 ? s.slice(0, 160) + '…' : s;
	}
	const ogTitle = $derived(wishlist && wishlist.title ? `${wishlist.title} — Wili` : defaultOgTitle);
	const ogDescription = $derived(wishlist && wishlist.description ? summarize(wishlist.description) : defaultOgDescription);

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
			if (err instanceof Error && err.message.includes('404')) {
				error = $_('wishlists.wishlistNotFound');
			} else {
				error = err instanceof Error ? err.message : $_('wishlists.failedToLoad');
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
			error =
				err instanceof Error
					? err.message
					: $_('wishlists.failedToUpdate');
			console.error("Error updating wishlist:", err);
		}
	}

	async function deleteWishlist() {
		if (!$authStore.token || !wishlistId) return;
		if (!confirm($_('wishlists.confirmDelete'))) return;

		try {
			await wishlistApi.deleteWishlist(wishlistId, $authStore.token);
			goto("/wishlists");
		} catch (err) {
			error =
				err instanceof Error
					? err.message
					: $_('wishlists.failedToDelete');
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
			error = err instanceof Error ? err.message : $_('items.failedToAdd');
			console.error("Error adding item:", err);
		}
	}

	async function deleteItem(itemId: string) {
		if (!$authStore.token || !wishlistId) return;
		if (!confirm($_('items.confirmDelete'))) return;

		try {
			await wishlistApi.deleteWishlistItem(wishlistId, itemId, $authStore.token);
			// Reload wishlist to get updated items
			await loadWishlist();
		} catch (err) {
			error =
				err instanceof Error ? err.message : $_('items.failedToDelete');
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

	async function shareWishlist() {
		if (!wishlist) return;
		
		const shareUrl = `${window.location.origin}/wishlists/${wishlist.id}`;
		
		try {
			await navigator.clipboard.writeText(shareUrl);
			showSuccessAlert($_('wishlists.shareLinkCopied'), undefined, "top-right");
		} catch (err) {
			// Fallback for older browsers or when clipboard API fails
			console.error("Failed to copy to clipboard:", err);
			showInfoAlert($_('wishlists.shareLink'), shareUrl, "top-right");
		}
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
		<Button
			variant="ghost"
			onclick={() => goto("/wishlists")}
			class="mb-6 gap-2"
		>
			<ArrowLeftIcon class="h-4 w-4" />
			<T key="wishlists.backToWishlists" fallback="Back to Wishlists" />
		</Button>
	{/if}

	{#if loading}
		<div class="space-y-4">
			<div class="h-8 bg-gray-200 dark:bg-black rounded animate-pulse w-1/3"></div>
			<div class="h-4 bg-gray-100 dark:bg-black rounded animate-pulse w-1/2"></div>
			<div class="h-32 bg-gray-100 dark:bg-black rounded animate-pulse"></div>
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
		<div class="flex items-start justify-between mb-8">
			<div class="flex-1">
				<!-- Show owner info for anonymous viewers -->
				{#if !$authStore.token}
					<div class="mb-6 p-3 bg-muted rounded-lg">
						<p class="text-sm text-muted-foreground">
							<T key="auth.anonymousViewing" fallback="Anonymous viewing" />
							<a href="/" class="text-primary hover:underline"><T key="nav.login" fallback="Login" /></a>
							<T key="auth.loginToCreateWishlists" fallback="Login to create wishlists" />
						</p>
					</div>
				{/if}
				{#if editing}
					<div class="space-y-4">
						<Input
							bind:value={editForm.title}
							placeholder={$_('wishlists.titlePlaceholder')}
							class="text-2xl font-bold {!validation.isValidWishlistTitle(editForm.title) && editForm.title.length > 0 ? 'border-destructive' : ''}"
							required
							maxlength={WISHLIST_TITLE_MAX_LENGTH}
						/>
						{#if validation.getWishlistTitleErrorKey(editForm.title)}
							<p class="text-sm text-destructive">{$_(validation.getWishlistTitleErrorKey(editForm.title))}</p>
						{/if}
						<Textarea
							bind:value={editForm.description}
							placeholder={$_('wishlists.descriptionPlaceholder')}
							maxlength={WISHLIST_DESCRIPTION_MAX_LENGTH}
						/>
						<p class="text-sm text-muted-foreground">{editForm.description.length || 0} / {WISHLIST_DESCRIPTION_MAX_LENGTH}</p>
						{#if validation.getWishlistDescriptionErrorKey(editForm.description)}
							<p class="text-sm text-destructive">{$_(validation.getWishlistDescriptionErrorKey(editForm.description))}</p>
						{/if}
						<div class="flex items-center gap-3 pt-2">
							<Button 
								onclick={saveWishlist} 
								class="gap-2"
								disabled={!validation.isValidWishlistTitle(editForm.title) || !validation.isValidWishlistDescription(editForm.description)}
							>
								<SaveIcon class="h-4 w-4" />
								{$_('common.save')}
							</Button>
							<Button
								variant="outline"
								onclick={() => (editing = false)}
								class="gap-2"
							>
								<XIcon class="h-4 w-4" />
								{$_('common.cancel')}
							</Button>
						</div>
					</div>
				{:else}
					<h1 class="text-3xl font-bold mb-2">{wishlist.title}</h1>
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
							<span class="text-sm text-muted-foreground">
								{wishlist.items?.length || 0} <T key="wishlists.items" fallback="items" />
							</span>
						</div>
					{/if}
				{/if}
			</div>

			<!-- Only show edit controls if user is authenticated and owns the wishlist -->
			{#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id && !editing}
				<div class="flex items-center gap-2">
					<Button
						variant="outline"
						onclick={() => (editing = true)}
						class="gap-2"
					>
						<EditIcon class="h-4 w-4" />
						<T key="common.edit" fallback="Edit" />
					</Button>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button
									variant="ghost"
									size="sm"
									class="h-9 w-9 p-0"
								>
									<MoreVerticalIcon class="h-4 w-4" />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content align="end">
								<DropdownMenu.Item onclick={shareWishlist}>
									<ShareIcon class="mr-2 h-4 w-4" />
									<T key="wishlists.shareLink" fallback="Share Link" />
								</DropdownMenu.Item>
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
							placeholder={$_('wishlists.itemNamePlaceholder')}
							required
							maxlength={ITEM_NAME_MAX_LENGTH}
							class={!validation.isValidItemName(newItem.name) && newItem.name.length > 0 ? "border-destructive" : ""}
						/>
						{#if validation.getItemNameErrorKey(newItem.name)}
							<p class="text-sm text-destructive">{$_(validation.getItemNameErrorKey(newItem.name))}</p>
						{/if}
						<Textarea
							bind:value={newItem.description}
							placeholder={$_('wishlists.itemDescriptionPlaceholder')}
							maxlength={ITEM_DESCRIPTION_MAX_LENGTH}
							rows={5}
							class="min-h-[100px]"
						/>
						<p class="text-sm text-muted-foreground">{newItem.description.length || 0} / {ITEM_DESCRIPTION_MAX_LENGTH}</p>
						{#if validation.getItemDescriptionErrorKey(newItem.description)}
							<p class="text-sm text-destructive">{$_(validation.getItemDescriptionErrorKey(newItem.description))}</p>
						{/if}
						<div class="flex gap-2">
							<Button 
								onclick={addItem} 
								class="gap-2"
								disabled={!validation.isValidItemName(newItem.name) || !validation.isValidItemDescription(newItem.description)}
							>
								<SaveIcon class="h-4 w-4" />
								<T key="wishlists.addItem" fallback="Add Item" />
							</Button>
							<Button
								variant="outline"
								onclick={() => (addingItem = false)}
								class="gap-2"
							>
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
					<Card class="hover:shadow-md transition-shadow min-h-[100px]">
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
									class={!validation.isValidItemName(editItemForm.name) && editItemForm.name.length > 0 ? "border-destructive" : ""}
								/>
								{#if validation.getItemNameErrorKey(editItemForm.name)}
									<p class="text-sm text-destructive">{$_(validation.getItemNameErrorKey(editItemForm.name))}</p>
								{/if}
								<Textarea
									bind:value={editItemForm.description}
									placeholder="Description (optional) - Markdown supported. URLs will become badges!"
									maxlength={ITEM_DESCRIPTION_MAX_LENGTH}
									rows={10}
									class="min-h-[100px]"
								/>
								<p class="text-sm text-muted-foreground">{editItemForm.description.length || 0} / {ITEM_DESCRIPTION_MAX_LENGTH}</p>
								{#if validation.getItemDescriptionErrorKey(editItemForm.description)}
									<p class="text-sm text-destructive">{$_(validation.getItemDescriptionErrorKey(editItemForm.description))}</p>
								{/if}
								<div class="flex gap-2">
									<Button 
										onclick={() => saveItemEdit(item.id, item)} 
										class="gap-2"
										size="sm"
										disabled={!validation.isValidItemName(editItemForm.name) || !validation.isValidItemDescription(editItemForm.description)}
									>
										<SaveIcon class="h-4 w-4" />
										<T key="common.save" fallback="Save" />
									</Button>
									<Button
										variant="outline"
										onclick={cancelEditingItem}
										class="gap-2"
										size="sm"
									>
										<XIcon class="h-4 w-4" />
										<T key="common.cancel" fallback="Cancel" />
									</Button>
								</div>
							</CardContent>
						{:else}
							<!-- Display mode -->
							<CardHeader class="relative">
								<CardTitle class="line-clamp-2 pr-10">{item.data?.name || $_('wishlists.untitledItem')}</CardTitle>
								{#if item.data?.description}
									<CardDescription>
										<ExpandableText 
											content={item.data.description} 
											maxHeight={300}
											className="mt-2 text-muted-foreground"
										/>
									</CardDescription>
								{/if}
								<!-- Actions menu in top-right corner -->
								{#if $authStore.token && $authStore.user && wishlist && wishlist.userId === $authStore.user.id}
									<div class="absolute top-4 right-4">
										<DropdownMenu.Root>
											<DropdownMenu.Trigger>
												<Button
													variant="ghost"
													size="sm"
													class="h-8 w-8 p-0"
												>
													<MoreVerticalIcon class="h-4 w-4" />
												</Button>
											</DropdownMenu.Trigger>
											<DropdownMenu.Content align="end">
												<DropdownMenu.Item onclick={() => startEditingItem(item)}>
													<EditIcon class="mr-2 h-4 w-4" />
													<T key="common.edit" fallback="Edit" />
												</DropdownMenu.Item>
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

						{/if}
					</Card>
				{/each}
			</div>
		{:else}
			<div class="text-center py-12">
				<PlusIcon
					class="h-12 w-12 text-muted-foreground mx-auto mb-2"
				/>
				<h3 class="text-lg font-semibold text-muted-foreground">
					<T key="wishlists.noItemsDescription" fallback="No items yet" />
				</h3>
				<p class="text-muted-foreground">
					<T key="wishlists.addFirstItem" fallback="Add your first item to this wishlist!" />
				</p>
			</div>
		{/if}
	{/if}
</div>
