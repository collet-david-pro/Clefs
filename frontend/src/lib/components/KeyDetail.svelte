<script lang="ts">
  // Fiche d'une clé : stock, accès ouverts, détenteurs actuels, historique.
  import { DoorOpen, HandCoins, Minus, Pencil, Plus, TriangleAlert, Trash2, Undo2 } from '@lucide/svelte'
  import { Api, run, type KeyDetail, type Loan } from '../api'
  import { fmtDate, fmtDays, keyCategoryLabel, loanStatusBadge, loanStatusLabel } from '../format'
  import { closeDetail, confirm, data, navigate, openDetail } from '../state.svelte'
  import Drawer from './Drawer.svelte'
  import KeyForm from './KeyForm.svelte'
  import ReturnDialog from './ReturnDialog.svelte'

  let { id }: { id: number } = $props()

  let d = $state<KeyDetail | null>(null)
  let editOpen = $state(false)
  let returnOpen = $state(false)
  let returning = $state<Loan[]>([])

  $effect(() => {
    void data.version
    void load(id)
  })

  async function load(keyId: number) {
    d = (await run(Api.GetKeyDetail(keyId))) ?? null
  }

  async function adjustTotal(delta: number) {
    if (!d) return
    await run(Api.SetKeyTotal(d.key.id, d.key.total + delta), 'Stock mis à jour.')
  }

  async function remove() {
    if (!d) return
    const ok = await confirm({
      title: `Supprimer la clé ${d.key.number} ?`,
      message: "La clé et son historique de prêts seront définitivement supprimés.\nLes accès qu'elle ouvre sont conservés.",
      confirmLabel: 'Supprimer',
      danger: true,
    })
    if (ok && (await run(Api.DeleteKey(d.key.id).then(() => true), 'Clé supprimée.'))) closeDetail()
  }

  function startReturn(loans: Loan[]) {
    returning = loans
    returnOpen = true
  }
</script>

