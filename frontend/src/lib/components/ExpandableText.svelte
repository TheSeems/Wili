<script lang="ts">
	import { ChevronDownIcon, ChevronUpIcon } from "@lucide/svelte";
	import { Button } from "$lib/components/ui/button";
	import Markdown from "./Markdown.svelte";
	
	export let content: string = '';
	export let maxHeight: number = 100; // pixels
	export let showMarkdown: boolean = true;
	export let className: string = '';
	export let allowYandexMarket: boolean = true;

	let expanded = false;
	let contentElement: HTMLElement;
	let needsExpansion = false;
	
	// Check if content needs expansion after it's rendered
	$: if (contentElement && content) {
		checkNeedsExpansion();
	}
	
	function checkNeedsExpansion() {
		if (!contentElement) return;
		
		// Temporarily expand to measure full height
		const originalMaxHeight = contentElement.style.maxHeight;
		contentElement.style.maxHeight = 'none';
		const fullHeight = contentElement.scrollHeight;
		contentElement.style.maxHeight = originalMaxHeight;
		
		needsExpansion = fullHeight > maxHeight;
	}
	
	function toggleExpanded(event: MouseEvent) {
		event.stopPropagation();
		expanded = !expanded;
	}
</script>

{#if content}
	<div class="expandable-text {className}">
		<div 
			bind:this={contentElement}
			class="content-wrapper"
			class:collapsed={!expanded && needsExpansion}
			style="max-height: {expanded || !needsExpansion ? 'none' : maxHeight + 'px'}"
		>
			{#if showMarkdown}
				<Markdown {content} {allowYandexMarket} />
			{:else}
				{content}
			{/if}
		</div>
		
		{#if needsExpansion}
			<Button
				variant="ghost"
				size="sm"
				onclick={toggleExpanded}
				class="mt-2 h-6 px-2 text-xs text-muted-foreground hover:text-foreground"
			>
				{#if expanded}
					<ChevronUpIcon class="h-3 w-3 mr-1" />
					Show less
				{:else}
					<ChevronDownIcon class="h-3 w-3 mr-1" />
					Show more
				{/if}
			</Button>
		{/if}
	</div>
{/if}

<style>
	.content-wrapper {
		overflow: hidden;
		transition: max-height 0.3s ease-in-out;
	}
	
	.content-wrapper.collapsed {
		position: relative;
	}
	
	.content-wrapper.collapsed::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 20px;
		background: linear-gradient(transparent, rgb(255 255 255 / 0.8));
		pointer-events: none;
	}
	
	:global(.dark) .content-wrapper.collapsed::after {
		background: linear-gradient(transparent, rgb(0 0 0 / 0.8));
	}
</style>