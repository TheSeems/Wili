<script lang="ts">
  import "../app.css";
  import NavBar from "$lib/components/NavBar.svelte";
  import AlertContainer from "$lib/components/AlertContainer.svelte";
  import { page } from "$app/state";
  import { initApi } from "$lib/auth";
  import { browser } from "$app/environment";
  import { ModeWatcher } from "mode-watcher";
  import { locale, isLoading } from '$lib/i18n';
  import { onMount } from 'svelte';
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
</script>

{#if i18nReady}
  {#if page.url.pathname !== "/auth/callback"}
    <NavBar />
  {/if}

  <AlertContainer />

  <ModeWatcher />
  {@render children?.()}
{:else}
  <!-- Loading state while i18n initializes -->
  <div class="flex items-center justify-center min-h-screen">
    <div class="animate-pulse">Loading...</div>
  </div>
{/if}
