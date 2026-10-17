<script lang="ts">
  // Accès (portes, portails, zones) avec les clés qui les ouvrent.
  import { DoorOpen, Plus } from '@lucide/svelte'
  import { Api, run, type Access, type Key } from '../lib/api'
  import AccessForm from '../lib/components/AccessForm.svelte'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { matches, plural } from '../lib/format'
  import { data, openDetail } from '../lib/state.svelte'

  let accesses = $state<Access[]>([])
  let keys = $state<Key[]>([])
  let loaded = $state(false)
  let query = $state('')
  let building = $state('')
  let floor = $state('')
  let formOpen = $state(false)

  $effect(() => {
    void data.version
    void load()
  })

  async function load() {
    const [a, k] = await Promise.all([run(Api.ListAccesses()), run(Api.ListKeys())])
    accesses = a ?? []
    keys = k ?? []
    loaded = true
  }

  const keyNumber = $derived(new Map(keys.map((k) => [k.id, k.number])))
  const buildings = $derived([...new Set(accesses.map((a) => a.buildingName).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'fr')))
  const floors = $derived([...new Set(accesses.map((a) => a.floor).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'fr')))
  const filtered = $derived(
    accesses.filter(
      (a) => (!building || a.buildingName === building) && (!floor || a.floor === floor) && matches(query, a.name, a.category, a.type, a.notes),
    ),
  )
  const withoutKey = $derived(accesses.filter((a) => a.keyIds.length === 0).length)
</script>

<PageHeader title="Accès" subtitle="{plural(accesses.length, 'accès', 'accès')} référencés{withoutKey ? ` · ${withoutKey} sans clé` : ''}">
  {#snippet actions()}
    <button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouvel accès</button>
  {/snippet}
</PageHeader>

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Désignation, catégorie…" />
  <select class="input w-auto" bind:value={building} aria-label="Bâtiment">
    <option value="">Tous les bâtiments</option>
    {#each buildings as b}<option value={b}>{b}</option>{/each}
  </select>
  {#if floors.length}
    <select class="input w-auto" bind:value={floor} aria-label="Étage">
      <option value="">Tous les étages</option>
      {#each floors as f}<option value={f}>{f}</option>{/each}
    </select>
  {/if}
</div>

<div class="card overflow-hidden">
  {#if filtered.length}
    <table class="table table-hover">
      <thead><tr><th>Désignation</th><th>Bâtiment</th><th>Étage</th><th>Catégorie</th><th>Clés qui l'ouvrent</th></tr></thead>
      <tbody>
        {#each filtered as a (a.id)}
          <tr onclick={() => openDetail('access', a.id)}>
            <td class="font-medium">{a.name}</td>
            <td class="text-muted">{a.buildingName || '—'}</td>
            <td class="text-muted">{a.floor || '—'}</td>
            <td class="text-muted">{a.category || '—'}</td>
            <td>
              {#if a.keyIds.length}
                <div class="flex flex-wrap gap-1">
                  {#each a.keyIds.slice(0, 6) as id}<span class="chip chip-key">{keyNumber.get(id)}</span>{/each}
                  {#if a.keyIds.length > 6}<span class="chip">+{a.keyIds.length - 6}</span>{/if}
                </div>
              {:else}
                <span class="badge badge-warning">aucune clé</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if loaded}
    <EmptyState icon={DoorOpen} title={accesses.length ? 'Aucun accès ne correspond' : 'Aucun accès enregistré'} text={accesses.length ? 'Modifiez la recherche ou les filtres.' : 'Enregistrez chaque porte, portail ou zone fermée à clé.'}>
      {#if !accesses.length}<button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouvel accès</button>{/if}
    </EmptyState>
  {/if}
</div>

<AccessForm bind:open={formOpen} onsaved={(id) => openDetail('access', id)} />
