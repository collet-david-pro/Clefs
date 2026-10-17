<script lang="ts">
  // En-tête de page : titre, sous-titre et actions à droite.
  import type { Snippet } from 'svelte'

  let { title, subtitle = '', actions }: { title: string; subtitle?: string; actions?: Snippet } = $props()
</script>

<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
  <div>
    <h1 class="text-[22px] font-semibold tracking-tight">{title}</h1>
    {#if subtitle}<p class="mt-1 text-[13.5px] text-muted">{subtitle}</p>{/if}
  </div>
  {#if actions}
    <div class="flex flex-wrap items-center gap-2">{@render actions()}</div>
  {/if}
</div>
