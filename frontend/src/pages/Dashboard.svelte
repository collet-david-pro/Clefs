<script lang="ts">
  // Tableau de bord orienté actions : chiffres clés, alertes à traiter,
  // derniers mouvements.
  import { ArrowRight, CircleCheck, Clock, CopyX, DoorOpen, HandCoins, KeyRound, TriangleAlert, Undo2, Upload, Users } from '@lucide/svelte'
  import { Api, run, type Dashboard } from '../lib/api'
  import { fmtDate, fmtLongDate, loanStatusBadge, loanStatusLabel, plural } from '../lib/format'
  import { data, navigate, openDetail } from '../lib/state.svelte'

  let d = $state<Dashboard | null>(null)

  $effect(() => {
    void data.version
    void load()
  })

  async function load() {
    d = (await run(Api.GetDashboard())) ?? null
  }

  const alertCount = $derived(d ? d.overdue.length + d.anomalies.length + d.redundancies.length : 0)
  const isEmpty = $derived(d !== null && d.keyTypes === 0 && d.borrowers === 0 && d.accesses === 0)
</script>

<div class="mb-7 flex flex-wrap items-end justify-between gap-4">
  <div>
    <p class="text-[13px] font-medium text-muted first-letter:uppercase">{fmtLongDate()}</p>
    <h1 class="mt-1 text-[24px] font-semibold tracking-tight">Tableau de bord</h1>
  </div>
  <div class="flex gap-2">
    <button class="btn btn-secondary btn-lg" onclick={() => navigate('loans')}><Undo2 size={17} />Enregistrer un retour</button>
    <button class="btn btn-primary btn-lg" onclick={() => navigate('new-loan')}><HandCoins size={17} />Nouveau prêt</button>
  </div>
</div>

