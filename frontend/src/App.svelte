<script lang="ts">
  // Structure de l'application : menu latéral, barre supérieure, page
  // courante, fiches détaillées et éléments globaux (notifications,
  // confirmations, recherche).
  import {
    ArrowLeftRight,
    Building2,
    ClipboardCheck,
    CircleHelp,
    CopyX,
    DatabaseBackup,
    DoorOpen,
    HandCoins,
    History,
    Info,
    KeyRound,
    LayoutDashboard,
    Map,
    Moon,
    Power,
    Search,
    Settings,
    Sun,
    Upload,
    Users,
    UserSearch,
  } from '@lucide/svelte'
  import type { Component } from 'svelte'
  import { Api, loadAppSettings } from './lib/api'
  import AccessDetail from './lib/components/AccessDetail.svelte'
  import BorrowerDetail from './lib/components/BorrowerDetail.svelte'
  import CommandPalette from './lib/components/CommandPalette.svelte'
  import ConfirmDialog from './lib/components/ConfirmDialog.svelte'
  import KeyDetail from './lib/components/KeyDetail.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import { appSettings, confirm, detail, inventoryDraft, navigate, route, setTheme, theme, type Page } from './lib/state.svelte'
  import About from './pages/About.svelte'
  import Accesses from './pages/Accesses.svelte'
  import ActiveLoans from './pages/ActiveLoans.svelte'
  import Backups from './pages/Backups.svelte'
  import Borrowers from './pages/Borrowers.svelte'
  import Buildings from './pages/Buildings.svelte'
  import Dashboard from './pages/Dashboard.svelte'
  import Help from './pages/Help.svelte'
  import HistoryPage from './pages/History.svelte'
  import ImportWizard from './pages/ImportWizard.svelte'
  import KeyPlan from './pages/KeyPlan.svelte'
  import Inventory from './pages/Inventory.svelte'
  import Keys from './pages/Keys.svelte'
  import NewLoan from './pages/NewLoan.svelte'
  import Redundancies from './pages/Redundancies.svelte'
  import SettingsPage from './pages/Settings.svelte'
  import WhoHasWhat from './pages/WhoHasWhat.svelte'

  type NavItem = { page: Page; label: string; icon: Component<{ size?: number; class?: string }> }
  const nav: { title: string; items: NavItem[] }[] = [
    {
      title: '',
      items: [
        { page: 'dashboard', label: 'Tableau de bord', icon: LayoutDashboard },
        { page: 'loans', label: 'Prêts en cours', icon: ArrowLeftRight },
        { page: 'history', label: 'Historique', icon: History },
      ],
    },
    {
      title: 'Référentiel',
      items: [
        { page: 'keys', label: 'Clés', icon: KeyRound },
        { page: 'inventory', label: 'Inventaire', icon: ClipboardCheck },
        { page: 'accesses', label: 'Accès', icon: DoorOpen },
        { page: 'borrowers', label: 'Détenteurs', icon: Users },
        { page: 'buildings', label: 'Bâtiments', icon: Building2 },
      ],
    },
    {
      title: 'Consultation',
      items: [
        { page: 'who-has-what', label: 'Qui a quoi ?', icon: UserSearch },
        { page: 'key-plan', label: 'Plan de clés', icon: Map },
        { page: 'redundancies', label: 'Doublons', icon: CopyX },
      ],
    },
    {
      title: 'Outils',
      items: [
        { page: 'backups', label: 'Sauvegardes', icon: DatabaseBackup },
        { page: 'import', label: 'Importer une base', icon: Upload },
        { page: 'settings', label: 'Paramètres', icon: Settings },
        { page: 'help', label: 'Aide', icon: CircleHelp },
        { page: 'about', label: 'À propos', icon: Info },
      ],
    },
  ]
  const allPages = [{ page: 'new-loan' as Page, label: 'Nouveau prêt' }, ...nav.flatMap((g) => g.items.map((i) => ({ page: i.page, label: i.label })))]

  const pages: Record<Page, Component> = {
    dashboard: Dashboard,
    'new-loan': NewLoan,
    loans: ActiveLoans,
    history: HistoryPage,
    keys: Keys,
    inventory: Inventory,
    accesses: Accesses,
    borrowers: Borrowers,
    buildings: Buildings,
    'who-has-what': WhoHasWhat,
    'key-plan': KeyPlan,
    redundancies: Redundancies,
    backups: Backups,
    import: ImportWizard,
    settings: SettingsPage,
    help: Help,
    about: About,
  }
  const CurrentPage = $derived(pages[route.page])

  let paletteOpen = $state(false)
  let version = $state('')
  void loadAppSettings()
  void Api.GetAppInfo().then((info) => {
    version = info.version
    // Nouvelle installation : l'assistant d'import est proposé d'office.
    if (info.isEmpty) navigate('import')
  })

  const isDark = $derived(theme.choice === 'dark' || (theme.choice === 'system' && document.documentElement.classList.contains('dark')))

  async function quit() {
    const pending = Object.keys(inventoryDraft.counts).length
    if (pending) {
      const ok = await confirm({
        title: 'Quitter sans valider l’inventaire ?',
        message: `La saisie d'inventaire en cours (${pending} clé(s) comptée(s)) sera perdue.`,
        confirmLabel: 'Quitter',
        danger: true,
      })
      if (!ok) return
    }
    void Api.Quit()
  }

  function onkey(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault()
      paletteOpen = true
    }
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'n') {
      e.preventDefault()
      navigate('new-loan')
    }
  }
