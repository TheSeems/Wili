<script lang="ts">
  import { _, isLoading } from "svelte-i18n";

  interface Props {
    key: string;
    fallback?: string;
    values?: Record<string, any>;
    tag?: string;
    class?: string;
  }

  let { key, fallback, values, tag = "span", class: className, ...restProps }: Props = $props();

  let autoFallback = $derived(fallback || key.split(".").pop() || key);

  let translatedText = $derived(
    $isLoading ? autoFallback : $_(key, values ? { values } : undefined) || autoFallback
  );
</script>

<svelte:element this={tag} class={className} {...restProps}>
  {translatedText}
</svelte:element>
