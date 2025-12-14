<script lang="ts">
  import { _ } from "svelte-i18n";
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import { PlusIcon } from "@lucide/svelte";

  interface Props {
    onAdd: (name: string, description: string) => Promise<void>;
  }

  let { onAdd }: Props = $props();

  let expanded = $state(false);
  let name = $state("");
  let description = $state("");

  function cancel() {
    expanded = false;
    name = "";
    description = "";
  }

  async function add() {
    const trimmedName = name.trim();
    if (!trimmedName) return;
    await onAdd(trimmedName, description.trim());
    name = "";
    description = "";
    expanded = false;
  }
</script>

{#if expanded}
  <Card>
    <CardHeader>
      <CardTitle>{$_("wishlists.addNewItem")}</CardTitle>
    </CardHeader>
    <CardContent class="flex flex-col gap-3">
      <Input placeholder={$_("items.namePlaceholder")} bind:value={name} />
      <Textarea placeholder={$_("items.descriptionPlaceholder")} bind:value={description} />
      <div class="flex gap-2">
        <Button onclick={add} disabled={!name.trim()} class="gap-2">
          <PlusIcon class="h-4 w-4" />
          {$_("common.add")}
        </Button>
        <Button variant="outline" onclick={cancel}>
          {$_("common.cancel")}
        </Button>
      </div>
    </CardContent>
  </Card>
{:else}
  <Button variant="outline" onclick={() => (expanded = true)} class="w-full gap-2">
    <PlusIcon class="h-4 w-4" />
    {$_("wishlists.addItem")}
  </Button>
{/if}
