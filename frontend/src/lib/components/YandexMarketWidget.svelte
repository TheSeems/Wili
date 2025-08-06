<script lang="ts">
	import { onMount } from 'svelte';
	import UrlBadge from './UrlBadge.svelte';
	
	interface Props {
		productUrl: string;
		className?: string;
	}

	let { productUrl, className = '' }: Props = $props();

	// Extract product ID from Yandex Market URL
	let productId = $derived((() => {
		try {
			const url = new URL(productUrl);
			const pathParts = url.pathname.split('/');
			const cardIndex = pathParts.indexOf('card');
			if (cardIndex !== -1 && pathParts[cardIndex + 2]) {
				const rawId = pathParts[cardIndex + 2];
				const cleanId = rawId.replace(/[^0-9]/g, '');
				return cleanId || null;
			}
			return null;
		} catch {
			return null;
		}
	})());

	// Generate unique container ID
	let containerId = $derived(`yandex-market-widget-${productId}-${Math.random().toString(36).substr(2, 9)}`);

	let widgetLoaded = $state(false);
	let widgetError = $state(false);

	onMount(() => {
		if (!productId || productId.length < 5) {
			widgetError = true;
			return;
		}
		
		// Load Yandex Market widget script if not already loaded
		if (!document.querySelector('script[src*="aflt.market.yandex.ru"]')) {
			const script = document.createElement('script');
			script.src = 'https://aflt.market.yandex.ru/widget/script/api';
			script.async = true;
			document.head.appendChild(script);
		}
		
		// Initialize widget
		const initWidget = () => {
			try {
				(window as any).removeEventListener('YaMarketAffiliateLoad', initWidget);
				(window as any).YaMarketAffiliate.createWidget({
					type: 'offers',
					containerId: containerId,
					params: {
						clid: 2274126,
						searchSkuIds: [productId],
						themeId: 3
					}
				});
				
				// Check if widget loaded
				let checkCount = 0;
				const maxChecks = 50; // 5 seconds max
				
				const checkWidgetLoaded = () => {
					checkCount++;
					const container = document.getElementById(containerId);
					if (container && (container.children.length > 0 || container.innerHTML.trim() !== '')) {
						widgetLoaded = true;
					} else if (checkCount < maxChecks) {
						setTimeout(checkWidgetLoaded, 100);
					} else {
						widgetError = true;
					}
				};
				
				setTimeout(checkWidgetLoaded, 500);
			} catch (error) {
				console.error('Yandex Market widget error:', error);
				widgetError = true;
			}
		};
		
		if ((window as any).YaMarketAffiliate) {
			initWidget();
		} else {
			(window as any).addEventListener('YaMarketAffiliateLoad', initWidget);
		}
		
		return () => {
			window.removeEventListener('YaMarketAffiliateLoad', initWidget);
		};
	});
</script>

{#if widgetError || !productId}
	<!-- Fallback to UrlBadge if widget fails -->
	<UrlBadge url={productUrl} className={className} />
{:else}
	<!-- Always render container but show UrlBadge overlay while loading -->
	<div class="widget-wrapper">
		<!-- Widget container (always present for API) -->
		<div id={containerId} class="yandex-market-widget {className} py-4" style="max-width: 100%; margin: 1rem 0;"></div>
		
		{#if !widgetLoaded}
			<!-- Show UrlBadge overlay while widget is loading -->
			<div class="loading-overlay">
				<UrlBadge url={productUrl} className={className} />
			</div>
		{/if}
	</div>
{/if}

<style>
	.widget-wrapper {
		position: relative;
	}
	
	.loading-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10;
	}
	
	:global(.yandex-market-widget) {
		border-radius: 0.5rem;
		overflow: hidden;
		min-height: 60px; /* Ensure some height for overlay */
	}
	
	:global(.yandex-market-widget iframe) {
		border-radius: 0.5rem;
	}
</style>