<script lang="ts">
  import { badgeVariants, type Variant } from "./index.js";
  import { cn } from "$lib/utils.js";

  let className: string | undefined = undefined;
  export { className as class };
  export let variant: Variant = "default";
</script>

<div class={cn(badgeVariants({ variant }), className)} {...$$restProps}>
  <slot />
</div>
