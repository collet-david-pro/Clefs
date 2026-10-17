<script lang="ts">
  // Qui a quoi ? Chaque détenteur avec ses clés et les accès qu'il peut ouvrir.
  import { CopyX, UserSearch } from '@lucide/svelte'
  import { Api, run, type WhoHasWhat } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { borrowerStatusLabel, matches, plural } from '../lib/format'
  import { data, openDetail } from '../lib/state.svelte'

  let rows = $state<WhoHasWhat[]>([])
  let loaded = $state(false)
  let query = $state('')

  $effect(() => {
    void data.version
    void run(Api.GetWhoHasWhat()).then((r) => {
      rows = r ?? []
      loaded = true
    })
  })

  const filtered = $derived(rows.filter((r) => matches(query, r.borrower.name, ...r.loans.map((l) => l.keyNumber), ...r.accesses)))
</script>

<PageHeader title="Qui a quoi ?" subtitle="Les clés détenues par chaque personne et les portes qu'elles lui ouvrent.">
  {#snippet actions()}<SearchInput bind:value={query} placeholder="Personne, clé ou porte…" />{/snippet}
</PageHeader>

{#if loaded && !filtered.length}
  <div class="card"><EmptyState icon={UserSearch} title={rows.length ? 'Aucun résultat' : 'Personne ne détient de clé'} /></div>
{/if}

<div class="grid grid-cols-2 gap-4">
  {#each filtered as r (r.borrower.id)}
    <button class="card cursor-pointer p-5 text-left transition-shadow hover:shadow-pop" onclick={() => openDetail('borrower', r.borrower.id)}>
      <div class="flex items-center gap-3">
        <span class="flex size-9 items-center justify-center rounded-full bg-primary-soft text-[14px] font-semibold text-primary">{r.borrower.name.charAt(0).toUpperCase()}</span>
        <div class="min-w-0 flex-1">
          <p class="truncate font-semibold">{r.borrower.name}</p>
          <p class="text-[12.5px] text-muted">{borrowerStatusLabel[r.borrower.status] ?? r.borrower.status} · {plural(r.loans.length, 'clé')} · {plural(r.accesses.length, 'porte')}</p>
        </div>
        {#if r.redundancies.length}<span class="badge badge-info" title="Clés en doublon"><CopyX size={11} />doublon</span>{/if}
      </div>
      <div class="mt-3 flex flex-wrap gap-1">
        {#each r.loans as l (l.id)}<span class="chip chip-key {l.status === 'overdue' ? 'border-warning text-warning' : ''}">{l.keyNumber}</span>{/each}
      </div>
      {#if r.accesses.length}
        <p class="mt-3 line-clamp-2 text-[12.5px] text-muted">{r.accesses.join(' · ')}</p>
      {/if}
    </button>
  {/each}
</div>
