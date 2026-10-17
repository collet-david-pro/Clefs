<script lang="ts">
  // Prêt par besoin : on choisit la personne et les portes à ouvrir,
  // l'application propose le plus petit trousseau qui les couvre.
  import { CircleCheck, CopyX, FileText, Plus, Sparkles, TriangleAlert, UserPlus, X } from '@lucide/svelte'
  import { Api, run, type Access, type Borrower, type Key, type LoanResult, type Suggestion } from '../lib/api'
  import AccessPicker from '../lib/components/AccessPicker.svelte'
  import BorrowerForm from '../lib/components/BorrowerForm.svelte'
  import Modal from '../lib/components/Modal.svelte'
  import { borrowerStatusLabel, matches, plural, todayISO } from '../lib/format'
  import { appSettings, navigate, refresh, route } from '../lib/state.svelte'

  let borrowers = $state<Borrower[]>([])
  let accesses = $state<Access[]>([])
  let keys = $state<Key[]>([])

  let borrowerId = $state(route.params.borrowerId ?? 0)
  let borrowerQuery = $state('')
  let borrowerListOpen = $state(false)
  let accessIds = $state<number[]>([])
  let excluded = $state<number[]>([])
  let manual = $state<number[]>(route.params.keyId ? [route.params.keyId] : [])
  let keyQuery = $state('')
  let keyListOpen = $state(false)
  // Valeurs par défaut choisies dans les Paramètres.
  const defaultReturn = () => (appSettings.defaultReturnDays > 0 ? todayISO(appSettings.defaultReturnDays) : '')
  const defaultType = (): 'ponctuel' | 'permanent' => (appSettings.defaultLoanType === 'permanent' ? 'permanent' : 'ponctuel')
  let plannedReturn = $state(defaultReturn())
  let loanType = $state<'ponctuel' | 'permanent'>(defaultType())

  let suggestion = $state<Suggestion | null>(null)
  let saving = $state(false)
  let result = $state<LoanResult | null>(null)
  let resultOpen = $state(false)
  let borrowerFormOpen = $state(false)

  void load()
  async function load() {
    const [b, a, k] = await Promise.all([run(Api.ListBorrowers()), run(Api.ListAccesses()), run(Api.ListKeys())])
    borrowers = b ?? []
    accesses = a ?? []
    keys = k ?? []
  }

  const borrower = $derived(borrowers.find((b) => b.id === borrowerId))
  const borrowerMatches = $derived(borrowers.filter((b) => matches(borrowerQuery, b.name, b.email)).slice(0, 30))

  // Recalcul du trousseau à chaque changement (léger délai pour regrouper les clics).
  let timer: ReturnType<typeof setTimeout> | undefined
  $effect(() => {
    const req = { borrowerId, accessIds: [...accessIds], excludedKeyIds: [...excluded], manualKeyIds: [...manual] }
    clearTimeout(timer)
    timer = setTimeout(async () => {
      suggestion = (await run(Api.SuggestKeys(req))) ?? null
    }, 120)
  })

  const trousseau = $derived(suggestion?.keys ?? [])
  const inTrousseau = $derived(new Set(trousseau.map((s) => s.key.id)))
  const addableKeys = $derived(keys.filter((k) => !inTrousseau.has(k.id) && matches(keyQuery, k.number, k.description)).slice(0, 30))
  const canValidate = $derived(borrowerId > 0 && trousseau.length > 0 && !saving)

  function onAccessesChanged() {
    // Nouvelle situation : les clés retirées à la main peuvent être reproposées.
    excluded = []
  }

  function removeKey(id: number, suggested: boolean) {
    if (suggested) excluded = [...excluded, id]
    manual = manual.filter((x) => x !== id)
  }

  function addKey(id: number) {
    excluded = excluded.filter((x) => x !== id)
    if (!manual.includes(id)) manual = [...manual, id]
    keyQuery = ''
    keyListOpen = false
  }

  function selectBorrower(b: Borrower) {
    borrowerId = b.id
    borrowerQuery = ''
    borrowerListOpen = false
  }

  async function validate() {
    if (!canValidate) return
    saving = true
    const res = await run(
      Api.CreateLoan({ borrowerId, keyIds: trousseau.map((s) => s.key.id), plannedReturn, loanType }),
    )
    saving = false
    if (res) {
      result = res
      resultOpen = true
      refresh()
    }
  }

  function reset() {
    resultOpen = false
    borrowerId = 0
    accessIds = []
    excluded = []
    manual = []
    plannedReturn = defaultReturn()
    loanType = defaultType()
    void load()
  }
</script>

<div class="mb-6">
  <h1 class="text-[22px] font-semibold tracking-tight">Nouveau prêt</h1>
  <p class="mt-1 text-[13.5px] text-muted">Choisissez la personne et les portes qu'elle doit ouvrir : le trousseau le plus réduit est calculé automatiquement.</p>
</div>

