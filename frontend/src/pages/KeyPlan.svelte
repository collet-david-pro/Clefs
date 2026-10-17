<script lang="ts">
  // Plan de clés : bâtiment → accès → clés. Sert aussi de vue « par bâtiment ».
  import { ChevronDown, FileText, Map as MapIcon } from '@lucide/svelte'
  import { Api, notifyDocument, run, type PlanBuilding } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { matches, plural } from '../lib/format'
  import { data, openDetail } from '../lib/state.svelte'

  let plan = $state<PlanBuilding[]>([])
  let loaded = $state(false)
  let query = $state('')
  let building = $state(0)
  let collapsed = $state<Record<number, boolean>>({})

  $effect(() => {
    void data.version
    void run(Api.GetKeyPlan()).then((p) => {
      plan = p ?? []
      loaded = true
    })
  })

  const visible = $derived(
    plan
      .filter((b) => !building || b.id === building)
      .map((b) => ({ ...b, accesses: b.accesses.filter((pa) => matches(query, pa.access.name, pa.access.floor, pa.access.category, ...pa.keys.map((k) => k.number))) }))
      .filter((b) => b.accesses.length || !query),
  )

  async function exportPDF() {
    const path = await run(Api.ExportKeyPlanPDF())
    if (path) notifyDocument('Plan de clés enregistré dans documents/.', path)
  }
</script>

<PageHeader title="Plan de clés" subtitle="Pour chaque bâtiment, les accès et les clés qui les ouvrent.">
  {#snippet actions()}
    <button class="btn btn-secondary" onclick={exportPDF}><FileText size={15} />Exporter en PDF</button>
  {/snippet}
</PageHeader>

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Porte ou numéro de clé…" />
  <div class="segmented flex-wrap">
    <button aria-pressed={building === 0} onclick={() => (building = 0)}>Tous</button>
    {#each plan as b (b.id)}<button aria-pressed={building === b.id} onclick={() => (building = b.id)}>{b.name}</button>{/each}
  </div>
</div>

{#if loaded && !visible.length}
  <div class="card"><EmptyState icon={MapIcon} title="Rien à afficher" text={plan.length ? 'Aucun accès ne correspond à la recherche.' : "Créez d'abord des bâtiments et des accès."} /></div>
{/if}

<div class="space-y-4">
  {#each visible as b (b.id)}
    <section class="card overflow-hidden">
      <button class="flex w-full cursor-pointer items-center gap-3 border-b border-line px-5 py-3 text-left" onclick={() => (collapsed[b.id] = !collapsed[b.id])}>
        <ChevronDown size={16} class="text-muted transition-transform {collapsed[b.id] ? '-rotate-90' : ''}" />
        <h2 class="flex-1 text-[14.5px] font-semibold">{b.name}</h2>
        <span class="text-[12.5px] text-muted">{plural(b.accesses.length, 'accès', 'accès')}</span>
      </button>
      {#if !collapsed[b.id]}
        {#if b.accesses.length}
          <table class="table">
            <tbody>
              {#each b.accesses as pa (pa.access.id)}
                <tr>
                  <td class="w-72">
                    <button class="cursor-pointer text-left font-medium hover:text-primary" onclick={() => openDetail('access', pa.access.id)}>{pa.access.name}</button>
                    {#if pa.access.floor || pa.access.category}<p class="text-[12px] text-muted">{[pa.access.floor, pa.access.category].filter(Boolean).join(' · ')}</p>{/if}
                  </td>
                  <td>
                    {#if pa.keys.length}
                      <div class="flex flex-wrap gap-1">
                        {#each pa.keys as k (k.id)}
                          <button class="chip chip-key cursor-pointer hover:border-primary" title={k.description} onclick={() => openDetail('key', k.id)}>{k.number}</button>
                        {/each}
                      </div>
                    {:else}
                      <span class="badge badge-warning">aucune clé</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {:else}
          <p class="px-5 py-4 text-[13px] text-muted">Aucun accès dans ce bâtiment.</p>
        {/if}
      {/if}
    </section>
  {/each}
</div>
