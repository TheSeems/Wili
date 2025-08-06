<script lang="ts">
	import { cn } from "$lib/utils.js";

	let className: string | undefined = undefined;
	export { className as class };
	export let orientation: "horizontal" | "vertical" = "horizontal";
	export let decorative = true;
</script>

<div
	role={decorative ? "none" : "separator"}
	aria-orientation={orientation}
	class={cn(
		"shrink-0 bg-border",
		orientation === "horizontal" ? "h-[1px] w-full" : "h-full w-[1px]",
		className,
	)}
	{...$$restProps}
></div>
