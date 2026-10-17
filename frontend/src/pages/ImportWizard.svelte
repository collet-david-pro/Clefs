<script lang="ts">
  // Assistant d'import : reprise des données d'une version précédente
  // (V3, V2 ou V1) dans une nouvelle installation, guidé pas à pas.
  import { ArrowLeft, ArrowRight, CircleCheck, Database, FileSearch, FolderSearch, TriangleAlert, Upload } from '@lucide/svelte'
  import { Api, errorMessage, run, type AppInfo, type DatabaseFileInfo, type ImportResult } from '../lib/api'
  import { plural } from '../lib/format'
  import { confirm, navigate, refresh } from '../lib/state.svelte'

  let step = $state(1)
  let info = $state<AppInfo | null>(null)
  let file = $state<DatabaseFileInfo | null>(null)
  let inspectError = $state('')
  let importing = $state(false)
  let result = $state<ImportResult | null>(null)
  let importError = $state('')

  void run(Api.GetAppInfo()).then((i) => (info = i ?? null))

  const steps = ['Avant de commencer', 'Choisir la base', 'Vérifier', 'Terminé']

  async function choose() {
    inspectError = ''
    const path = await run(Api.ChooseDatabaseFile())
    if (!path) return
    try {
      file = await Api.InspectDatabaseFile(path)
      step = 3
    } catch (e) {
      inspectError = errorMessage(e)
    }
  }

  async function doImport() {
    if (!file) return
    if (file.kind === 'v3' && info && !info.isEmpty) {
      const ok = await confirm({
        title: 'Remplacer les données actuelles ?',
        message: "La base actuelle n'est pas vide : elle sera remplacée par la base importée.\nUne copie de sécurité est faite automatiquement avant.",
        confirmLabel: 'Remplacer et importer',
        danger: true,
      })
      if (!ok) return
    }
    importing = true
    importError = ''
    try {
      result = await Api.ImportDatabaseFile(file.path)
      step = 4
      refresh()
    } catch (e) {
      importError = errorMessage(e)
    }
    importing = false
  }

  async function startEmpty() {
    navigate('buildings')
  }

  async function loadDemo() {
    if ((await run(Api.LoadDemoData().then(() => true), 'Données de démonstration chargées.')) === true) navigate('dashboard')
  }
</script>

