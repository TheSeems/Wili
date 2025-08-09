<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { exchangeCode } from "$lib/auth";
  import T from "$lib/components/T.svelte";

  onMount(async () => {
    const url = new URL(page.url);
    const code = url.searchParams.get("code");
    if (!code) {
      console.error("No code found");
      await goto("/");
      return;
    }
    try {
      await exchangeCode(code);
      await goto("/");
    } catch (e) {
      alert("Authentication failed. Please try again: " + e);
      console.error(e);
    }
  });
</script>

<div class="m-auto flex h-[80vh] flex-col items-center justify-center gap-4">
  <h1 class="text-2xl font-bold">
    <T key="auth.authenticating" fallback="Authenticating…" /> <span class="animate-spin">⏳</span>
  </h1>
  <p>
    <T key="auth.redirectMessage" fallback="If you are not redirected in a few seconds, please" />
    <a href="/" class="text-blue-500 underline"><T key="auth.clickHere" fallback="click here" /></a
    >.
  </p>
</div>
