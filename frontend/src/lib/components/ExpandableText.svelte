<script lang="ts">
  import { ChevronDownIcon, ChevronUpIcon } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import Markdown from "./Markdown.svelte";
  import { browser } from "$app/environment";
  import { onDestroy } from "svelte";

  export let content: string = "";
  // If provided (>0), overrides responsive calculation
  export let maxHeight: number | null = null; // pixels
  // Enable responsive calculation of collapsed height based on viewport
  export let useResponsive: boolean = true;
  // Fractions of viewport height used when responsive is enabled
  export let mobileFraction: number = 0.45; // ~45% on mobile (one card per row)
  export let desktopFraction: number = 0.30; // ~30% on desktop (multi-column)
  // Clamp bounds to avoid extremes
  export let minMaxHeight: number = 180;
  export let maxMaxHeight: number = 520;
  export let showMarkdown: boolean = true;
  export let className: string = "";
  export let allowYandexMarket: boolean = true;
  // If content overflows only slightly, show it fully instead of forcing "Show more"
  export let smallOverflowThreshold: number = 28; // px

  let expanded = false;
  let contentElement: HTMLElement;
  let needsExpansion = false;
  let measuredHeight = 0;
  let computedResponsiveMaxHeight = 220;
  let effectiveMaxHeight = 220;
  let appliedMaxHeight = 220;

  // Check if content needs expansion after it's rendered
  $: if (contentElement && content) {
    checkNeedsExpansion();
  }

  // Recompute responsive max height on mount and when viewport changes
  function recomputeResponsiveMaxHeight() {
    if (!browser) return;
    const isDesktop = window.matchMedia("(min-width: 768px)").matches; // align with Tailwind md
    const fraction = isDesktop ? desktopFraction : mobileFraction;
    const suggested = Math.round(window.innerHeight * fraction);
    const clamped = Math.max(minMaxHeight, Math.min(maxMaxHeight, suggested));
    computedResponsiveMaxHeight = clamped;
  }

  if (browser) {
    recomputeResponsiveMaxHeight();
    window.addEventListener("resize", recomputeResponsiveMaxHeight);
  }

  onDestroy(() => {
    if (browser) {
      window.removeEventListener("resize", recomputeResponsiveMaxHeight);
    }
  });

  // Determine effective max height (explicit > responsive > default)
  $: effectiveMaxHeight =
    typeof maxHeight === "number" && maxHeight > 0
      ? maxHeight
      : useResponsive
      ? computedResponsiveMaxHeight
      : 220;
  $: appliedMaxHeight = effectiveMaxHeight;

  function checkNeedsExpansion() {
    if (!contentElement) return;

    // Create a temporary clone to measure the full height without affecting the DOM
    const clone = contentElement.cloneNode(true) as HTMLElement;
    clone.style.position = "absolute";
    clone.style.visibility = "hidden";
    clone.style.maxHeight = "none";
    clone.style.height = "auto";
    document.body.appendChild(clone);

    const fullHeight = clone.scrollHeight;
    document.body.removeChild(clone);

    measuredHeight = fullHeight;
    const overflow = fullHeight - effectiveMaxHeight;
    // Add a small buffer (5px) to avoid toggling on exact equals
    if (overflow > 5) {
      // If overflow is small, just show fully without a toggle
      if (overflow <= smallOverflowThreshold) {
        needsExpansion = false;
        appliedMaxHeight = fullHeight;
      } else {
        needsExpansion = true;
        appliedMaxHeight = effectiveMaxHeight;
      }
    } else {
      needsExpansion = false;
      appliedMaxHeight = effectiveMaxHeight;
    }
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
      style="max-height: {expanded || !needsExpansion ? 'none' : appliedMaxHeight + 'px'}"
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
        class="text-muted-foreground hover:text-foreground mt-2 h-6 px-2 text-xs"
      >
        {#if expanded}
          <ChevronUpIcon class="mr-1 h-3 w-3" />
          Show less
        {:else}
          <ChevronDownIcon class="mr-1 h-3 w-3" />
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
    content: "";
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
