<script lang="ts">
  // Détenteurs : liste, création, import / export CSV.
  import { Download, FileSpreadsheet, Plus, Upload, Users } from '@lucide/svelte'
  import { Api, notifyDocument, run, type Borrower } from '../lib/api'
  import BorrowerForm from '../lib/components/BorrowerForm.svelte'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import Modal from '../lib/components/Modal.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { borrowerStatusLabel, matches, plural } from '../lib/format'
  import { data, openDetail, refresh } from '../lib/state.svelte'

  let borrowers = $state<Borrower[]>([])
  let loaded = $state(false)
  let query = $state('')
  let status = $state('')
  let withKeys = $state(false)
  let formOpen = $state(false)
  let importOpen = $state(false)
  let report = $state<{ created: number; problems: string[] } | null>(null)

  $effect(() => {
    void data.version
    void run(Api.ListBorrowers()).then((b) => {
      borrowers = b ?? []
      loaded = true
    })
  })

  const filtered = $derived(
    borrowers.filter((b) => (!status || b.status === status) && (!withKeys || b.activeLoans > 0) && matches(query, b.name, b.email, b.phone)),
  )

  async function exportCSV() {
    const path = await run(Api.ExportBorrowersCSV())
    if (path) notifyDocument('Liste exportée dans documents/.', path)
  }
  async function template() {
    const path = await run(Api.DownloadBorrowersTemplate())
    if (path) notifyDocument('Modèle CSV enregistré dans documents/.', path)
  }
  async function importCSV() {
    const r = await run(Api.ImportBorrowersCSV())
    if (!r || r.created < 0) return
    report = r
    if (r.created > 0) refresh()
  }
</script>

<PageHeader title="Détenteurs" subtitle="{plural(borrowers.length, 'personne')} · {borrowers.filter((b) => b.activeLoans > 0).length} avec des clés">
  {#snippet actions()}
    <button class="btn btn-secondary" onclick={() => { report = null; importOpen = true }}><Upload size={15} />Importer</button>
    <button class="btn btn-secondary" onclick={exportCSV}><FileSpreadsheet size={15} />Exporter (Excel)</button>
    <button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouveau détenteur</button>
  {/snippet}
</PageHeader>

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Nom, e-mail, téléphone…" />
  <select class="input w-auto" bind:value={status} aria-label="Statut">
    <option value="">Tous les statuts</option>
    {#each Object.entries(borrowerStatusLabel) as [value, label]}<option {value}>{label}</option>{/each}
  </select>
  <label class="flex cursor-pointer items-center gap-2 text-[13px]">
    <input type="checkbox" class="size-4 accent-[var(--c-primary)]" bind:checked={withKeys} />Seulement ceux qui ont des clés
  </label>
</div>

<div class="card overflow-hidden">
  {#if filtered.length}
    <table class="table table-hover">
      <thead><tr><th>Nom</th><th>Statut</th><th>E-mail</th><th>Téléphone</th><th class="text-right">Clés détenues</th></tr></thead>
      <tbody>
        {#each filtered as b (b.id)}
          <tr onclick={() => openDetail('borrower', b.id)}>
            <td class="font-medium">{b.name}</td>
            <td><span class="badge badge-neutral">{borrowerStatusLabel[b.status] ?? b.status}</span></td>
            <td class="text-muted">{b.email || '—'}</td>
            <td class="text-muted">{b.phone || '—'}</td>
            <td class="text-right">{#if b.activeLoans}<span class="badge badge-primary">{plural(b.activeLoans, 'clé')}</span>{:else}<span class="text-muted">—</span>{/if}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if loaded}
    <EmptyState icon={Users} title={borrowers.length ? 'Aucun détenteur ne correspond' : 'Aucun détenteur enregistré'} text={borrowers.length ? 'Modifiez la recherche ou les filtres.' : 'Ajoutez les personnes une par une ou importez une liste.'}>
      {#if !borrowers.length}
        <button class="btn btn-secondary" onclick={() => (importOpen = true)}><Upload size={15} />Importer une liste</button>
        <button class="btn btn-primary" onclick={() => (formOpen = true)}><Plus size={16} />Nouveau détenteur</button>
      {/if}
    </EmptyState>
  {/if}
</div>

<BorrowerForm bind:open={formOpen} onsaved={(id) => openDetail('borrower', id)} />

<Modal bind:open={importOpen} title="Importer des détenteurs" subtitle="À partir d'un fichier CSV (Excel : « Enregistrer sous » → CSV point-virgule)." size="md">
  {#if report}
    <div class="alert {report.created > 0 ? 'alert-success' : 'alert-warning'}">
      <p><strong>{plural(report.created, 'détenteur créé', 'détenteurs créés')}.</strong></p>
    </div>
    {#if report.problems.length}
      <p class="mt-4 mb-2 text-[13px] font-medium">{plural(report.problems.length, 'ligne ignorée', 'lignes ignorées')} :</p>
      <ul class="max-h-48 list-inside list-disc overflow-y-auto text-[13px] text-muted" data-selectable>
        {#each report.problems as p}<li>{p}</li>{/each}
      </ul>
    {/if}
  {:else}
    <ol class="space-y-4 text-[13.5px]">
      <li class="flex gap-3">
        <span class="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary-soft text-[12px] font-semibold text-primary">1</span>
        <div>
          <p class="font-medium">Téléchargez le modèle et remplissez-le dans Excel</p>
          <p class="text-muted">Colonnes : Nom (obligatoire), Email, Statut (permanent, contractuel, intervenant, entreprise), Téléphone.</p>
          <button class="btn btn-secondary btn-sm mt-2" onclick={template}><Download size={14} />Télécharger le modèle</button>
        </div>
      </li>
      <li class="flex gap-3">
        <span class="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary-soft text-[12px] font-semibold text-primary">2</span>
        <div>
          <p class="font-medium">Choisissez le fichier rempli</p>
          <p class="text-muted">Les lignes incomplètes sont ignorées et listées à la fin, sans bloquer les autres.</p>
        </div>
      </li>
    </ol>
  {/if}
  {#snippet footer()}
    {#if report}
      <button class="btn btn-primary" onclick={() => (importOpen = false)}>Terminer</button>
    {:else}
      <button class="btn btn-secondary" onclick={() => (importOpen = false)}>Annuler</button>
      <button class="btn btn-primary" onclick={importCSV}><Upload size={15} />Choisir le fichier…</button>
    {/if}
  {/snippet}
</Modal>
