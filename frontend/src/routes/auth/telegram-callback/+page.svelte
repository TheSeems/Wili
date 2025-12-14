<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { handleTelegramCallback } from "$lib/auth";
  import { _ } from "svelte-i18n";
  import { Loader2Icon, AlertCircleIcon } from "@lucide/svelte";

  let error = $state<string | null>(null);
  let processing = $state(true);

  onMount(() => {
    const token = page.url.searchParams.get("token");
    const state = page.url.searchParams.get("state");

    if (!token || !state) {
      error = $_("auth.telegramCallbackMissingParams");
      processing = false;
      return;
    }

    const success = handleTelegramCallback(token, state);
    if (success) {
      window.history.replaceState({}, "", "/auth/telegram-callback");
      goto("/");
    } else {
      error = $_("auth.telegramCallbackStateMismatch");
      processing = false;
    }
  });
</script>

<svelte:head>
  <title>{$_("auth.telegramCallback")} - Wili</title>
</svelte:head>

<div class="flex h-[80vh] flex-col items-center justify-center px-4">
  {#if processing}
    <div class="flex flex-col items-center gap-4">
      <Loader2Icon class="h-12 w-12 animate-spin text-primary" />
      <p class="text-muted-foreground">{$_("auth.processingTelegramLogin")}</p>
    </div>
  {:else if error}
    <div class="flex flex-col items-center gap-4 text-center">
      <AlertCircleIcon class="text-destructive h-12 w-12" />
      <p class="text-destructive">{error}</p>
      <a href="/" class="text-primary hover:underline">{$_("common.backToHome")}</a>
    </div>
  {/if}
</div>
