<script lang="ts">
  // Paramètres : établissement et documents, apparence, données (dossiers,
  // import, démonstration, réinitialisation).
  import { FolderOpen, Monitor, Moon, Sun, Upload } from '@lucide/svelte'
  import { Api, run, type AppInfo } from '../lib/api'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import SettingsForm from '../lib/components/SettingsForm.svelte'
  import { confirm, data, navigate, setTheme, theme, type ThemeChoice } from '../lib/state.svelte'

  let info = $state<AppInfo | null>(null)
  let tab = $state<'general' | 'appearance' | 'data'>('general')

  $effect(() => {
    void data.version
    void run(Api.GetAppInfo()).then((i) => (info = i ?? null))
  })

  const themes: [ThemeChoice, string, typeof Sun][] = [
    ['light', 'Clair', Sun],
    ['dark', 'Sombre', Moon],
    ['system', 'Comme Windows', Monitor],
  ]

  async function demo() {
    await run(Api.LoadDemoData(), 'Données de démonstration chargées.')
  }

  async function reset() {
    const ok = await confirm({
      title: 'Effacer toutes les données ?',
      message: 'Clés, accès, bâtiments, détenteurs et historique seront supprimés.\nUne sauvegarde de sécurité est faite automatiquement avant.',
      confirmLabel: 'Tout effacer',
      danger: true,
      typeToConfirm: 'EFFACER',
    })
    if (ok) await run(Api.ResetDatabase(), 'Base réinitialisée.')
  }
</script>

<PageHeader title="Paramètres" subtitle="Les réglages de l'établissement et des documents sont communs à tous les postes ; l'apparence est propre à chaque poste." />

<div class="segmented mb-5">
  <button aria-pressed={tab === 'general'} onclick={() => (tab = 'general')}>Établissement et documents</button>
  <button aria-pressed={tab === 'appearance'} onclick={() => (tab = 'appearance')}>Apparence</button>
  <button aria-pressed={tab === 'data'} onclick={() => (tab = 'data')}>Données</button>
</div>

{#if tab === 'general'}
  <div class="max-w-4xl"><SettingsForm /></div>
{:else}
<div class="max-w-3xl space-y-5">
  {#if tab === 'appearance'}
  <section class="card p-6">
    <h2 class="text-[15px] font-semibold">Apparence</h2>
    <div class="mt-4 grid grid-cols-3 gap-3">
      {#each themes as [value, label, Icon]}
        <button class="flex cursor-pointer items-center gap-3 rounded-control border px-4 py-3 text-left {theme.choice === value ? 'border-primary bg-primary-soft' : 'border-line hover:bg-surface-2'}" onclick={() => setTheme(value)}>
          <Icon size={18} class={theme.choice === value ? 'text-primary' : 'text-muted'} />
          <span class="text-[13.5px] font-medium">{label}</span>
        </button>
      {/each}
    </div>
  </section>
  {/if}

  {#if info && tab === 'data'}
    <section class="card p-6">
      <h2 class="text-[15px] font-semibold">Emplacement des données</h2>
      <p class="mt-1 text-[13px] text-muted">Tout est rangé dans le dossier de l'application : copiez ce dossier pour déplacer l'installation.</p>
      <dl class="mt-4 space-y-2 text-[13px]" data-selectable>
        <div class="flex gap-3"><dt class="w-40 text-muted">Base de données</dt><dd class="font-mono text-[12.5px]">{info.dbPath}</dd></div>
        <div class="flex gap-3"><dt class="w-40 text-muted">Documents (PDF, CSV)</dt><dd class="font-mono text-[12.5px]">{info.documentsDir}</dd></div>
        <div class="flex gap-3"><dt class="w-40 text-muted">Sauvegardes</dt><dd class="font-mono text-[12.5px]">{info.backupsDir}</dd></div>
        <div class="flex gap-3"><dt class="w-40 text-muted">Agent</dt><dd>{info.agent}</dd></div>
      </dl>
      <div class="mt-4 flex gap-2">
        <button class="btn btn-secondary" onclick={() => run(Api.OpenPath(info!.dataDir))}><FolderOpen size={15} />Ouvrir le dossier</button>
        <button class="btn btn-secondary" onclick={() => run(Api.OpenDocumentsFolder())}><FolderOpen size={15} />Ouvrir les documents</button>
      </div>
    </section>

    <section class="card p-6">
      <h2 class="text-[15px] font-semibold">Données</h2>
      <div class="mt-4 divide-y divide-line">
        <div class="flex items-center gap-4 py-3">
          <div class="flex-1">
            <p class="font-medium">Importer une base existante</p>
            <p class="text-[13px] text-muted">Reprendre les données d'une version précédente (V3, V2 ou V1).</p>
          </div>
          <button class="btn btn-secondary" onclick={() => navigate('import')}><Upload size={15} />Ouvrir l'assistant</button>
        </div>
        <div class="flex items-center gap-4 py-3">
          <div class="flex-1">
            <p class="font-medium">Données de démonstration</p>
            <p class="text-[13px] text-muted">Pour découvrir l'application. Uniquement dans une base vide.</p>
          </div>
          <button class="btn btn-secondary" disabled={!info.isEmpty} onclick={demo}>Charger la démo</button>
        </div>
        <div class="flex items-center gap-4 py-3">
          <div class="flex-1">
            <p class="font-medium text-danger">Réinitialiser la base</p>
            <p class="text-[13px] text-muted">Efface toutes les données, après une sauvegarde de sécurité automatique.</p>
          </div>
          <button class="btn btn-danger-ghost border border-danger/30" onclick={reset}>Tout effacer…</button>
        </div>
      </div>
    </section>
  {/if}
</div>
{/if}
