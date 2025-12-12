<script lang="ts">
  import { ChevronDownIcon, ChevronUpIcon } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import Markdown from "./Markdown.svelte";
  import { browser } from "$app/environment";
  import { onDestroy } from "svelte";

  export let content: string = "";
  export let maxHeight: number | null = null;
  export let useResponsive: boolean = true;
  export let mobileFraction: number = 0.45;
  export let desktopFraction: number = 0.3;
  export let minMaxHeight: number = 180;
  export let maxMaxHeight: number = 520;
  export let showMarkdown: boolean = true;
  export let className: string = "";
  export let allowYandexMarket: boolean = true;
  export let smallOverflowThreshold: number = 28;

  let expanded = false;
  let contentElement: HTMLElement;
  let needsExpansion = false;
  let computedResponsiveMaxHeight = 220;
  let effectiveMaxHeight = 220;
  let appliedMaxHeight = 220;

  $: if (contentElement && content) {
    checkNeedsExpansion();
  }

  function recomputeResponsiveMaxHeight() {
    if (!browser) return;
    const isDesktop = window.matchMedia("(min-width: 768px)").matches;
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

  $: effectiveMaxHeight =
    typeof maxHeight === "number" && maxHeight > 0
      ? maxHeight
      : useResponsive
        ? computedResponsiveMaxHeight
        : 220;
  $: appliedMaxHeight = effectiveMaxHeight;

  function checkNeedsExpansion() {
    if (!contentElement) return;

    const clone = contentElement.cloneNode(true) as HTMLElement;
    clone.style.position = "absolute";
    clone.style.visibility = "hidden";
    clone.style.maxHeight = "none";
    clone.style.height = "auto";
    document.body.appendChild(clone);

    const fullHeight = clone.scrollHeight;
    document.body.removeChild(clone);

    const overflow = fullHeight - effectiveMaxHeight;
    if (overflow > 5) {
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
