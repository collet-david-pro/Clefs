<script lang="ts">
  // Recherche globale (Ctrl+K) : clés, accès, détenteurs et pages.
  import { DoorOpen, KeyRound, LayoutGrid, Search, User } from '@lucide/svelte'
  import { Api, run, type Access, type Borrower, type Key } from '../api'
  import { matches } from '../format'
  import { navigate, openDetail, type Page } from '../state.svelte'

  let { open = $bindable(false), pages }: { open?: boolean; pages: { page: Page; label: string }[] } = $props()

  let query = $state('')
  let active = $state(0)
  let keys = $state<Key[]>([])
  let accesses = $state<Access[]>([])
  let borrowers = $state<Borrower[]>([])
  let input = $state<HTMLInputElement>()

  type Item = { kind: 'key' | 'access' | 'borrower' | 'page'; id: number; page?: Page; title: string; subtitle: string }

  $effect(() => {
    if (open) {
      query = ''
      active = 0
      void load()
      setTimeout(() => input?.focus(), 0)
    }
  })

  async function load() {
    const [k, a, b] = await Promise.all([run(Api.ListKeys()), run(Api.ListAccesses()), run(Api.ListBorrowers())])
    keys = k ?? []
    accesses = a ?? []
    borrowers = b ?? []
  }

  const items = $derived.by<Item[]>(() => {
    const q = query.trim()
    const out: Item[] = []
    for (const p of pages) if (matches(q, p.label)) out.push({ kind: 'page', id: 0, page: p.page, title: p.label, subtitle: 'Page' })
    if (!q) return out.slice(0, 8)
    for (const b of borrowers)
      if (matches(q, b.name, b.email)) out.push({ kind: 'borrower', id: b.id, title: b.name, subtitle: b.activeLoans ? `Détenteur · ${b.activeLoans} clé(s)` : 'Détenteur' })
    for (const k of keys)
      if (matches(q, k.number, k.description, k.storageLocation)) out.push({ kind: 'key', id: k.id, title: `${k.number} — ${k.description}`, subtitle: `Clé · ${k.available} disponible(s) sur ${k.total}` })
    for (const a of accesses)
      if (matches(q, a.name, a.buildingName, a.category)) out.push({ kind: 'access', id: a.id, title: a.name, subtitle: `Accès · ${a.buildingName}${a.floor ? ' · ' + a.floor : ''}` })
    return out.slice(0, 40)
  })

  const icons = { key: KeyRound, access: DoorOpen, borrower: User, page: LayoutGrid }

  function choose(item: Item) {
    open = false
    if (item.kind === 'page') navigate(item.page!)
    else openDetail(item.kind, item.id)
  }

  function onkey(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      active = Math.min(active + 1, items.length - 1)
      e.preventDefault()
    } else if (e.key === 'ArrowUp') {
      active = Math.max(active - 1, 0)
      e.preventDefault()
    } else if (e.key === 'Enter' && items[active]) {
      choose(items[active])
    } else if (e.key === 'Escape') {
      open = false
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex justify-center px-6 pt-[12vh]">
    <button class="absolute inset-0 cursor-default bg-black/40" aria-label="Fermer" tabindex="-1" onclick={() => (open = false)}></button>
    <div class="relative flex max-h-[70vh] w-full max-w-2xl flex-col overflow-hidden rounded-card border border-line bg-surface shadow-pop" role="dialog" aria-label="Recherche">
      <div class="flex items-center gap-3 border-b border-line px-4">
        <Search size={18} class="text-muted" />
        <input
          bind:this={input}
          class="h-14 flex-1 bg-transparent text-[15px] outline-none placeholder:text-muted"
          placeholder="Rechercher une clé, une porte, une personne…"
          bind:value={query}
          oninput={() => (active = 0)}
          onkeydown={onkey}
        />
        <span class="kbd">Échap</span>
      </div>
      <ul class="overflow-y-auto p-2">
        {#each items as item, i (item.kind + item.id + (item.page ?? ''))}
          {@const Icon = icons[item.kind]}
          <li>
            <button
              class="flex w-full cursor-pointer items-center gap-3 rounded-control px-3 py-2.5 text-left {i === active ? 'bg-primary-soft' : 'hover:bg-surface-2'}"
              onmouseenter={() => (active = i)}
              onclick={() => choose(item)}
            >
              <Icon size={16} class={i === active ? 'text-primary' : 'text-muted'} />
              <span class="min-w-0 flex-1 truncate text-[13.5px] font-medium">{item.title}</span>
              <span class="shrink-0 text-[12px] text-muted">{item.subtitle}</span>
            </button>
          </li>
        {:else}
          <li class="px-3 py-10 text-center text-[13px] text-muted">Aucun résultat pour « {query} ».</li>
        {/each}
      </ul>
    </div>
  </div>
{/if}
