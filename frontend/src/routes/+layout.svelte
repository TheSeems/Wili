<script lang="ts">
  import "../app.css";
  import NavBar from "$lib/components/NavBar.svelte";
  import AlertContainer from "$lib/components/AlertContainer.svelte";
  import { page } from "$app/state";
  import { initApi } from "$lib/auth";
  import { browser } from "$app/environment";
  import { ModeWatcher } from "mode-watcher";
  import { locale, isLoading } from "$lib/i18n";
  import { onMount } from "svelte";
  import T from "$lib/components/T.svelte";
  let { children } = $props();

  let i18nReady = $state(!browser); // SSR is ready immediately

  onMount(() => {
    if (browser) {
      initApi();
    }
  });

  // Wait for i18n to be ready
  $effect(() => {
    if (browser) {
      // Ready when not loading and locale is set
      i18nReady = !$isLoading && $locale !== null && $locale !== undefined;
    }
  });

  // Basic SEO defaults
  const defaultTitle = "Wili — Create and share wishlists";
  const defaultDescription = "Wili lets you create, manage, and share wishlists effortlessly.";
  const ogImage = "/android-chrome-512x512.png";
  const twitterCard = "summary_large_image";
  const ogLocale = $derived($locale === "ru" ? "ru_RU" : "en_US");
  const pageTitle = $derived(
    page.url.pathname === "/privacy" ? "Privacy Policy — Wili" : defaultTitle
  );
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={defaultDescription} />
  <link rel="canonical" href={page.url.href} />

  <meta property="og:site_name" content="Wili" />
  <meta property="og:type" content="website" />
  <meta property="og:title" content={pageTitle} />
  <meta property="og:description" content={defaultDescription} />
  <meta property="og:url" content={page.url.href} />
  <meta property="og:image" content={ogImage} />
  <meta property="og:locale" content={ogLocale} />

  <meta name="twitter:card" content={twitterCard} />
  <meta name="twitter:title" content={pageTitle} />
  <meta name="twitter:description" content={defaultDescription} />
  <meta name="twitter:image" content={ogImage} />
  <link rel="sitemap" type="application/xml" href="/sitemap.xml" />

  <!-- JSON-LD: WebSite -->
  {@html `<script type="application/ld+json">${JSON.stringify({
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "Wili",
    url: page.url.origin,
    inLanguage: $locale === "ru" ? "ru-RU" : "en-US",
    description: defaultDescription
  })}</script>`}
</svelte:head>
{#if i18nReady}
  {#if page.url.pathname !== "/auth/callback"}
    <NavBar />
  {/if}

  <AlertContainer />

  <ModeWatcher />
  {@render children?.()}
  {#if page.url.pathname !== "/auth/callback"}
    <footer class="mt-10 py-6 text-center text-sm text-gray-500 dark:text-gray-400">
      <div class="flex items-center justify-center gap-4 opacity-70">
        <a
          href="https://t.me/wiliwish"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 hover:text-gray-700 dark:hover:text-gray-300"
        >
          <svg
            class="h-5 w-5"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <path
              fill="currentColor"
              d="M12 0C5.372 0 0 5.372 0 12s5.372 12 12 12 12-5.372 12-12S18.628 0 12 0zm5.403 7.206a.8.8 0 0 1 .277.777l-2.137 10.05c-.144.676-.566.838-1.148.52l-3.176-2.335-1.532 1.475c-.169.17-.31.312-.635.312l.228-3.228 5.873-5.304c.256-.228-.055-.356-.397-.128l-7.262 4.577-3.131-.978c-.68-.214-.693-.68.142-1.004l12.249-4.722c.568-.208 1.064.138.881 1.004z"
            />
          </svg>
          <span class="sr-only"><T key="footer.telegram" fallback="Telegram" /></span>
        </a>
        <a
          href="https://github.com/theseems/wili"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 hover:text-gray-700 dark:hover:text-gray-300"
        >
          <svg
            class="h-5 w-5"
            viewBox="0 0 98 96"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <path
              fill-rule="evenodd"
              clip-rule="evenodd"
              d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z"
              fill="currentColor"
            />
          </svg>
          <span class="sr-only"><T key="footer.github" fallback="GitHub" /></span>
        </a>
        <span class="h-5 w-px bg-gray-300 dark:bg-gray-700" aria-hidden="true"></span>
        <a href="/privacy" class="leading-5 hover:text-gray-700 dark:hover:text-gray-300"
          ><T key="footer.privacy" fallback="Privacy Policy" /></a
        >
      </div>
    </footer>
  {/if}
{:else}
  <!-- Loading state while i18n initializes -->
  <div class="flex min-h-screen items-center justify-center">
    <div class="animate-pulse">Loading...</div>
  </div>
{/if}
