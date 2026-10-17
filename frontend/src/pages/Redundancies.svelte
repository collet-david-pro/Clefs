<script lang="ts">
  // Doublons : personnes dont plusieurs clés ouvrent les mêmes portes.
  import { CircleCheck } from '@lucide/svelte'
  import { Api, run, type Redundancy } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import { data, openDetail } from '../lib/state.svelte'

  let rows = $state<Redundancy[]>([])
  let loaded = $state(false)

  $effect(() => {
    void data.version
    void run(Api.GetRedundancies()).then((r) => {
      rows = r ?? []
      loaded = true
    })
  })
</script>

<PageHeader title="Doublons" subtitle="Personnes dont plusieurs clés ouvrent les mêmes portes : une clé pourrait souvent être récupérée." />

{#if loaded && !rows.length}
  <div class="card"><EmptyState icon={CircleCheck} title="Aucun doublon" text="Aucune personne ne détient deux clés ouvrant la même porte." /></div>
{/if}

<div class="space-y-3">
  {#each rows as r (r.borrowerId)}
    <button class="card flex w-full cursor-pointer items-start gap-5 p-5 text-left transition-shadow hover:shadow-pop" onclick={() => openDetail('borrower', r.borrowerId)}>
      <div class="w-56 shrink-0">
        <p class="font-semibold">{r.borrowerName}</p>
        <div class="mt-2 flex flex-wrap gap-1">{#each r.keyNumbers as n}<span class="chip chip-key">{n}</span>{/each}</div>
      </div>
      <div class="min-w-0 flex-1">
        <p class="text-[12.5px] font-medium text-muted">Portes ouvertes par plusieurs de ses clés</p>
        <div class="mt-2 flex flex-wrap gap-1">{#each r.accesses as a}<span class="chip">{a}</span>{/each}</div>
      </div>
    </button>
  {/each}
</div>
