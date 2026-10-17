// État global de l'interface (runes Svelte 5) : navigation, rafraîchissement
// des données, notifications, confirmations, fiches détaillées et thème.

export type Page =
  | 'dashboard'
  | 'new-loan'
  | 'loans'
  | 'history'
  | 'keys'
  | 'inventory'
  | 'accesses'
  | 'borrowers'
  | 'buildings'
  | 'who-has-what'
  | 'key-plan'
  | 'redundancies'
  | 'backups'
  | 'import'
  | 'settings'
  | 'help'
  | 'about'

export type RouteParams = { borrowerId?: number; keyId?: number }

export const route = $state<{ page: Page; params: RouteParams }>({ page: 'dashboard', params: {} })

/** Change de page (les paramètres servent à pré-remplir, ex. nouveau prêt). */
export function navigate(page: Page, params: RouteParams = {}) {
  route.page = page
  route.params = params
  detail.type = null
}

/**
 * Compteur de version des données : incrémenté après chaque modification,
 * il déclenche le rechargement des pages et fiches ouvertes ($effect).
 */
export const data = $state({ version: 0 })
export function refresh() {
  data.version++
}

/**
 * Paramètres de l'application utiles à l'affichage (nom de l'établissement,
 * valeurs par défaut d'un prêt…). Chargés au démarrage, mis à jour à
 * l'enregistrement des Paramètres.
 */
export const appSettings = $state({
  appName: 'Gestionnaire de Clés',
  establishment: '',
  address: '',
  logoDataUrl: '',
  defaultLoanType: 'ponctuel',
  defaultReturnDays: 0,
})

/**
 * Saisie d'inventaire en cours (ID de clé → exemplaires comptés en stock).
 * Conservée hors de la page pour ne pas être perdue en changeant d'onglet.
 */
export const inventoryDraft = $state<{ counts: Record<number, number> }>({ counts: {} })

// --- Notifications non bloquantes ---

export type Toast = {
  id: number
  kind: 'success' | 'error' | 'info'
  message: string
  action?: { label: string; run: () => void }
}
export const toasts = $state<Toast[]>([])
let toastSeq = 0

export function notify(kind: Toast['kind'], message: string, action?: Toast['action']) {
  const id = ++toastSeq
  toasts.push({ id, kind, message, action })
  setTimeout(() => dismiss(id), kind === 'error' ? 8000 : action ? 7000 : 4000)
}
export function dismiss(id: number) {
  const i = toasts.findIndex((t) => t.id === id)
  if (i >= 0) toasts.splice(i, 1)
}

// --- Confirmation des actions sensibles ---

export type ConfirmOptions = {
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
  /** Mot à saisir pour confirmer (actions irréversibles). */
  typeToConfirm?: string
}
export const confirmState = $state<{ open: boolean; options: ConfirmOptions | null }>({ open: false, options: null })
let confirmResolve: ((ok: boolean) => void) | null = null

export function confirm(options: ConfirmOptions): Promise<boolean> {
  confirmState.options = options
  confirmState.open = true
  return new Promise((resolve) => (confirmResolve = resolve))
}
export function closeConfirm(ok: boolean) {
  confirmState.open = false
  confirmResolve?.(ok)
  confirmResolve = null
}

// --- Fiches détaillées (panneau latéral) ---

export const detail = $state<{ type: 'key' | 'access' | 'borrower' | null; id: number }>({ type: null, id: 0 })
export function openDetail(type: 'key' | 'access' | 'borrower', id: number) {
  detail.type = type
  detail.id = id
}
export function closeDetail() {
  detail.type = null
}

// --- Thème clair / sombre ---

export type ThemeChoice = 'light' | 'dark' | 'system'
export const theme = $state<{ choice: ThemeChoice }>({ choice: loadTheme() })

function loadTheme(): ThemeChoice {
  try {
    const v = localStorage.getItem('theme')
    if (v === 'light' || v === 'dark' || v === 'system') return v
  } catch {
    /* stockage indisponible : thème système */
  }
  return 'system'
}

export function setTheme(choice: ThemeChoice) {
  theme.choice = choice
  try {
    localStorage.setItem('theme', choice)
  } catch {
    /* préférence non mémorisée */
  }
  applyTheme()
}

export function applyTheme() {
  const dark =
    theme.choice === 'dark' || (theme.choice === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', dark)
}