<div class="grid grid-cols-[1fr_420px] items-start gap-6">
  <div class="space-y-6">
    <!-- 1. Détenteur -->
    <section class="card p-5">
      <h2 class="mb-3 flex items-center gap-2.5 text-[14.5px] font-semibold">
        <span class="flex size-6 items-center justify-center rounded-full bg-primary text-[12px] text-primary-fg">1</span>À qui remettre les clés ?
      </h2>
      {#if borrower}
        <div class="flex items-center gap-3 rounded-control border border-primary/40 bg-primary-soft px-4 py-3">
          <div class="flex size-9 items-center justify-center rounded-full bg-primary text-[13px] font-semibold text-primary-fg">{borrower.name.charAt(0).toUpperCase()}</div>
          <div class="min-w-0 flex-1">
            <p class="font-semibold">{borrower.name}</p>
            <p class="text-[12.5px] text-muted">
              {borrowerStatusLabel[borrower.status] ?? borrower.status}{borrower.activeLoans ? ` · détient déjà ${plural(borrower.activeLoans, 'clé')}` : ''}
            </p>
          </div>
          <button class="btn btn-secondary btn-sm" onclick={() => (borrowerId = 0)}>Changer</button>
        </div>
      {:else}
        <div class="flex gap-2">
          <div class="relative flex-1">
            <input
              class="input h-10"
              placeholder="Tapez un nom…"
              bind:value={borrowerQuery}
              onfocus={() => (borrowerListOpen = true)}
              onblur={() => setTimeout(() => (borrowerListOpen = false), 150)}
            />
            {#if borrowerListOpen}
              <ul class="absolute top-full right-0 left-0 z-10 mt-1 max-h-72 overflow-y-auto rounded-control border border-line bg-surface p-1 shadow-pop">
                {#each borrowerMatches as b (b.id)}
                  <li>
                    <button class="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-surface-2" onmousedown={() => selectBorrower(b)}>
                      <span class="flex-1 text-[13.5px]">{b.name}</span>
                      <span class="text-[12px] text-muted">{borrowerStatusLabel[b.status] ?? b.status}</span>
                      {#if b.activeLoans}<span class="badge badge-primary">{plural(b.activeLoans, 'clé')}</span>{/if}
                    </button>
                  </li>
                {:else}
                  <li class="px-3 py-3 text-[13px] text-muted">Aucun détenteur trouvé.</li>
                {/each}
              </ul>
            {/if}
          </div>
          <button class="btn btn-secondary h-10" onclick={() => (borrowerFormOpen = true)}><UserPlus size={16} />Nouveau détenteur</button>
        </div>
      {/if}
    </section>

    <!-- 2. Accès -->
    <section class="card p-5">
      <h2 class="mb-3 flex items-center gap-2.5 text-[14.5px] font-semibold">
        <span class="flex size-6 items-center justify-center rounded-full bg-primary text-[12px] text-primary-fg">2</span>Quelles portes doit-elle pouvoir ouvrir ?
      </h2>
      <AccessPicker {accesses} bind:selected={accessIds} onchange={onAccessesChanged} class="h-[420px]" />
    </section>
  </div>

  <!-- 3. Trousseau -->
  <aside class="card sticky top-0 flex max-h-[calc(100vh-7.5rem)] flex-col">
    <div class="border-b border-line px-5 py-4">
      <h2 class="flex items-center gap-2.5 text-[14.5px] font-semibold">
        <span class="flex size-6 items-center justify-center rounded-full bg-primary text-[12px] text-primary-fg">3</span>Trousseau à remettre
        {#if trousseau.length}<span class="badge badge-primary ml-auto">{plural(trousseau.length, 'clé')}</span>{/if}
      </h2>
    </div>

    <div class="min-h-0 flex-1 space-y-3 overflow-y-auto px-5 py-4">
      {#if suggestion?.uncoverable.length}
        <div class="alert alert-danger">
          <TriangleAlert size={16} class="mt-0.5 shrink-0 text-danger" />
          <p>Aucune clé disponible n'ouvre : <strong>{suggestion.uncoverable.map((a) => a.name).join(', ')}</strong></p>
        </div>
      {/if}
      {#if suggestion?.redundancies.length}
        <div class="alert alert-info">
          <CopyX size={16} class="mt-0.5 shrink-0 text-info" />
          <p>Cette personne peut déjà ouvrir : <strong>{suggestion.redundancies.join(', ')}</strong></p>
        </div>
      {/if}

      {#each trousseau as s (s.key.id)}
        <div class="rounded-control border border-line p-3">
          <div class="flex items-start gap-2">
            <span class="chip chip-key">{s.key.number}</span>
            <div class="min-w-0 flex-1">
              <p class="truncate text-[13px] font-medium">{s.key.description || 'Sans description'}</p>
              <p class="text-[12px] text-muted">
                {#if s.key.available <= 0}<span class="text-danger">plus d'exemplaire disponible</span>{:else}{plural(s.key.available, 'exemplaire')} disponible{s.key.available > 1 ? 's' : ''}{/if}
                {#if s.key.storageLocation}· {s.key.storageLocation}{/if}
              </p>
            </div>
            {#if s.suggested}<span class="badge badge-primary" title="Proposée automatiquement"><Sparkles size={11} />Proposée</span>{:else}<span class="badge badge-neutral">Ajoutée</span>{/if}
            <button class="-mr-1 cursor-pointer text-muted hover:text-danger" onclick={() => removeKey(s.key.id, s.suggested)} aria-label="Retirer la clé {s.key.number}"><X size={16} /></button>
          </div>
          {#if s.covers.length}
            <p class="mt-2 text-[12px] text-muted">Ouvre : {s.covers.join(', ')}</p>
          {/if}
        </div>
      {:else}
        <div class="py-8 text-center text-[13px] text-muted">
          {accessIds.length ? 'Aucune clé disponible pour ces accès.' : 'Cochez les portes à ouvrir : les clés nécessaires apparaîtront ici.'}
        </div>
      {/each}

      <div class="relative">
        <div class="flex items-center gap-2">
          <Plus size={15} class="text-muted" />
          <input
            class="input h-8 text-[13px]"
            placeholder="Ajouter une clé à la main…"
            bind:value={keyQuery}
            onfocus={() => (keyListOpen = true)}
            onblur={() => setTimeout(() => (keyListOpen = false), 150)}
          />
        </div>
        {#if keyListOpen}
          <ul class="absolute right-0 bottom-full left-0 z-10 mb-1 max-h-64 overflow-y-auto rounded-control border border-line bg-surface p-1 shadow-pop">
            {#each addableKeys as k (k.id)}
              <li>
                <button class="flex w-full cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-left hover:bg-surface-2" onmousedown={() => addKey(k.id)}>
                  <span class="chip chip-key">{k.number}</span>
                  <span class="min-w-0 flex-1 truncate text-[13px]">{k.description}</span>
                  <span class="text-[12px] {k.available > 0 ? 'text-muted' : 'text-danger'}">{k.available} dispo</span>
                </button>
              </li>
            {:else}
              <li class="px-3 py-3 text-[13px] text-muted">Aucune clé trouvée.</li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>

    <div class="space-y-3 border-t border-line px-5 py-4">
      <div class="grid grid-cols-2 gap-3">
        <div>
          <span class="label">Type de prêt</span>
          <div class="segmented flex">
            <button class="flex-1" aria-pressed={loanType === 'ponctuel'} onclick={() => (loanType = 'ponctuel')}>Ponctuel</button>
            <button class="flex-1" aria-pressed={loanType === 'permanent'} onclick={() => (loanType = 'permanent')}>Permanent</button>
          </div>
        </div>
        <label>
          <span class="label">Retour prévu</span>
          <input class="input" type="date" min={todayISO()} bind:value={plannedReturn} />
        </label>
      </div>
      <button class="btn btn-primary btn-lg w-full" disabled={!canValidate} onclick={validate}>
        <FileText size={17} />
        {trousseau.length ? `Remettre ${plural(trousseau.length, 'clé')} et éditer le bon` : 'Remettre les clés'}
      </button>
      {#if !borrowerId && trousseau.length}<p class="text-center text-[12px] text-warning">Choisissez d'abord la personne (étape 1).</p>{/if}
    </div>
  </aside>
</div>

<BorrowerForm bind:open={borrowerFormOpen} onsaved={async (id) => { await load(); borrowerId = id }} />

<Modal bind:open={resultOpen} title="Prêt enregistré" size="sm" dismissable={false}>
  {#if result}
    <div class="flex items-start gap-3">
      <CircleCheck size={24} class="shrink-0 text-success" />
      <p class="text-[14px]"><strong>{plural(result.count, 'clé remise', 'clés remises')}</strong> à {result.borrowerName}.</p>
    </div>
    {#if result.receiptPath}
      <button class="btn btn-secondary mt-4 w-full" onclick={() => run(Api.OpenPath(result!.receiptPath))}><FileText size={16} />Ouvrir le bon de remise à imprimer</button>
    {:else if result.receiptError}
      <div class="alert alert-warning mt-4">Le bon de remise n'a pas pu être généré : {result.receiptError}</div>
    {/if}
    {#each result.anomalies as a (a.keyId)}
      <div class="alert alert-danger mt-3">
        <TriangleAlert size={16} class="mt-0.5 shrink-0 text-danger" />
        <p>Erreur d'inventaire, vérifier le stock : clé <strong>{a.keyNumber}</strong> ({a.loaned} sortie(s) pour {a.total - a.reserve} utilisable(s)).</p>
      </div>
    {/each}
  {/if}
  {#snippet footer()}
    <button class="btn btn-secondary" onclick={() => { resultOpen = false; navigate('dashboard') }}>Tableau de bord</button>
    <button class="btn btn-primary" onclick={reset}>Nouveau prêt</button>
  {/snippet}
</Modal>
