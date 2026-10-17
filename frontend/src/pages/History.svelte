<script lang="ts">
  // Historique complet des prêts, filtrable et exportable.
  import { FileSpreadsheet, History, RotateCcw } from '@lucide/svelte'
  import { Api, notifyDocument, run, type Borrower, type Key, type Loan } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import { fmtDate, loanStatusBadge, loanStatusLabel, plural } from '../lib/format'
  import { data, openDetail } from '../lib/state.svelte'

  let loans = $state<Loan[]>([])
  let borrowers = $state<Borrower[]>([])
  let keys = $state<Key[]>([])
  let filter = $state({ borrowerId: 0, keyId: 0, from: '', to: '', status: '' })

  void Promise.all([run(Api.ListBorrowers()), run(Api.ListKeys())]).then(([b, k]) => {
    borrowers = b ?? []
    keys = k ?? []
  })

  $effect(() => {
    void data.version
    const f = { ...filter, borrowerId: Number(filter.borrowerId), keyId: Number(filter.keyId) }
    void run(Api.GetHistory(f)).then((l) => (loans = l ?? []))
  })

  function resetFilters() {
    filter = { borrowerId: 0, keyId: 0, from: '', to: '', status: '' }
  }

  async function exportCSV() {
    const path = await run(Api.ExportHistoryCSV({ ...filter, borrowerId: Number(filter.borrowerId), keyId: Number(filter.keyId) }))
    if (path) notifyDocument('Historique exporté dans documents/.', path)
  }

  const statuses = [
    ['', 'Tous'],
    ['active', 'En cours'],
    ['overdue', 'En retard'],
    ['returned', 'Retournés'],
  ]
</script>

<PageHeader title="Historique" subtitle="Tous les prêts, en cours et terminés.">
  {#snippet actions()}
    <button class="btn btn-secondary" onclick={exportCSV}><FileSpreadsheet size={15} />Exporter (Excel)</button>
  {/snippet}
</PageHeader>

<div class="card mb-4 flex flex-wrap items-end gap-3 p-4">
  <label class="w-56">
    <span class="label">Détenteur</span>
    <select class="input" bind:value={filter.borrowerId}>
      <option value={0}>Tous</option>
      {#each borrowers as b}<option value={b.id}>{b.name}</option>{/each}
    </select>
  </label>
  <label class="w-56">
    <span class="label">Clé</span>
    <select class="input" bind:value={filter.keyId}>
      <option value={0}>Toutes</option>
      {#each keys as k}<option value={k.id}>{k.number} — {k.description}</option>{/each}
    </select>
  </label>
  <label>
    <span class="label">Remis du</span>
    <input class="input" type="date" bind:value={filter.from} />
  </label>
  <label>
    <span class="label">au</span>
    <input class="input" type="date" bind:value={filter.to} />
  </label>
  <div>
    <span class="label">Statut</span>
    <div class="segmented">
      {#each statuses as [value, label]}
        <button aria-pressed={filter.status === value} onclick={() => (filter.status = value)}>{label}</button>
      {/each}
    </div>
  </div>
  <button class="btn btn-ghost" onclick={resetFilters}><RotateCcw size={14} />Réinitialiser</button>
</div>

<div class="card overflow-hidden">
  {#if loans.length}
    <div class="max-h-[calc(100vh-20rem)] overflow-y-auto">
      <table class="table table-hover">
        <thead>
          <tr><th>Clé</th><th>Détenteur</th><th>Remise</th><th>Retour prévu</th><th>Retour</th><th>Statut</th><th>État au retour</th><th>Agent</th></tr>
        </thead>
        <tbody>
          {#each loans as l (l.id)}
            <tr onclick={() => openDetail('borrower', l.borrowerId)}>
              <td><span class="chip chip-key">{l.keyNumber}</span></td>
              <td class="font-medium">{l.borrowerName}</td>
              <td class="tabular-nums">{fmtDate(l.loanDate)}</td>
              <td class="text-muted tabular-nums">{fmtDate(l.plannedReturnDate)}</td>
              <td class="tabular-nums">{fmtDate(l.returnDate)}</td>
              <td><span class="badge {loanStatusBadge[l.status]}">{loanStatusLabel[l.status]}</span></td>
              <td class="text-muted">{l.condition || '—'}</td>
              <td class="text-muted">{l.createdBy || '—'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    <p class="border-t border-line px-4 py-2.5 text-[12.5px] text-muted">{plural(loans.length, 'prêt')}</p>
  {:else}
    <EmptyState icon={History} title="Aucun prêt" text="Aucun prêt ne correspond à ces critères." />
  {/if}
</div>