<div class="mx-auto max-w-3xl">
  <div class="mb-6">
    <h1 class="text-[22px] font-semibold tracking-tight">Importer une base existante</h1>
    <p class="mt-1 text-[13.5px] text-muted">Retrouvez toutes les données de votre ancienne version : clés, accès, détenteurs et historique des prêts.</p>
  </div>

  <!-- Étapes -->
  <ol class="mb-6 flex items-center gap-2">
    {#each steps as label, i}
      {@const n = i + 1}
      <li class="flex flex-1 items-center gap-2">
        <span class="flex size-7 shrink-0 items-center justify-center rounded-full text-[12.5px] font-semibold {n < step ? 'bg-success text-white' : n === step ? 'bg-primary text-primary-fg' : 'bg-surface-2 text-muted'}">
          {#if n < step}<CircleCheck size={15} />{:else}{n}{/if}
        </span>
        <span class="text-[12.5px] font-medium {n === step ? '' : 'text-muted'}">{label}</span>
        {#if n < steps.length}<span class="h-px flex-1 bg-line"></span>{/if}
      </li>
    {/each}
  </ol>

  <div class="card p-7">
    {#if step === 1}
      <h2 class="text-[16px] font-semibold">Avant de commencer</h2>
      <ul class="mt-4 space-y-4 text-[13.5px]">
        <li class="flex gap-3">
          <FolderSearch size={20} class="shrink-0 text-primary" />
          <div>
            <p class="font-medium">Repérez le fichier <code class="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[12.5px]">clefs.db</code> de l'ancienne version</p>
            <p class="text-muted">Il se trouve dans le dossier de l'ancienne application, à côté de son exécutable (par exemple <span class="font-mono text-[12.5px]">C:\Clefs\clefs.db</span> ou sur le partage réseau). Une sauvegarde du dossier <span class="font-mono text-[12.5px]">backups\</span> convient aussi.</p>
          </div>
        </li>
        <li class="flex gap-3">
          <TriangleAlert size={20} class="shrink-0 text-warning" />
          <div>
            <p class="font-medium">Fermez l'ancienne application sur tous les postes</p>
            <p class="text-muted">Ainsi, aucune saisie ne sera perdue pendant l'import. L'ancien fichier n'est jamais modifié : il reste utilisable en cas de besoin.</p>
          </div>
        </li>
        <li class="flex gap-3">
          <Database size={20} class="shrink-0 text-primary" />
          <div>
            <p class="font-medium">Toutes les versions sont reconnues</p>
            <p class="text-muted">Version 3 (3.x), version 2 et version 1 (Python). La version est détectée automatiquement et un aperçu est affiché avant tout import.</p>
          </div>
        </li>
      </ul>
      <div class="mt-7 flex items-center justify-between border-t border-line pt-5">
        {#if info?.isEmpty}
          <div class="flex gap-2">
            <button class="btn btn-ghost" onclick={startEmpty}>Commencer avec une base vide</button>
            <button class="btn btn-ghost" onclick={loadDemo}>Essayer avec des données de démonstration</button>
          </div>
        {:else}
          <span></span>
        {/if}
        <button class="btn btn-primary" onclick={() => (step = 2)}>Continuer<ArrowRight size={15} /></button>
      </div>
    {:else if step === 2}
      <h2 class="text-[16px] font-semibold">Choisissez l'ancienne base</h2>
      <p class="mt-1 text-[13.5px] text-muted">Sélectionnez le fichier <span class="font-mono">clefs.db</span> (ou une sauvegarde <span class="font-mono">.db</span>).</p>
      <button class="mt-6 flex w-full cursor-pointer flex-col items-center gap-3 rounded-card border-2 border-dashed border-line px-6 py-10 hover:border-primary hover:bg-primary-soft/40" onclick={choose}>
        <FileSearch size={28} class="text-primary" />
        <span class="text-[14px] font-medium">Parcourir…</span>
      </button>
      {#if inspectError}<div class="alert alert-danger mt-4">{inspectError}</div>{/if}
      <div class="mt-7 flex justify-between border-t border-line pt-5">
        <button class="btn btn-ghost" onclick={() => (step = 1)}><ArrowLeft size={15} />Retour</button>
      </div>
    {:else if step === 3 && file}
      <h2 class="text-[16px] font-semibold">Vérifiez avant d'importer</h2>
      <p class="mt-1 truncate font-mono text-[12.5px] text-muted" data-selectable title={file.path}>{file.path}</p>

      {#if file.kind === 'invalid'}
        <div class="alert alert-danger mt-5">
          <TriangleAlert size={18} class="mt-0.5 shrink-0 text-danger" />
          <p>Ce fichier n'est pas une base du Gestionnaire de Clés. Vérifiez que vous avez choisi le bon fichier <span class="font-mono">clefs.db</span>.</p>
        </div>
      {:else}
        <div class="mt-5 flex items-center gap-2">
          <span class="badge badge-success"><CircleCheck size={12} />Base reconnue</span>
          <span class="text-[13.5px] font-medium">{file.label}</span>
        </div>
        <div class="mt-4 grid grid-cols-5 gap-2 text-center">
          {#each [[file.buildings, 'bâtiments'], [file.accesses, 'accès'], [file.keys, 'clés'], [file.borrowers, 'détenteurs'], [file.loans, 'prêts']] as [n, label]}
            <div class="rounded-control bg-surface-2 px-2 py-3">
              <p class="text-[20px] font-semibold tabular-nums">{n}</p>
              <p class="text-[12px] text-muted">{label}</p>
            </div>
          {/each}
        </div>
        <p class="mt-2 text-[12.5px] text-muted">dont {plural(file.activeLoans, 'prêt en cours', 'prêts en cours')}.</p>

        {#if file.kind === 'legacy'}
          <div class="alert alert-info mt-5">
            Les informations apparues en version 3 (étage, catégorie, statut des détenteurs…) seront vides : vous pourrez les compléter ensuite.
            {#if info && !info.isEmpty}<br /><strong>La base actuelle doit être vide</strong> : réinitialisez-la d'abord depuis les Paramètres.{/if}
          </div>
        {:else if info && !info.isEmpty}
          <div class="alert alert-warning mt-5">La base actuelle contient déjà des données : elles seront remplacées (une copie de sécurité est faite avant).</div>
        {/if}
      {/if}

      {#if importError}<div class="alert alert-danger mt-4">{importError}</div>{/if}
      <div class="mt-7 flex justify-between border-t border-line pt-5">
        <button class="btn btn-ghost" onclick={() => { step = 2; file = null; importError = '' }}><ArrowLeft size={15} />Choisir un autre fichier</button>
        {#if file.kind !== 'invalid'}
          <button class="btn btn-primary" disabled={importing} onclick={doImport}><Upload size={15} />{importing ? 'Import en cours…' : 'Importer'}</button>
        {/if}
      </div>
    {:else if step === 4 && result}
      <div class="flex flex-col items-center py-4 text-center">
        <div class="flex size-14 items-center justify-center rounded-full bg-success-soft text-success"><CircleCheck size={28} /></div>
        <h2 class="mt-4 text-[18px] font-semibold">Import terminé</h2>
        <p class="mt-1 text-[13.5px] text-muted">
          {plural(result.keys, 'clé')}, {plural(result.accesses, 'accès', 'accès')}, {plural(result.borrowers, 'détenteur')} et {plural(result.loans, 'prêt')} ont été repris.
        </p>
        <p class="mt-4 max-w-md text-[13px] text-muted">Pensez à faire une première sauvegarde depuis la page Sauvegardes, puis vérifiez le tableau de bord.</p>
        <div class="mt-6 flex gap-2">
          <button class="btn btn-secondary" onclick={() => navigate('backups')}>Faire une sauvegarde</button>
          <button class="btn btn-primary" onclick={() => navigate('dashboard')}>Aller au tableau de bord<ArrowRight size={15} /></button>
        </div>
      </div>
    {/if}
  </div>
</div>
