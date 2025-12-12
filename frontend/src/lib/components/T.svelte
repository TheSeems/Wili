<script lang="ts">
  import { _, isLoading } from "svelte-i18n";

  interface Props {
    key: string;
    fallback?: string;
    values?: Record<string, any>;
  }

  let { key, fallback, values }: Props = $props();

  let autoFallback = $derived(
    fallback ||
      key
        .split(".")
        .pop()
        ?.replace(/([A-Z])/g, " $1")
        .trim() ||
      key
  );

  let text = $derived(
    $isLoading ? autoFallback : $_(key, values ? { values } : undefined) || autoFallback
  );
</script>

{text}
