// Mise en forme française des dates, quantités et libellés.

const dateFmt = new Intl.DateTimeFormat('fr-FR', { day: '2-digit', month: '2-digit', year: 'numeric' })
const dateTimeFmt = new Intl.DateTimeFormat('fr-FR', { dateStyle: 'short', timeStyle: 'short' })
const longDateFmt = new Intl.DateTimeFormat('fr-FR', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })

/** 17/10/2026, ou « — » si la date est vide. */
export function fmtDate(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : dateFmt.format(d)
}

/** 17/10/2026 14:05 */
export function fmtDateTime(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : dateTimeFmt.format(d)
}

/** samedi 17 octobre 2026 */
export function fmtLongDate(d: Date = new Date()): string {
  return longDateFmt.format(d)
}

/** « aujourd'hui », « 1 jour », « 12 jours » */
export function fmtDays(days: number): string {
  if (days <= 0) return "aujourd'hui"
  return days === 1 ? '1 jour' : `${days} jours`
}

/** Accord simple : plural(3, 'clé') → « 3 clés ». */
export function plural(n: number, word: string, pluralWord = word + 's'): string {
  return `${n} ${n > 1 ? pluralWord : word}`
}

/** Date du jour au format AAAA-MM-JJ (champ date). */
export function todayISO(offsetDays = 0): string {
  const d = new Date()
  d.setDate(d.getDate() + offsetDays)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

export const loanStatusLabel: Record<string, string> = {
  active: 'En cours',
  overdue: 'En retard',
  returned: 'Retourné',
}

export const loanStatusBadge: Record<string, string> = {
  active: 'badge-primary',
  overdue: 'badge-warning',
  returned: 'badge-neutral',
}

export const borrowerStatusLabel: Record<string, string> = {
  permanent: 'Permanent',
  contractuel: 'Contractuel',
  intervenant: 'Intervenant',
  entreprise: 'Entreprise',
}

export const keyCategoryLabel: Record<string, string> = {
  simple: 'Clé simple',
  trousseau: 'Trousseau',
  badge: 'Badge',
  passe: 'Passe',
}

/** Recherche insensible à la casse et aux accents. */
export function normalize(s: string): string {
  return s
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
}

export function matches(query: string, ...fields: (string | undefined | null)[]): boolean {
  const q = normalize(query.trim())
  if (!q) return true
  return fields.some((f) => f && normalize(f).includes(q))
}
