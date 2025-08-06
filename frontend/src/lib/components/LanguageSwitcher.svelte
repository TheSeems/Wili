<script lang="ts">
	import { Languages } from '@lucide/svelte';
	import { setLanguage, supportedLanguages, type SupportedLanguage } from '$lib/stores/language';
	import { locale } from '$lib/i18n';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import T from '$lib/components/T.svelte';

	$: currentLang = $locale as SupportedLanguage;

	function handleLanguageChange(lang: SupportedLanguage) {
		setLanguage(lang);
	}

</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-8 w-8">
		<Languages class="h-4 w-4" />
		<span class="sr-only"><T key="language.switchLanguage" fallback="Switch Language" /></span>
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end" class="w-40">
		<DropdownMenu.Label class="text-xs font-medium">
			<T key="language.switchLanguage" fallback="Switch Language" />
		</DropdownMenu.Label>
		<DropdownMenu.Separator />
		{#each Object.entries(supportedLanguages) as [lang, label]}
			<DropdownMenu.Item
				class="text-xs {currentLang === lang ? 'bg-accent' : ''}"
				onclick={() => handleLanguageChange(lang as SupportedLanguage)}
			>
				{label}
				{#if currentLang === lang}
					<span class="ml-auto text-xs">✓</span>
				{/if}
			</DropdownMenu.Item>
		{/each}
	</DropdownMenu.Content>
</DropdownMenu.Root>