{#if d}
  {#if isEmpty}
    <div class="card mb-6 flex items-center gap-6 p-6">
      <div class="flex size-12 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary"><Upload size={22} /></div>
      <div class="flex-1">
        <p class="text-[15px] font-semibold">La base est vide</p>
        <p class="mt-0.5 text-[13px] text-muted">Importez la base de votre ancienne version pour retrouver toutes vos données, ou commencez par créer vos bâtiments, accès et clés.</p>
      </div>
      <button class="btn btn-primary" onclick={() => navigate('import')}>Importer une base</button>
      <button class="btn btn-secondary" onclick={() => navigate('buildings')}>Commencer à zéro</button>
    </div>
  {/if}

  <div class="mb-6 grid grid-cols-4 gap-4">
    <button class="card cursor-pointer p-5 text-left transition-shadow hover:shadow-pop" onclick={() => navigate('loans')}>
      <div class="flex items-center justify-between text-muted"><span class="text-[13px] font-medium">Exemplaires sortis</span><HandCoins size={17} /></div>
      <p class="mt-3 text-[28px] font-semibold tabular-nums">{d.loanedCopies}<span class="text-[15px] font-normal text-muted"> / {d.keyCopies}</span></p>
      <p class="mt-1 text-[12.5px] text-muted">chez {plural(d.borrowersWithKeys, 'détenteur')}</p>
    </button>
    <button class="card cursor-pointer p-5 text-left transition-shadow hover:shadow-pop" onclick={() => navigate('keys')}>
      <div class="flex items-center justify-between text-muted"><span class="text-[13px] font-medium">Disponibles</span><KeyRound size={17} /></div>
      <p class="mt-3 text-[28px] font-semibold text-success tabular-nums">{d.availableCopies}</p>
      <p class="mt-1 text-[12.5px] text-muted">exemplaires prêtables · {plural(d.keyTypes, 'clé')}</p>
    </button>
    <button class="card cursor-pointer p-5 text-left transition-shadow hover:shadow-pop" onclick={() => navigate('borrowers')}>
      <div class="flex items-center justify-between text-muted"><span class="text-[13px] font-medium">Détenteurs</span><Users size={17} /></div>
      <p class="mt-3 text-[28px] font-semibold tabular-nums">{d.borrowers}</p>
      <p class="mt-1 text-[12.5px] text-muted">{d.borrowersWithKeys} avec des clés</p>
    </button>
    <button class="card cursor-pointer p-5 text-left transition-shadow hover:shadow-pop" onclick={() => navigate('accesses')}>
      <div class="flex items-center justify-between text-muted"><span class="text-[13px] font-medium">Accès référencés</span><DoorOpen size={17} /></div>
      <p class="mt-3 text-[28px] font-semibold tabular-nums">{d.accesses}</p>
      <p class="mt-1 text-[12.5px] text-muted">portes, portails et zones</p>
    </button>
  </div>

  <div class="grid grid-cols-[1.25fr_1fr] gap-6">
    <section class="card overflow-hidden">
      <header class="flex items-center justify-between border-b border-line px-5 py-3.5">
        <h2 class="text-[14.5px] font-semibold">À traiter</h2>
        {#if alertCount}<span class="badge badge-danger">{alertCount}</span>{/if}
      </header>
      {#if alertCount === 0}
        <div class="flex items-center gap-3 px-5 py-8">
          <CircleCheck size={22} class="text-success" />
          <div>
            <p class="font-medium">Tout est en ordre</p>
            <p class="text-[13px] text-muted">Aucun retard, aucune erreur d'inventaire, aucune clé en doublon.</p>
          </div>
        </div>
      {:else}
        <ul class="divide-y divide-line">
          {#each d.anomalies as a (a.keyId)}
            <li>
              <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-3 text-left hover:bg-surface-2" onclick={() => openDetail('key', a.keyId)}>
                <span class="flex size-8 shrink-0 items-center justify-center rounded-full bg-danger-soft text-danger"><TriangleAlert size={15} /></span>
                <div class="min-w-0 flex-1">
                  <p class="text-[13.5px] font-medium">Erreur d'inventaire — clé {a.keyNumber}</p>
                  <p class="text-[12.5px] text-muted">{a.loaned} sortie(s) pour {a.total - a.reserve} utilisable(s) : vérifier le stock.</p>
                </div>
                <ArrowRight size={15} class="text-muted" />
              </button>
            </li>
          {/each}
          {#each d.overdue as l (l.id)}
            <li>
              <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-3 text-left hover:bg-surface-2" onclick={() => openDetail('borrower', l.borrowerId)}>
                <span class="flex size-8 shrink-0 items-center justify-center rounded-full bg-warning-soft text-warning"><Clock size={15} /></span>
                <div class="min-w-0 flex-1">
                  <p class="text-[13.5px] font-medium">Retour en retard — {l.borrowerName}</p>
                  <p class="text-[12.5px] text-muted">Clé {l.keyNumber}, retour prévu le {fmtDate(l.plannedReturnDate)}.</p>
                </div>
                <ArrowRight size={15} class="text-muted" />
              </button>
            </li>
          {/each}
          {#each d.redundancies as r (r.borrowerId)}
            <li>
              <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-3 text-left hover:bg-surface-2" onclick={() => openDetail('borrower', r.borrowerId)}>
                <span class="flex size-8 shrink-0 items-center justify-center rounded-full bg-info-soft text-info"><CopyX size={15} /></span>
                <div class="min-w-0 flex-1">
                  <p class="text-[13.5px] font-medium">Clés en doublon — {r.borrowerName}</p>
                  <p class="truncate text-[12.5px] text-muted">Plusieurs clés ouvrent : {r.accesses.join(', ')}</p>
                </div>
                <ArrowRight size={15} class="text-muted" />
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    <section class="card overflow-hidden">
      <header class="flex items-center justify-between border-b border-line px-5 py-3.5">
        <h2 class="text-[14.5px] font-semibold">Derniers mouvements</h2>
        <button class="btn btn-ghost btn-sm" onclick={() => navigate('history')}>Tout voir<ArrowRight size={14} /></button>
      </header>
      {#if d.recentLoans.length}
        <ul class="divide-y divide-line">
          {#each d.recentLoans as l (l.id)}
            <li>
              <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-2.5 text-left hover:bg-surface-2" onclick={() => openDetail('borrower', l.borrowerId)}>
                <span class="chip chip-key">{l.keyNumber}</span>
                <span class="min-w-0 flex-1 truncate text-[13px]">{l.borrowerName}</span>
                <span class="badge {loanStatusBadge[l.status]}">{l.status === 'returned' ? 'Rendue' : loanStatusLabel[l.status]}</span>
                <span class="w-20 text-right text-[12px] text-muted">{fmtDate(l.loanDate)}</span>
              </button>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="px-5 py-8 text-center text-[13px] text-muted">Aucun prêt enregistré pour le moment.</p>
      {/if}
    </section>
  </div>
{/if}
