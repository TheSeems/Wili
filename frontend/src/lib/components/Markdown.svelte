<script lang="ts">
	import { marked } from 'marked';
	import YandexMarketWidget from './YandexMarketWidget.svelte';
	import UrlBadge from './UrlBadge.svelte';
	
	interface Props {
		content?: string;
		allowYandexMarket?: boolean;
	}
	
	let { content = '', allowYandexMarket = true }: Props = $props();
	
	// Check if URL is a Yandex Market product link
	function isYandexMarketUrl(url: string): boolean {
		try {
			const urlObj = new URL(url);
			return urlObj.hostname.includes('market.yandex.ru') && urlObj.pathname.includes('/card/');
		} catch {
			return false;
		}
	}
	
	// Extract all standalone URLs for component rendering
	let standaloneUrls = $derived((() => {
		if (!content) return [];
		
		const urls: Array<{url: string, isYandexMarket: boolean, id: string}> = [];
		
		// First, let's extract URLs that are NOT part of markdown syntax
		const urlRegex = /https?:\/\/[^\s\)]+/g;
		let match;
		const matches: string[] = [];
		
		// Find all URLs and their positions
		while ((match = urlRegex.exec(content)) !== null) {
			const url = match[0];
			const startIndex = match.index;
			const endIndex = startIndex + url.length;
			
			// Check if this URL is part of markdown image syntax ![alt](url)
			const beforeUrl = content.substring(0, startIndex);
			const afterUrl = content.substring(endIndex);
			
			// Skip if it's an image URL
			if (/\.(jpg|jpeg|png|gif|webp|svg|ico|bmp|tiff)$/i.test(url)) {
				continue;
			}
			
			// Skip if it's part of markdown image syntax ![alt](url)
			if (afterUrl.startsWith(')')) {
				const beforeUrl = content.substring(0, startIndex);
				if (beforeUrl.includes('![') && beforeUrl.lastIndexOf('![') > beforeUrl.lastIndexOf('](')) {
					continue;
				}
			}
			
			// Skip if it's part of markdown link syntax [text](url)
			if (afterUrl.startsWith(')')) {
				const beforeUrl = content.substring(0, startIndex);
				if (beforeUrl.includes('[') && beforeUrl.lastIndexOf('[') > beforeUrl.lastIndexOf('](')) {
					continue;
				}
			}
			
			// Skip if it's inside parentheses but not part of markdown syntax
			if (afterUrl.startsWith(')')) {
				const beforeUrl = content.substring(0, startIndex);
				if (beforeUrl.includes('(') && beforeUrl.lastIndexOf('(') > beforeUrl.lastIndexOf(')')) {
					continue;
				}
			}
			
			// Only add URLs that are standalone (not part of any markdown syntax)
			matches.push(url);
		}
		
		return matches.map(url => ({
			url,
			isYandexMarket: isYandexMarketUrl(url),
			id: `url-${Math.random().toString(36).substr(2, 9)}`
		}));
	})());
	
	// Configure marked for security and styling
	marked.setOptions({
		breaks: true, // Convert line breaks to <br>
		gfm: true, // Enable GitHub Flavored Markdown
	});
	
	// Custom renderer to make all links open in new tabs and create placeholders for standalone URLs
	const renderer = new marked.Renderer();
	renderer.link = function(token) {
		const href = token.href;
		const title = token.title;
		const text = token.text;
		const titleAttr = title ? ` title="${title}"` : '';
		
		// If the link text is the same as the URL (or just domain), style it as a badge
		const isStandaloneUrl = text === href || text.replace(/^https?:\/\//, '') === href.replace(/^https?:\/\//, '');
		
		if (isStandaloneUrl) {
			// Generate unique ID for the placeholder
			const urlId = `url-${Math.random().toString(36).substr(2, 9)}`;
			return `<span class="url-placeholder" data-url="${href}" data-url-id="${urlId}"></span>`;
		}
		
		// Otherwise, render as a normal link
		return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
	};
	
	marked.setOptions({
		renderer: renderer
	});
	
	let htmlContent = $derived(content ? marked(content) : '');
</script>

{#if htmlContent}
	<div class="markdown-content">
		{@html htmlContent}
		
		<!-- Render URL badges and Yandex Market widgets for detected URLs -->
		{#each standaloneUrls as urlInfo}
			{#if urlInfo.isYandexMarket && allowYandexMarket}
				<YandexMarketWidget productUrl={urlInfo.url} />
			{:else}
				<UrlBadge url={urlInfo.url} variant="inline" />
			{/if}
		{/each}
	</div>
{/if}

<style>
	:global(.markdown-content) {
		color: inherit;
	}
	
	:global(.markdown-content h1) {
		font-size: 1.25rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		margin-top: 1rem;
	}
	
	:global(.markdown-content h1:first-child) {
		margin-top: 0;
	}
	
	:global(.markdown-content h2) {
		font-size: 1.125rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
		margin-top: 0.75rem;
	}
	
	:global(.markdown-content h2:first-child) {
		margin-top: 0;
	}
	
	:global(.markdown-content h3) {
		font-size: 1rem;
		font-weight: 600;
		margin-bottom: 0.25rem;
		margin-top: 0.5rem;
	}
	
	:global(.markdown-content h3:first-child) {
		margin-top: 0;
	}
	
	:global(.markdown-content p) {
		margin-bottom: 0.5rem;
	}
	
	:global(.markdown-content p:last-child) {
		margin-bottom: 0;
	}
	
	:global(.markdown-content ul), :global(.markdown-content ol) {
		margin-bottom: 0.5rem;
		padding-left: 1rem;
	}
	
	:global(.markdown-content li) {
		margin-bottom: 0.25rem;
	}
	
	:global(.markdown-content strong) {
		font-weight: 600;
	}
	
	:global(.markdown-content em) {
		font-style: italic;
	}
	
	:global(.markdown-content code) {
		background-color: rgb(248 250 252);
		padding: 0.125rem 0.25rem;
		border-radius: 0.25rem;
		font-size: 0.875rem;
		font-family: ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace;
	}
	
	:global(.dark .markdown-content code) {
		background-color: rgb(30 41 59);
	}
	
	:global(.markdown-content pre) {
		background-color: rgb(248 250 252);
		padding: 0.75rem;
		border-radius: 0.25rem;
		overflow-x: auto;
		margin-bottom: 0.5rem;
	}
	
	:global(.dark .markdown-content pre) {
		background-color: rgb(30 41 59);
	}
	
	:global(.markdown-content pre code) {
		background-color: transparent;
		padding: 0;
	}
	
	:global(.markdown-content blockquote) {
		border-left: 4px solid rgb(203 213 225);
		padding-left: 1rem;
		font-style: italic;
		margin-bottom: 0.5rem;
	}
	
	:global(.dark .markdown-content blockquote) {
		border-left-color: rgb(71 85 105);
	}
	
	:global(.markdown-content a) {
		color: rgb(59 130 246);
		text-decoration: none;
	}
	
	:global(.markdown-content a:hover) {
		text-decoration: underline;
	}
	
	:global(.dark .markdown-content a) {
		color: rgb(96 165 250);
	}
	
	:global(.markdown-content hr) {
		border: none;
		border-top: 1px solid rgb(203 213 225);
		margin: 0.75rem 0;
	}
	
	:global(.dark .markdown-content hr) {
		border-top-color: rgb(71 85 105);
	}

	/* URL Placeholder */
	:global(.markdown-content .url-placeholder) {
		display: none; /* Hide placeholder as we render components separately */
	}
</style>