<script lang="ts">
  // Inventaire : on saisit les exemplaires présents dans l'armoire ; le total
  // devient « en stock + prêtés ». Mise à jour rapide après une clé retrouvée,
  // cassée ou perdue. Les clés non saisies ne sont pas modifiées.
  import { CircleCheck, ClipboardCheck, Eraser, FileText, Minus, Plus, TriangleAlert, WandSparkles } from '@lucide/svelte'
  import { Api, run, type Key } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import Modal from '../lib/components/Modal.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { matches, plural } from '../lib/format'
  import { confirm, data, inventoryDraft, openDetail, refresh } from '../lib/state.svelte'

  let keys = $state<Key[]>([])
  let loaded = $state(false)
  let query = $state('')
  let storage = $state('')
  let view = $state<'all' | 'todo' | 'diff'>('all')
  let saving = $state(false)
  let result = $state<{ counted: number; changed: number; delta: number; reportPath: string; reportError: string } | null>(null)
  let resultOpen = $state(false)
  const inputs: Record<number, HTMLInputElement> = {}

  $effect(() => {
    void data.version
    void run(Api.ListKeys()).then((k) => {
      keys = k ?? []
      loaded = true
    })
  })

  const counts = $derived(inventoryDraft.counts)
  const expected = (k: Key) => k.total - k.loaned // exemplaires censés être en stock
  const isCounted = (k: Key) => counts[k.id] !== undefined
  const newTotal = (k: Key) => (isCounted(k) ? counts[k.id] + k.loaned : k.total)
  const diff = (k: Key) => newTotal(k) - k.total

  const storages = $derived([...new Set(keys.map((k) => k.storageLocation).filter(Boolean))].sort((a, b) => a.localeCompare(b, 'fr')))
  const visible = $derived(
    keys.filter(
      (k) =>
        (!storage || k.storageLocation === storage) &&
        (view === 'all' || (view === 'todo' && !isCounted(k)) || (view === 'diff' && isCounted(k) && diff(k) !== 0)) &&
        matches(query, k.number, k.description, k.storageLocation),
    ),
  )
  const countedKeys = $derived(keys.filter(isCounted))
  const changedKeys = $derived(countedKeys.filter((k) => diff(k) !== 0 || counts[k.id] < k.reserve))
  const totalDelta = $derived(countedKeys.reduce((n, k) => n + diff(k), 0))

  function setCount(k: Key, value: number | null) {
    if (value === null || Number.isNaN(value)) {
      delete inventoryDraft.counts[k.id]
    } else {
      inventoryDraft.counts[k.id] = Math.max(0, Math.floor(value))
    }
  }

  function step(k: Key, delta: number) {
    const base = isCounted(k) ? counts[k.id] : Math.max(expected(k), 0)
    setCount(k, base + delta)
  }

  function oninput(k: Key, e: Event) {
    const v = (e.currentTarget as HTMLInputElement).value
    setCount(k, v === '' ? null : Number(v))
  }

  // Entrée : passer à la clé suivante (saisie rapide au clavier).
  function onkeydown(k: Key, e: KeyboardEvent) {
    if (e.key !== 'Enter' && e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return
    e.preventDefault()
    const i = visible.findIndex((x) => x.id === k.id)
    const next = visible[e.key === 'ArrowUp' ? i - 1 : i + 1]
    if (next) {
      inputs[next.id]?.focus()
      inputs[next.id]?.select()
    }
  }

  function fillExpected() {
    for (const k of visible) if (!isCounted(k)) setCount(k, Math.max(expected(k), 0))
  }

  async function clearAll() {
    if (!countedKeys.length) return
    if (await confirm({ title: 'Effacer la saisie ?', message: `Les comptages de ${plural(countedKeys.length, 'clé')} seront effacés.`, confirmLabel: 'Effacer' })) {
      inventoryDraft.counts = {}
    }
  }

  async function validate() {
    if (!countedKeys.length) return
    const lines = changedKeys
      .slice(0, 8)
      .map((k) => `• ${k.number} : ${k.total} → ${newTotal(k)}`)
      .join('\n')
    const more = changedKeys.length > 8 ? `\n… et ${changedKeys.length - 8} autre(s)` : ''
    const ok = await confirm({
      title: "Valider l'inventaire ?",
      message: changedKeys.length
        ? `${plural(countedKeys.length, 'clé comptée', 'clés comptées')}, ${plural(changedKeys.length, 'modifiée', 'modifiées')} :\n${lines}${more}\n\nUn rapport d'inventaire PDF sera créé.`
        : `${plural(countedKeys.length, 'clé comptée', 'clés comptées')}, aucun écart. Un rapport d'inventaire PDF sera créé.`,
      confirmLabel: "Valider l'inventaire",
    })
    if (!ok) return
    saving = true
    const items = countedKeys.map((k) => ({ keyId: k.id, inStock: counts[k.id] }))
    const res = await run(Api.ApplyInventory(items))
    saving = false
    if (res) {
      result = res
      resultOpen = true
      inventoryDraft.counts = {}
      refresh()
    }
  }
</script>

<PageHeader title="Inventaire" subtitle="Comptez les exemplaires présents dans l'armoire (réserve comprise) : le nouveau total = en stock + prêtés.">
  {#snippet actions()}
    <button class="btn btn-secondary" onclick={fillExpected} title="Reporter le stock attendu pour les clés pas encore comptées"><WandSparkles size={15} />Pré-remplir avec le stock attendu</button>
    <button class="btn btn-ghost" disabled={!countedKeys.length} onclick={clearAll}><Eraser size={15} />Effacer la saisie</button>
  {/snippet}
</PageHeader>

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Numéro, description…" />
  {#if storages.length}
    <select class="input w-auto" bind:value={storage} aria-label="Rangement">
      <option value="">Tous les rangements</option>
      {#each storages as s}<option value={s}>{s}</option>{/each}
    </select>
  {/if}
  <div class="segmented">
    <button aria-pressed={view === 'all'} onclick={() => (view = 'all')}>Toutes</button>
    <button aria-pressed={view === 'todo'} onclick={() => (view = 'todo')}>Pas encore comptées</button>
    <button aria-pressed={view === 'diff'} onclick={() => (view = 'diff')}>Avec écart</button>
  </div>
</div>

<div class="card mb-24 overflow-hidden">
  {#if visible.length}
    <table class="table">
      <thead>
        <tr>
          <th class="w-32">Numéro</th>
          <th>Description</th>
          <th class="w-24 text-right">Prêtées</th>
          <th class="w-28 text-right">Attendu en stock</th>
          <th class="w-48 text-center">Compté en stock</th>
          <th class="w-28 text-right">Nouveau total</th>
          <th class="w-24 text-right">Écart</th>
        </tr>
      </thead>
      <tbody>
        {#each visible as k (k.id)}
          {@const d = diff(k)}
          <tr class={isCounted(k) ? (d > 0 ? 'bg-success-soft/40' : d < 0 ? 'bg-danger-soft/40' : 'bg-primary-soft/30') : ''}>
            <td><button class="chip chip-key cursor-pointer hover:border-primary" onclick={() => openDetail('key', k.id)}>{k.number}</button></td>
            <td>
              <p class="font-medium">{k.description || '—'}</p>
              {#if k.storageLocation}<p class="text-[12px] text-muted">{k.storageLocation}</p>{/if}
            </td>
            <td class="text-right tabular-nums">{k.loaned}</td>
            <td class="text-right tabular-nums">
              {#if expected(k) < 0}
                <span class="badge badge-danger" title="Plus d'exemplaires prêtés que le total enregistré">sur-prêt</span>
              {:else}
                <span class="text-muted">{expected(k)}</span>
              {/if}
            </td>
            <td>
              <div class="flex items-center justify-center gap-1">
                <button class="btn btn-ghost btn-sm btn-icon" onclick={() => step(k, -1)} disabled={isCounted(k) && counts[k.id] <= 0} aria-label="Un de moins"><Minus size={14} /></button>
                <input
                  bind:this={inputs[k.id]}
                  class="input h-8 w-16 text-center font-semibold tabular-nums"
                  type="number"
                  min="0"
                  placeholder="—"
                  value={isCounted(k) ? counts[k.id] : ''}
                  oninput={(e) => oninput(k, e)}
                  onkeydown={(e) => onkeydown(k, e)}
                  aria-label="Exemplaires de {k.number} comptés en stock"
                />
                <button class="btn btn-ghost btn-sm btn-icon" onclick={() => step(k, 1)} aria-label="Un de plus"><Plus size={14} /></button>
              </div>
              {#if isCounted(k) && counts[k.id] < k.reserve}
                <p class="mt-1 text-center text-[11.5px] text-warning">réserve ramenée à {counts[k.id]}</p>
              {/if}
            </td>
            <td class="text-right font-semibold tabular-nums">{isCounted(k) ? newTotal(k) : ''}</td>
            <td class="text-right">
              {#if isCounted(k)}
                <span class="badge {d > 0 ? 'badge-success' : d < 0 ? 'badge-danger' : 'badge-neutral'}">{d > 0 ? `+${d}` : d < 0 ? `−${-d}` : '='}</span>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if loaded}
    <EmptyState icon={ClipboardCheck} title={keys.length ? 'Aucune clé à afficher' : 'Aucune clé enregistrée'} text={keys.length ? 'Modifiez la recherche ou les filtres.' : "Créez d'abord vos clés."} />
  {/if}
</div>

<!-- Barre de validation, toujours visible en bas de l'écran -->
<div class="pointer-events-none sticky bottom-4 z-10 -mt-20 flex justify-center">
  <div class="pointer-events-auto flex items-center gap-5 rounded-card border border-line bg-surface px-5 py-3 shadow-pop">
    <div class="text-[13px]">
      <span class="font-semibold">{plural(countedKeys.length, 'clé comptée', 'clés comptées')}</span>
      <span class="text-muted"> sur {keys.length}</span>
      {#if countedKeys.length}
        <span class="text-muted"> · {plural(changedKeys.length, 'modification', 'modifications')} · écart </span>
        <span class="font-semibold {totalDelta > 0 ? 'text-success' : totalDelta < 0 ? 'text-danger' : ''}">{totalDelta > 0 ? `+${totalDelta}` : totalDelta < 0 ? `−${-totalDelta}` : '0'}</span>
      {/if}
    </div>
    <button class="btn btn-primary" disabled={!countedKeys.length || saving} onclick={validate}><CircleCheck size={16} />Valider l'inventaire</button>
  </div>
</div>

<Modal bind:open={resultOpen} title="Inventaire enregistré" size="sm">
  {#if result}
    <div class="flex items-start gap-3">
      <CircleCheck size={24} class="shrink-0 text-success" />
      <p class="text-[14px]">
        <strong>{plural(result.counted, 'clé comptée', 'clés comptées')}</strong>,
        {result.changed ? `${plural(result.changed, 'stock mis à jour', 'stocks mis à jour')} (écart ${result.delta > 0 ? '+' : result.delta < 0 ? '−' : ''}${Math.abs(result.delta)}).` : 'aucun écart.'}
      </p>
    </div>
    {#if result.reportPath}
      <button class="btn btn-secondary mt-4 w-full" onclick={() => run(Api.OpenPath(result!.reportPath))}><FileText size={16} />Ouvrir le rapport d'inventaire</button>
    {:else if result.reportError}
      <div class="alert alert-warning mt-4"><TriangleAlert size={16} class="mt-0.5 shrink-0 text-warning" />Le rapport n'a pas pu être généré : {result.reportError}</div>
    {/if}
  {/if}
  {#snippet footer()}
    <button class="btn btn-primary" onclick={() => (resultOpen = false)}>Terminer</button>
  {/snippet}
</Modal>
