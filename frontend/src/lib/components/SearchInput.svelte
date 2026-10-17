<script lang="ts">
  // Champ de recherche avec icône et bouton d'effacement.
  import { Search, X } from '@lucide/svelte'

  let { value = $bindable(''), placeholder = 'Rechercher…', class: cls = 'w-72' }: { value?: string; placeholder?: string; class?: string } = $props()
</script>

<div class="relative {cls}">
  <Search size={15} class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted" />
  <input class="input pr-8 pl-9" type="search" bind:value {placeholder} />
  {#if value}
    <button class="absolute top-1/2 right-2 -translate-y-1/2 cursor-pointer text-muted hover:text-ink" onclick={() => (value = '')} aria-label="Effacer">
      <X size={14} />
    </button>
  {/if}
</div>
