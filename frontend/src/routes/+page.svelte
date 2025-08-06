<script lang="ts">
	import { browser } from "$app/environment";
	import { onMount } from "svelte";
	import { initApi, JUST_LOGGED_IN_KEY } from "$lib/auth";
	import { authStore } from "$lib/stores/auth";
	import { makeAlert } from "$lib/stores/alerts";
	import CheckCircle2Icon from "@lucide/svelte/icons/check-circle-2";
	import { Button } from "$lib/components/ui/button";
	import { ListIcon } from "@lucide/svelte";
	import { _, isLoading as i18nLoading } from 'svelte-i18n';
	import T from "$lib/components/T.svelte";
	import I18nText from "$lib/components/I18nText.svelte";

	$: ({ token, user, isLoading, justLoggedIn } = $authStore);

	$: if (justLoggedIn && user) {
		makeAlert({
			title: $i18nLoading ? 'Welcome!' : ($_('home.welcomeTitle') || 'Welcome!'),
			description: $i18nLoading ? `You have successfully logged in to Wili, ${user.displayName}!` : ($_('home.welcomeDescription', { values: { name: user.displayName } }) || `You have successfully logged in to Wili, ${user.displayName}!`),
			icon: CheckCircle2Icon,
			duration: 5000
		});
		authStore.update((state) => ({ ...state, justLoggedIn: false }));
		localStorage.removeItem(JUST_LOGGED_IN_KEY);
	}

	let yandexInitializing = false;
	let yandexInitialized = false;

	async function initYandexAuth() {
		if (!browser || !(window as any).YaAuthSuggest) return;
		if (token || isLoading || yandexInitializing) return;
		
		// Check if DOM element exists and is visible
		const btnElement = document.getElementById('yandex-btn');
		if (!btnElement || btnElement.hidden) return;
		
		yandexInitializing = true;
		
		try {
			// Clear any existing button content
			btnElement.innerHTML = '';
			
			// Wait a bit more for DOM to be fully ready
			await new Promise(resolve => setTimeout(resolve, 200));
			
			const { handler } = await (window as any).YaAuthSuggest.init(
				{
					client_id: import.meta.env.VITE_YANDEX_CLIENT_ID,
					response_type: "code",
					redirect_uri: `${window.location.origin}/auth/callback`,
				},
				window.location.origin,
				{
					view: "button",
					parentId: "yandex-btn",
					buttonSize: "s",
					buttonView: "additional",
					buttonTheme: "light",
					buttonBorderRadius: "28",
					buttonIcon: "ya",
				},
			);
			await handler();
			yandexInitialized = true;
		} catch (error) {
			// Only log if it's not the "in_progress" error
			if (error?.code !== 'in_progress') {
				console.error('Yandex auth init error:', error);
			}
		} finally {
			yandexInitializing = false;
		}
	}

	// Initialize on mount
	onMount(() => {
		if (!token && !isLoading) {
			// Wait for next tick to ensure DOM is ready
			setTimeout(initYandexAuth, 300);
		}
	});

	// Reset initialization state when user logs out
	$: if (!token && yandexInitialized) {
		yandexInitialized = false;
	}

	// Reinitialize when auth state changes (but only if not already initialized)
	$: if (!token && !isLoading && browser && !yandexInitialized && !yandexInitializing) {
		// Use longer timeout to ensure DOM is fully updated
		setTimeout(initYandexAuth, 400);
	}

</script>

<section class="flex flex-col items-center justify-center h-[80vh] px-4">
	{#if isLoading}
		<div class="flex flex-col items-center gap-4">
			<div
				class="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"
			></div>
			<p class="text-gray-600"><T key="common.loading" fallback="Loading..." /></p>
		</div>
	{:else}
		<I18nText key="home.title" fallback="Welcome to Wili 🎁" tag="h1" class="text-4xl font-bold mb-4" />
		<p class="text-gray-600 max-w-md text-center">
			<T key="home.description" fallback="Create and share wish-lists with friends, family, or anyone else." />
			{#if !token}
				<T key="home.loginPrompt" fallback=" Log in to start building yours." />
			{/if}
		</p>

		{#if token && user}
			<div class="mt-8 flex flex-col items-center gap-4">
				<Button href="/wishlists" class="flex items-center gap-2">
					<ListIcon class="h-4 w-4" />
<T key="nav.wishlists" fallback="My Wishlists" />
				</Button>
			</div>
		{:else}
			<div
				id="yandex-btn"
				class="mt-8 w-1/3"
				hidden={isLoading || token !== undefined}
			></div>
		{/if}
	{/if}
</section>

<svelte:head>
	{#if !user}
		<script
			src="https://yastatic.net/s3/passport-sdk/autofill/v1/sdk-suggest-with-polyfills-latest.js"
		></script>
	{/if}
</svelte:head>