</script>

<svelte:window onkeydown={onkey} />

<div class="flex h-screen overflow-hidden">
  <aside class="flex w-60 shrink-0 flex-col bg-sidebar text-sidebar-text">
    <div class="flex items-center gap-2.5 px-5 pt-4 pb-3">
      {#if appSettings.logoDataUrl}
        <img src={appSettings.logoDataUrl} alt="" class="h-9 max-w-14 shrink-0 rounded-[8px] bg-white object-contain p-0.5" />
      {:else}
        <div class="flex size-9 shrink-0 items-center justify-center rounded-[10px] bg-sidebar-accent text-white shadow-sm"><KeyRound size={18} /></div>
      {/if}
      <div class="min-w-0 leading-tight">
        <p class="truncate text-[14px] font-semibold text-white" title={appSettings.appName}>{appSettings.appName}</p>
        <p class="truncate text-[11.5px] text-sidebar-text/70" title={appSettings.establishment}>{appSettings.establishment}</p>
      </div>
    </div>

    <div class="px-3 pb-3">
      <button class="btn h-10 w-full bg-sidebar-accent text-white shadow-sm hover:bg-sidebar-accent-hover" onclick={() => navigate('new-loan')}>
        <HandCoins size={17} />Nouveau prêt
      </button>
    </div>

    <nav class="sidebar-nav flex-1 space-y-3 overflow-y-auto px-3 pb-3">
      {#each nav as group}
        <div>
          {#if group.title}<p class="px-3 pb-1 text-[11px] font-semibold tracking-wider text-sidebar-text/50 uppercase">{group.title}</p>{/if}
          {#each group.items as item}
            {@const Icon = item.icon}
            <button
              class="flex h-8 w-full cursor-pointer items-center gap-3 rounded-control px-3 text-[13.5px] transition-colors {route.page === item.page
                ? 'bg-sidebar-active font-medium text-white'
                : 'hover:bg-sidebar-active/60 hover:text-white'}"
              onclick={() => navigate(item.page)}
            >
              <Icon size={17} class={route.page === item.page ? 'text-sidebar-icon' : ''} />
              {item.label}
            </button>
          {/each}
        </div>
      {/each}
    </nav>
    <div class="flex items-center justify-between border-t border-white/5 px-3 py-2.5">
      <span class="px-3 text-[11px] text-sidebar-text/50">Version {version}</span>
      <button class="flex h-8 cursor-pointer items-center gap-2 rounded-control px-3 text-[13px] hover:bg-sidebar-active/60 hover:text-white" onclick={quit} title="Fermer l'application">
        <Power size={15} />Quitter
      </button>
    </div>
  </aside>

  <div class="flex min-w-0 flex-1 flex-col">
    <header class="flex h-14 shrink-0 items-center gap-3 border-b border-line bg-surface px-6">
      <button class="flex h-9 w-full max-w-md cursor-pointer items-center gap-2.5 rounded-control border border-line bg-surface-2 px-3 text-[13px] text-muted hover:border-primary/40" onclick={() => (paletteOpen = true)}>
        <Search size={15} />
        <span class="flex-1 text-left">Rechercher une clé, une porte, une personne…</span>
        <span class="kbd">Ctrl K</span>
      </button>
      <div class="flex-1"></div>
      <button class="btn btn-ghost btn-icon" onclick={() => setTheme(isDark ? 'light' : 'dark')} title={isDark ? 'Passer en mode clair' : 'Passer en mode sombre'} aria-label="Changer de thème">
        {#if isDark}<Sun size={17} />{:else}<Moon size={17} />{/if}
      </button>
      <button class="btn btn-ghost btn-icon" onclick={() => navigate('help')} title="Aide" aria-label="Aide"><CircleHelp size={17} /></button>
    </header>

    <main class="min-h-0 flex-1 overflow-y-auto">
      <div class="mx-auto max-w-[1280px] px-8 py-7">
        {#key route.page + JSON.stringify(route.params)}
          <CurrentPage />
        {/key}
      </div>
    </main>
  </div>
</div>

{#if detail.type === 'key'}
  {#key detail.id}<KeyDetail id={detail.id} />{/key}
{:else if detail.type === 'access'}
  {#key detail.id}<AccessDetail id={detail.id} />{/key}
{:else if detail.type === 'borrower'}
  {#key detail.id}<BorrowerDetail id={detail.id} />{/key}
{/if}

<CommandPalette bind:open={paletteOpen} pages={allPages} />
<ConfirmDialog />
<Toasts />
