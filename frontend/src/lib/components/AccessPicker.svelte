<script lang="ts">
  // Sélection d'accès par cases à cocher, regroupés par bâtiment, avec
  // recherche et sélection d'un bâtiment entier.
  import { Check, Minus } from '@lucide/svelte'
  import type { Access } from '../api'
  import { matches } from '../format'
  import SearchInput from './SearchInput.svelte'

  let {
    accesses,
    selected = $bindable([]),
    onchange,
    class: cls = 'max-h-80',
  }: { accesses: Access[]; selected?: number[]; onchange?: () => void; class?: string } = $props()

  let query = $state('')
  let building = $state('')

  const buildingNames = $derived([...new Set(accesses.map((a) => a.buildingName || 'Sans bâtiment'))].sort((a, b) => a.localeCompare(b, 'fr')))

  const groups = $derived.by(() => {
    const visible = accesses.filter(
      (a) =>
        (!building || (a.buildingName || 'Sans bâtiment') === building) &&
        matches(query, a.name, a.buildingName, a.floor, a.category, a.type),
    )
    const map = new Map<string, Access[]>()
    for (const a of visible) {
      const b = a.buildingName || 'Sans bâtiment'
      if (!map.has(b)) map.set(b, [])
      map.get(b)!.push(a)
    }
    return [...map.entries()].sort((x, y) => x[0].localeCompare(y[0], 'fr'))
  })

  const sel = $derived(new Set(selected))

  function toggle(id: number) {
    selected = sel.has(id) ? selected.filter((x) => x !== id) : [...selected, id]
    onchange?.()
  }

  function toggleGroup(items: Access[]) {
    const all = items.every((a) => sel.has(a.id))
    const ids = new Set(items.map((a) => a.id))
    selected = all ? selected.filter((x) => !ids.has(x)) : [...new Set([...selected, ...ids])]
    onchange?.()
  }
</script>

<div class="flex min-h-0 flex-col gap-3">
  <div class="flex flex-wrap items-center gap-2">
    <SearchInput bind:value={query} placeholder="Rechercher une porte, une salle…" class="min-w-48 flex-1" />
    {#if buildingNames.length > 1}
      <select class="input w-auto" bind:value={building} aria-label="Filtrer par bâtiment">
        <option value="">Tous les bâtiments</option>
        {#each buildingNames as b}<option value={b}>{b}</option>{/each}
      </select>
    {/if}
  </div>

  <div class="{cls} min-h-0 overflow-y-auto rounded-control border border-line">
    {#each groups as [name, items] (name)}
      {@const all = items.every((a) => sel.has(a.id))}
      {@const some = !all && items.some((a) => sel.has(a.id))}
      <div class="sticky top-0 z-[1] flex items-center justify-between border-b border-line bg-surface-2 px-3 py-1.5">
        <span class="text-[12px] font-semibold text-muted">{name}</span>
        <button class="flex cursor-pointer items-center gap-1.5 text-[12px] font-medium text-primary hover:underline" onclick={() => toggleGroup(items)}>
          {#if all}<Check size={13} /> Tout décocher{:else if some}<Minus size={13} /> Tout cocher{:else}Tout cocher{/if}
        </button>
      </div>
      {#each items as a (a.id)}
        <label class="flex cursor-pointer items-center gap-3 border-b border-line px-3 py-2 last:border-b-0 hover:bg-surface-2/60">
          <input type="checkbox" class="size-4 accent-[var(--c-primary)]" checked={sel.has(a.id)} onchange={() => toggle(a.id)} />
          <span class="min-w-0 flex-1 truncate text-[13px]">{a.name}</span>
          {#if a.floor}<span class="badge badge-neutral">{a.floor}</span>{/if}
          {#if a.keyIds.length === 0}<span class="badge badge-warning" title="Aucune clé n'ouvre cet accès">aucune clé</span>{/if}
        </label>
      {/each}
    {:else}
      <p class="px-3 py-8 text-center text-[13px] text-muted">
        {accesses.length === 0 ? 'Aucun accès enregistré.' : 'Aucun accès ne correspond à la recherche.'}
      </p>
    {/each}
  </div>
  <p class="text-[12px] text-muted">{selected.length} accès sélectionné{selected.length > 1 ? 's' : ''}</p>
</div>
