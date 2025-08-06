<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/state";
	import { goto } from "$app/navigation";
	import { exchangeCode } from "$lib/auth";

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
			close();
		} catch (e) {
			alert("Authentication failed. Please try again: " + e);
			console.error(e);
		}
	});
</script>

<div class="flex items-center justify-center h-[80vh] flex-col gap-4">
	<h1 class="text-2xl font-bold">
		Authenticating…
	</h1>
	<p>
		If you are not redirected in a few seconds, please <a
			href="/"
			class="text-blue-500 underline">click here</a
		>.
	</p>
</div>
