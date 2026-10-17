<script lang="ts">
  // Inventaire des clés : stock, détenteurs, accès. Clic → fiche détaillée.
  import { ClipboardCheck, FileSpreadsheet, FileText, KeyRound, Plus, TriangleAlert } from '@lucide/svelte'
  import { Api, notifyDocument, run, type Access, type Key } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import KeyForm from '../lib/components/KeyForm.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import StockBar from '../lib/components/StockBar.svelte'
  import { keyCategoryLabel, matches, plural } from '../lib/format'
  import { data, navigate, openDetail } from '../lib/state.svelte'

  let keys = $state<Key[]>([])
  let accesses = $state<Access[]>([])
  let loaded = $state(false)
  let query = $state('')
  let category = $state('')
  let view = $state<'all' | 'out' | 'problems'>('all')
  let formOpen = $state(false)

  $effect(() => {
    void data.version
    void load()
  })

  async function load() {
    const [k, a] = await Promise.all([run(Api.ListKeys()), run(Api.ListAccesses())])
    keys = k ?? []
    accesses = a ?? []
    loaded = true
  }

  const accessName = $derived(new Map(accesses.map((a) => [a.id, a.name])))
  const anomalies = $derived(keys.filter((k) => k.available < 0))
  const filtered = $derived(
    keys.filter(
      (k) =>
        (!category || k.category === category) &&
        (view === 'all' || (view === 'out' && k.loaned > 0) || (view === 'problems' && k.available <= 0)) &&
        matches(query, k.number, k.description, k.storageLocation, ...k.holders, ...k.accessIds.map((id) => accessName.get(id))),
    ),
  )

  async function stockPDF() {
    const path = await run(Api.ExportKeyStockPDF())
    if (path) notifyDocument('Bilan du stock enregistré dans documents/.', path)
  }
  async function exportCSV() {
    const path = await run(Api.ExportKeysCSV())
    if (path) notifyDocument('Inventaire exporté dans documents/.', path)
  }
</script>

<PageHeader title="Clés" subtitle="{plural(keys.length, 'clé')} · {keys.reduce((n, k) => n + k.total, 0)} exemplaires">
  {#snippet actions()}
    <button class="btn btn-secondary" onclick={() => navigate('inventory')}><ClipboardCheck size={15} />Inventaire</button>
    <button class="btn btn-secondary" onclick={stockPDF}><FileText size={15} />Bilan du stock (PDF)</button>
    <button class="btn btn-secondary" onclick={exportCSV}><FileSpreadsheet size={15} />Exporter (Excel)</button>
    <button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouvelle clé</button>
  {/snippet}
</PageHeader>

{#if anomalies.length}
  <div class="alert alert-danger mb-4">
    <TriangleAlert size={18} class="mt-0.5 shrink-0 text-danger" />
    <div>
      <p class="font-semibold">Erreur d'inventaire, vérifier le stock</p>
      <p class="text-muted">Plus d'exemplaires sont sortis que le stock ne le permet pour : {anomalies.map((k) => k.number).join(', ')}.</p>
    </div>
    <button class="btn btn-secondary btn-sm ml-auto" onclick={() => navigate('inventory')}>Faire un inventaire</button>
  </div>
{/if}

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Numéro, description, porte, détenteur…" class="w-80" />
  <select class="input w-auto" bind:value={category} aria-label="Catégorie">
    <option value="">Toutes catégories</option>
    {#each Object.entries(keyCategoryLabel) as [value, label]}<option {value}>{label}</option>{/each}
  </select>
  <div class="segmented">
    <button aria-pressed={view === 'all'} onclick={() => (view = 'all')}>Toutes</button>
    <button aria-pressed={view === 'out'} onclick={() => (view = 'out')}>Sorties</button>
    <button aria-pressed={view === 'problems'} onclick={() => (view = 'problems')}>Épuisées ou en erreur</button>
  </div>
</div>

<div class="card overflow-hidden">
  {#if filtered.length}
    <table class="table table-hover">
      <thead>
        <tr><th class="w-32">Numéro</th><th>Description</th><th class="w-56">Disponibles / total</th><th>Détenteurs</th><th>Accès</th><th>Rangement</th></tr>
      </thead>
      <tbody>
        {#each filtered as k (k.id)}
          <tr onclick={() => openDetail('key', k.id)}>
            <td><span class="chip chip-key">{k.number}</span></td>
            <td>
              <p class="font-medium">{k.description || '—'}</p>
              {#if k.category && k.category !== 'simple'}<span class="text-[12px] text-muted">{keyCategoryLabel[k.category] ?? k.category}</span>{/if}
            </td>
            <td><StockBar key={k} /></td>
            <td class="max-w-56 truncate text-muted">{k.holders.length ? k.holders.join(', ') : '—'}</td>
            <td class="text-muted">{k.accessIds.length ? plural(k.accessIds.length, 'porte') : '—'}</td>
            <td class="max-w-44 truncate text-muted">{k.storageLocation || '—'}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if loaded}
    <EmptyState icon={KeyRound} title={keys.length ? 'Aucune clé ne correspond' : 'Aucune clé enregistrée'} text={keys.length ? 'Modifiez la recherche ou les filtres.' : 'Ajoutez vos clés et indiquez les portes que chacune ouvre.'}>
      {#if !keys.length}<button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouvelle clé</button>{/if}
    </EmptyState>
  {/if}
</div>

<KeyForm bind:open={formOpen} onsaved={(id) => openDetail('key', id)} />
