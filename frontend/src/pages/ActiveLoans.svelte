<script lang="ts">
  // Prêts en cours, regroupés par détenteur : retours et réédition du bon.
  import { Archive, ArrowLeftRight, FileSpreadsheet, FileText, HandCoins, Undo2 } from '@lucide/svelte'
  import { Api, notifyDocument, run, type Loan } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import ReturnDialog from '../lib/components/ReturnDialog.svelte'
  import SearchInput from '../lib/components/SearchInput.svelte'
  import { fmtDate, fmtDays, matches, plural } from '../lib/format'
  import { data, navigate, openDetail } from '../lib/state.svelte'

  let loans = $state<Loan[]>([])
  let loaded = $state(false)
  let query = $state('')
  let onlyOverdue = $state(false)
  let returnOpen = $state(false)
  let returning = $state<Loan[]>([])

  $effect(() => {
    void data.version
    void load()
  })

  async function load() {
    loans = (await run(Api.ListActiveLoans())) ?? []
    loaded = true
  }

  const groups = $derived.by(() => {
    const map = new Map<number, Loan[]>()
    for (const l of loans) {
      if (onlyOverdue && l.status !== 'overdue') continue
      if (!matches(query, l.borrowerName, l.keyNumber, l.keyDescription)) continue
      if (!map.has(l.borrowerId)) map.set(l.borrowerId, [])
      map.get(l.borrowerId)!.push(l)
    }
    return [...map.values()].sort((a, b) => a[0].borrowerName.localeCompare(b[0].borrowerName, 'fr'))
  })
  const overdueCount = $derived(loans.filter((l) => l.status === 'overdue').length)

  function startReturn(list: Loan[]) {
    returning = list
    returnOpen = true
  }

  async function reissue(borrowerId: number) {
    const path = await run(Api.ReissueReceipt(borrowerId))
    if (path) notifyDocument('Bon de remise réédité.', path)
  }

  let archiving = $state(false)
  async function exportAll() {
    archiving = true
    const path = await run(Api.ExportAllReceipts())
    archiving = false
    if (path) notifyDocument('Tous les bons de remise sont réunis dans un seul PDF (documents/).', path)
  }

  async function exportCSV() {
    const path = await run(Api.ExportActiveLoansCSV())
    if (path) notifyDocument('Export CSV enregistré dans documents/.', path)
  }
</script>

<PageHeader title="Prêts en cours" subtitle="{plural(loans.length, 'clé sortie', 'clés sorties')} · {plural(groups.length, 'détenteur')}">
  {#snippet actions()}
    <button class="btn btn-secondary" disabled={archiving || !loans.length} onclick={exportAll} title="Un seul PDF : sommaire puis un bon par détenteur, pour l'impression ou l'archivage"><Archive size={15} />{archiving ? 'Édition…' : 'Éditer tous les bons (PDF)'}</button>
    <button class="btn btn-secondary" onclick={exportCSV}><FileSpreadsheet size={15} />Exporter (Excel)</button>
    <button class="btn btn-primary" onclick={() => navigate('new-loan')}><HandCoins size={15} />Nouveau prêt</button>
  {/snippet}
</PageHeader>

<div class="mb-4 flex flex-wrap items-center gap-3">
  <SearchInput bind:value={query} placeholder="Personne ou clé…" />
  {#if overdueCount}
    <label class="flex cursor-pointer items-center gap-2 text-[13px]">
      <input type="checkbox" class="size-4 accent-[var(--c-primary)]" bind:checked={onlyOverdue} />
      Seulement les retards <span class="badge badge-warning">{overdueCount}</span>
    </label>
  {/if}
</div>

{#if loaded && groups.length === 0}
  <div class="card">
    <EmptyState icon={ArrowLeftRight} title={loans.length ? 'Aucun prêt ne correspond' : 'Aucune clé sortie'} text={loans.length ? 'Modifiez la recherche.' : 'Toutes les clés sont rangées.'} />
  </div>
{/if}

<div class="space-y-4">
  {#each groups as group (group[0].borrowerId)}
    {@const b = group[0]}
    <section class="card overflow-hidden">
      <header class="flex items-center gap-3 border-b border-line px-5 py-3">
        <button class="flex cursor-pointer items-center gap-3 text-left" onclick={() => openDetail('borrower', b.borrowerId)}>
          <span class="flex size-8 items-center justify-center rounded-full bg-primary-soft text-[13px] font-semibold text-primary">{b.borrowerName.charAt(0).toUpperCase()}</span>
          <span class="text-[14px] font-semibold hover:text-primary">{b.borrowerName}</span>
        </button>
        <span class="badge badge-neutral">{plural(group.length, 'clé')}</span>
        <div class="flex-1"></div>
        <button class="btn btn-ghost btn-sm" onclick={() => reissue(b.borrowerId)}><FileText size={14} />Rééditer le bon</button>
        {#if group.length > 1}<button class="btn btn-secondary btn-sm" onclick={() => startReturn(group)}><Undo2 size={14} />Tout rendre</button>{/if}
      </header>
      <table class="table">
        <tbody>
          {#each group as l (l.id)}
            <tr>
              <td class="w-28"><button class="chip chip-key cursor-pointer hover:border-primary" onclick={() => openDetail('key', l.keyId)}>{l.keyNumber}</button></td>
              <td class="truncate">{l.keyDescription}</td>
              <td class="w-44 text-muted">depuis le {fmtDate(l.loanDate)}<br /><span class="text-[12px]">{fmtDays(l.days)} · {l.loanType}</span></td>
              <td class="w-40">
                {#if l.status === 'overdue'}
                  <span class="badge badge-warning">En retard depuis le {fmtDate(l.plannedReturnDate)}</span>
                {:else if l.plannedReturnDate}
                  <span class="text-[12.5px] text-muted">retour le {fmtDate(l.plannedReturnDate)}</span>
                {/if}
              </td>
              <td class="w-28 text-right"><button class="btn btn-secondary btn-sm" onclick={() => startReturn([l])}><Undo2 size={14} />Retour</button></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </section>
  {/each}
</div>

<ReturnDialog bind:open={returnOpen} loans={returning} />
