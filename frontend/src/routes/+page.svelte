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
  import WiliLogo from "$lib/components/WiliLogo.svelte";

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

	// Manual OAuth approach - no widgets needed
	function redirectToYandexAuth() {
		if (!browser) return;
		
		const clientId = import.meta.env.VITE_YANDEX_CLIENT_ID;
		const redirectUri = `${window.location.origin}/auth/callback`;
		
		// Build OAuth URL manually
		const oauthUrl = new URL('https://oauth.yandex.ru/authorize');
		oauthUrl.searchParams.set('response_type', 'code');
		oauthUrl.searchParams.set('client_id', clientId);
		oauthUrl.searchParams.set('redirect_uri', redirectUri);
		oauthUrl.searchParams.set('scope', 'login:email login:info');
		
		// Redirect to Yandex OAuth
		window.location.href = oauthUrl.toString();
	}

	// No initialization needed for manual OAuth

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
        <h1 class="sr-only"><T key="home.title" fallback="Welcome to Wili" /></h1>
        <WiliLogo className="h-20 md:h-28 lg:h-32 mb-6" />
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
			<Button 
				onclick={redirectToYandexAuth}
				class="mt-8 w-full sm:w-1/3 bg-black hover:bg-black/90 text-white border border-white/10"
				aria-label="Login with Yandex"
			>
				<div class="flex items-center justify-center gap-2">
					<svg class="h-4 w-4" viewBox="0 0 24 24" aria-hidden="true">
						<circle cx="12" cy="12" r="10" fill="#FF0000" />
						<path d="M12 6c-1.6 0-2.7 1.1-2.7 2.6 0 1.3.8 2.1 1.9 2.5v6.8h1.6V11c1.1-.4 1.9-1.2 1.9-2.5C14.7 7.1 13.6 6 12 6z" fill="#fff"/>
					</svg>
					<T key="auth.loginWithYandex" fallback="Login with Yandex" />
				</div>
			</Button>
		{/if}
	{/if}
</section>

<!-- Manual OAuth - no external scripts needed -->
