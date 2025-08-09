<script lang="ts">
  import { alertsStore, removeAlert, type AlertItem } from "$lib/stores/alerts";
  import * as Alert from "$lib/components/ui/alert/index.js";
  import { fly } from "svelte/transition";
  import XIcon from "@lucide/svelte/icons/x";

  $: alerts = $alertsStore;

  function dismissAlert(id: string) {
    removeAlert(id);
  }

  function getPositionClasses(position: string) {
    switch (position) {
      case "top-center":
        return "fixed top-4 left-1/2 transform -translate-x-1/2 z-50";
      case "top-right":
        return "fixed top-4 right-4 z-50";
      case "top-left":
        return "fixed top-4 left-4 z-50";
      case "bottom-center":
        return "fixed bottom-4 left-1/2 transform -translate-x-1/2 z-50";
      case "bottom-right":
        return "fixed bottom-4 right-4 z-50";
      case "bottom-left":
        return "fixed bottom-4 left-4 z-50";
      case "center":
        return "fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50";
      default:
        return "fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50";
    }
  }

  function getFlyTransition(position: string) {
    switch (position) {
      case "top-center":
        return { y: -50, duration: 300 };
      case "top-right":
        return { x: 300, duration: 300 };
      case "top-left":
        return { x: -300, duration: 300 };
      case "bottom-center":
        return { y: 50, duration: 300 };
      case "bottom-right":
        return { x: 300, duration: 300 };
      case "bottom-left":
        return { x: -300, duration: 300 };
      case "center":
        return { y: -20, duration: 300 }; // Fallback for center
      default:
        return { y: -20, duration: 300 };
    }
  }

  function getScaleTransition(position: string) {
    return { duration: 300, start: 0.8 };
  }

  function isScalePosition(position: string) {
    return position === "center";
  }

  // Group alerts by position
  $: alertsByPosition = alerts.reduce(
    (acc, alert) => {
      const position = alert.position || "center";
      if (!acc[position]) acc[position] = [];
      acc[position].push(alert);
      return acc;
    },
    {} as Record<string, AlertItem[]>
  );
</script>

{#each Object.entries(alertsByPosition) as [position, positionAlerts]}
  <div
    class="{getPositionClasses(position)} pointer-events-none flex w-full max-w-md flex-col gap-2"
  >
    {#each positionAlerts as alert (alert.id)}
      <div
        in:fly|global={{ y: -50, duration: 400 }}
        out:fly|global={{ y: -50, duration: 200 }}
        class="pointer-events-auto relative"
      >
        <Alert.Root variant={alert.variant || "default"} class="pr-8">
          {#if alert.icon}
            <svelte:component this={alert.icon} class="h-4 w-4" />
          {/if}
          <Alert.Title>{alert.title}</Alert.Title>
          {#if alert.description}
            <Alert.Description>{alert.description}</Alert.Description>
          {/if}

          <!-- Dismiss button -->
          <button
            onclick={() => dismissAlert(alert.id)}
            class="absolute top-2 right-2 rounded-sm p-1 opacity-70 transition-opacity hover:opacity-100"
          >
            <XIcon class="h-3 w-3" />
          </button>
        </Alert.Root>
      </div>
    {/each}
  </div>
{/each}
