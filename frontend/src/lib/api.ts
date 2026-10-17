// Accès aux services Go. Les fonctions sont générées par Wails dans
// ./wailsjs (ne pas modifier ces fichiers : `make bindings` les régénère).
import * as Api from './wailsjs/go/app/App'
import { app } from './wailsjs/go/models'
import { appSettings, notify, refresh } from './state.svelte'

export { Api }
export type Key = app.KeyDTO
export type Access = app.AccessDTO
export type Building = app.BuildingDTO
export type Borrower = app.BorrowerDTO
export type Loan = app.LoanDTO
export type Anomaly = app.AnomalyDTO
export type Redundancy = app.RedundancyDTO
export type Dashboard = app.Dashboard
export type KeyDetail = app.KeyDetail
export type AccessDetail = app.AccessDetail
export type BorrowerDetail = app.BorrowerDetail
export type Suggestion = app.Suggestion
export type SuggestedKey = app.SuggestedKey
export type LoanResult = app.LoanResult
export type WhoHasWhat = app.WhoHasWhat
export type PlanBuilding = app.PlanBuilding
export type Backup = app.BackupDTO
export type DatabaseFileInfo = app.DatabaseFileInfo
export type ImportResult = app.ImportResult
export type AppInfo = app.AppInfo
export type HistoryFilter = app.HistoryFilter
export type Settings = app.Settings

/** Recharge les paramètres utiles à l'affichage (appSettings). */
export async function loadAppSettings() {
  try {
    const s = await Api.GetSettings()
    Object.assign(appSettings, {
      appName: s.appName,
      establishment: s.establishment,
      address: s.address,
      logoDataUrl: s.logoDataUrl,
      defaultLoanType: s.defaultLoanType,
      defaultReturnDays: s.defaultReturnDays,
    })
    document.title = s.appName
  } catch (e) {
    notify('error', errorMessage(e))
  }
}

/** Met une majuscule au message d'erreur renvoyé par Go. */
export function errorMessage(e: unknown): string {
  const s = typeof e === 'string' ? e : e instanceof Error ? e.message : String(e)
  return s.charAt(0).toUpperCase() + s.slice(1)
}

/**
 * Exécute un appel au moteur : en cas d'erreur, l'affiche en notification et
 * retourne undefined. Avec `success`, notifie la réussite et rafraîchit les
 * données affichées (l'appel a modifié la base).
 */
export async function run<T>(call: Promise<T>, success?: string): Promise<T | undefined> {
  try {
    const result = await call
    if (success) {
      notify('success', success)
      refresh()
    }
    return result
  } catch (e) {
    notify('error', errorMessage(e))
    return undefined
  }
}

/** Notifie un document créé, avec un bouton pour l'ouvrir. */
export function notifyDocument(message: string, path: string) {
  notify('success', message, { label: 'Ouvrir', run: () => void run(Api.OpenPath(path)) })
}
