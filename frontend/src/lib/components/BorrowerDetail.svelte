<script lang="ts">
  // Fiche d'un détenteur : coordonnées, clés détenues, accès, historique.
  import { FileText, HandCoins, Mail, Pencil, Phone, Trash2, Undo2, CopyX } from '@lucide/svelte'
  import { Api, notifyDocument, run, type BorrowerDetail, type Loan } from '../api'
  import { borrowerStatusLabel, fmtDate, fmtDays, loanStatusBadge, loanStatusLabel } from '../format'
  import { closeDetail, confirm, data, navigate, openDetail } from '../state.svelte'
  import BorrowerForm from './BorrowerForm.svelte'
  import Drawer from './Drawer.svelte'
  import ReturnDialog from './ReturnDialog.svelte'

  let { id }: { id: number } = $props()

  let d = $state<BorrowerDetail | null>(null)
  let editOpen = $state(false)
  let returnOpen = $state(false)
  let returning = $state<Loan[]>([])

  $effect(() => {
    void data.version
    void load(id)
  })

  async function load(borrowerId: number) {
    d = (await run(Api.GetBorrowerDetail(borrowerId))) ?? null
  }

  async function reissue() {
    if (!d) return
    const path = await run(Api.ReissueReceipt(d.borrower.id))
    if (path) notifyDocument('Bon de remise réédité.', path)
  }

  async function remove() {
    if (!d) return
    const ok = await confirm({
      title: `Supprimer ${d.borrower.name} ?`,
      message: 'Le détenteur et tout son historique de prêts seront définitivement supprimés.',
      confirmLabel: 'Supprimer',
      danger: true,
    })
    if (ok && (await run(Api.DeleteBorrower(d.borrower.id).then(() => true), 'Détenteur supprimé.'))) closeDetail()
  }

  function startReturn(loans: Loan[]) {
    returning = loans
    returnOpen = true
  }
</script>

<Drawer onclose={closeDetail}>
  {#snippet header()}
    {#if d}
      <span class="badge badge-neutral">{borrowerStatusLabel[d.borrower.status] ?? d.borrower.status}</span>
      <h2 class="mt-2 truncate text-[18px] font-semibold">{d.borrower.name}</h2>
      <div class="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-[13px] text-muted" data-selectable>
        {#if d.borrower.email}<span class="flex items-center gap-1.5"><Mail size={13} />{d.borrower.email}</span>{/if}
        {#if d.borrower.phone}<span class="flex items-center gap-1.5"><Phone size={13} />{d.borrower.phone}</span>{/if}
      </div>
    {:else}
      <h2 class="text-[18px] font-semibold">Chargement…</h2>
    {/if}
  {/snippet}

  {#if d}
    {#if d.redundancies.length}
      <div class="alert alert-info">
        <CopyX size={18} class="mt-0.5 shrink-0 text-info" />
        <div>
          <p class="font-semibold">Clés en doublon</p>
          <p class="text-muted">Plusieurs de ses clés ouvrent : {d.redundancies.map((r) => r.name).join(', ')}. Une clé pourrait être récupérée.</p>
        </div>
      </div>
    {/if}

    <section>
      <div class="mb-2 flex items-center justify-between">
        <h3 class="section-title">Clés détenues ({d.activeLoans.length})</h3>
        {#if d.activeLoans.length > 1}
          <button class="btn btn-ghost btn-sm" onclick={() => startReturn(d!.activeLoans)}><Undo2 size={14} />Tout rendre</button>
        {/if}
      </div>
      {#if d.activeLoans.length}
        <ul class="divide-y divide-line rounded-control border border-line">
          {#each d.activeLoans as l (l.id)}
            <li class="flex items-center gap-3 px-3 py-2.5">
              <button class="chip chip-key cursor-pointer hover:border-primary" onclick={() => openDetail('key', l.keyId)}>{l.keyNumber}</button>
              <div class="min-w-0 flex-1">
                <p class="truncate text-[13px]">{l.keyDescription}</p>
                <p class="text-[12px] text-muted">
                  depuis le {fmtDate(l.loanDate)} · {l.loanType}{l.plannedReturnDate ? ` · retour prévu le ${fmtDate(l.plannedReturnDate)}` : ''}
                </p>
              </div>
              {#if l.status === 'overdue'}<span class="badge badge-warning">En retard</span>{/if}
              <button class="btn btn-ghost btn-sm" onclick={() => startReturn([l])}><Undo2 size={14} />Retour</button>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-[13px] text-muted">Aucune clé en sa possession.</p>
      {/if}
    </section>

    {#if d.coveredAccesses.length}
      <section>
        <h3 class="section-title mb-2">Accès ouverts ({d.coveredAccesses.length})</h3>
        <div class="flex flex-wrap gap-1.5">
          {#each d.coveredAccesses as a (a.id)}
            <button class="chip cursor-pointer hover:border-primary" onclick={() => openDetail('access', a.id)}>{a.name}</button>
          {/each}
        </div>
      </section>
    {/if}

    <section>
      <h3 class="section-title mb-2">Historique</h3>
      {#if d.history.length}
        <ul class="space-y-1.5">
          {#each d.history.slice(0, 20) as l (l.id)}
            <li class="flex items-center gap-2 text-[13px]">
              <span class="w-20 text-muted tabular-nums">{fmtDate(l.loanDate)}</span>
              <span class="chip chip-key">{l.keyNumber}</span>
              <span class="min-w-0 flex-1"></span>
              <span class="badge {loanStatusBadge[l.status]}">{l.status === 'returned' ? `rendue le ${fmtDate(l.returnDate)}` : `${loanStatusLabel[l.status]} · ${fmtDays(l.days)}`}</span>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-[13px] text-muted">Aucun prêt enregistré.</p>
      {/if}
    </section>
  {/if}

  {#snippet footer()}
    {#if d}
      <button class="btn btn-primary" onclick={() => navigate('new-loan', { borrowerId: d!.borrower.id })}><HandCoins size={15} />Nouveau prêt</button>
      {#if d.activeLoans.length}
        <button class="btn btn-secondary" onclick={reissue} title="Bon complet couvrant toutes les clés détenues"><FileText size={14} />Rééditer le bon</button>
      {/if}
      <button class="btn btn-secondary" onclick={() => (editOpen = true)}><Pencil size={14} />Modifier</button>
      <button class="btn btn-danger-ghost ml-auto" onclick={remove} aria-label="Supprimer"><Trash2 size={14} /></button>
    {/if}
  {/snippet}
</Drawer>

<BorrowerForm bind:open={editOpen} borrowerId={id} />
<ReturnDialog bind:open={returnOpen} loans={returning} />