<Drawer onclose={closeDetail}>
  {#snippet header()}
    {#if d}
      <div class="flex items-center gap-2">
        <span class="chip chip-key text-[13px]">{d.key.number}</span>
        <span class="badge badge-neutral">{keyCategoryLabel[d.key.category] ?? d.key.category}</span>
      </div>
      <h2 class="mt-2 truncate text-[18px] font-semibold">{d.key.description || 'Sans description'}</h2>
      {#if d.key.storageLocation}<p class="mt-0.5 text-[13px] text-muted">Rangement : {d.key.storageLocation}</p>{/if}
    {:else}
      <h2 class="text-[18px] font-semibold">Chargement…</h2>
    {/if}
  {/snippet}

  {#if d}
    {#if d.key.available < 0}
      <div class="alert alert-danger">
        <TriangleAlert size={18} class="mt-0.5 shrink-0 text-danger" />
        <div>
          <p class="font-semibold">Erreur d'inventaire, vérifier le stock</p>
          <p class="text-muted">{d.key.loaned} exemplaire(s) sorti(s) pour {d.key.total - d.key.reserve} utilisable(s).</p>
        </div>
      </div>
    {/if}

    <section>
      <div class="grid grid-cols-4 gap-2 text-center">
        <div class="rounded-control bg-surface-2 px-2 py-3">
          <p class="text-[20px] font-semibold tabular-nums">{d.key.total}</p>
          <p class="text-[11.5px] text-muted">au total</p>
        </div>
        <div class="rounded-control bg-surface-2 px-2 py-3">
          <p class="text-[20px] font-semibold tabular-nums">{d.key.reserve}</p>
          <p class="text-[11.5px] text-muted">en réserve</p>
        </div>
        <div class="rounded-control bg-surface-2 px-2 py-3">
          <p class="text-[20px] font-semibold tabular-nums">{d.key.loaned}</p>
          <p class="text-[11.5px] text-muted">sortie(s)</p>
        </div>
        <div class="rounded-control px-2 py-3 {d.key.available < 0 ? 'bg-danger-soft text-danger' : d.key.available === 0 ? 'bg-warning-soft text-warning' : 'bg-success-soft text-success'}">
          <p class="text-[20px] font-semibold tabular-nums">{d.key.available}</p>
          <p class="text-[11.5px]">disponible(s)</p>
        </div>
      </div>
      <div class="mt-3 flex items-center justify-between rounded-control border border-line px-3 py-2">
        <span class="text-[13px] text-muted">Ajuster le nombre total d'exemplaires</span>
        <div class="flex items-center gap-1">
          <button class="btn btn-secondary btn-sm btn-icon" onclick={() => adjustTotal(-1)} disabled={d.key.total <= Math.max(0, d.key.loaned, d.key.reserve)} aria-label="Retirer un exemplaire"><Minus size={14} /></button>
          <span class="w-8 text-center font-semibold tabular-nums">{d.key.total}</span>
          <button class="btn btn-secondary btn-sm btn-icon" onclick={() => adjustTotal(1)} aria-label="Ajouter un exemplaire"><Plus size={14} /></button>
        </div>
      </div>
    </section>

    <section>
      <h3 class="section-title mb-2">Accès ouverts ({d.accesses.length})</h3>
      {#if d.accesses.length}
        <div class="flex flex-wrap gap-1.5">
          {#each d.accesses as a (a.id)}
            <button class="chip cursor-pointer hover:border-primary" onclick={() => openDetail('access', a.id)}>
              <DoorOpen size={12} class="text-muted" />{a.name}<span class="text-muted">· {a.buildingName}</span>
            </button>
          {/each}
        </div>
      {:else}
        <p class="text-[13px] text-muted">Cette clé n'est associée à aucun accès.</p>
      {/if}
    </section>

    <section>
      <h3 class="section-title mb-2">Détenue actuellement par ({d.activeLoans.length})</h3>
      {#if d.activeLoans.length}
        <ul class="divide-y divide-line rounded-control border border-line">
          {#each d.activeLoans as l (l.id)}
            <li class="flex items-center gap-3 px-3 py-2.5">
              <button class="min-w-0 flex-1 cursor-pointer truncate text-left text-[13px] font-medium hover:text-primary" onclick={() => openDetail('borrower', l.borrowerId)}>{l.borrowerName}</button>
              <span class="badge {loanStatusBadge[l.status]}">{l.status === 'overdue' ? 'En retard' : fmtDays(l.days)}</span>
              <button class="btn btn-ghost btn-sm" onclick={() => startReturn([l])}><Undo2 size={14} />Retour</button>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-[13px] text-muted">Aucun exemplaire n'est sorti.</p>
      {/if}
    </section>

    <section>
      <h3 class="section-title mb-2">Historique</h3>
      {#if d.history.length}
        <ul class="space-y-1.5">
          {#each d.history.slice(0, 15) as l (l.id)}
            <li class="flex items-center gap-2 text-[13px]">
              <span class="w-20 text-muted tabular-nums">{fmtDate(l.loanDate)}</span>
              <span class="min-w-0 flex-1 truncate">{l.borrowerName}</span>
              <span class="badge {loanStatusBadge[l.status]}">{l.status === 'returned' ? `rendue le ${fmtDate(l.returnDate)}` : loanStatusLabel[l.status]}</span>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-[13px] text-muted">Jamais prêtée.</p>
      {/if}
    </section>

    {#if d.key.notes}
      <section>
        <h3 class="section-title mb-2">Observations</h3>
        <p class="text-[13px] whitespace-pre-line" data-selectable>{d.key.notes}</p>
      </section>
    {/if}
  {/if}

  {#snippet footer()}
    {#if d}
      <button class="btn btn-primary" onclick={() => navigate('new-loan', { keyId: d!.key.id })}><HandCoins size={15} />Prêter cette clé</button>
      <button class="btn btn-secondary" onclick={() => (editOpen = true)}><Pencil size={14} />Modifier</button>
      <button class="btn btn-danger-ghost ml-auto" onclick={remove}><Trash2 size={14} />Supprimer</button>
    {/if}
  {/snippet}
</Drawer>

<KeyForm bind:open={editOpen} keyId={id} />
<ReturnDialog bind:open={returnOpen} loans={returning} />
