<script lang="ts">
  // Sauvegardes : création, restauration, suppression.
  import { DatabaseBackup, FolderOpen, Plus, RotateCcw, Trash2 } from '@lucide/svelte'
  import { Api, run, type AppInfo, type Backup } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import { fmtDateTime } from '../lib/format'
  import { confirm, data } from '../lib/state.svelte'

  let backups = $state<Backup[]>([])
  let info = $state<AppInfo | null>(null)
  let loaded = $state(false)
  let busy = $state(false)

  $effect(() => {
    void data.version
    void Promise.all([run(Api.ListBackups()), run(Api.GetAppInfo())]).then(([b, i]) => {
      backups = b ?? []
      info = i ?? null
      loaded = true
    })
  })

  async function create() {
    busy = true
    await run(Api.CreateBackup(), 'Sauvegarde créée.')
    busy = false
  }

  async function restore(b: Backup) {
    const ok = await confirm({
      title: 'Restaurer cette sauvegarde ?',
      message: `Toutes les données actuelles seront remplacées par celles du ${fmtDateTime(b.date)}.\nUne copie de sécurité de l'état actuel est faite automatiquement avant.`,
      confirmLabel: 'Restaurer',
      danger: true,
    })
    if (!ok) return
    busy = true
    await run(Api.RestoreBackup(b.path), 'Sauvegarde restaurée.')
    busy = false
  }

  async function remove(b: Backup) {
    const ok = await confirm({ title: 'Supprimer cette sauvegarde ?', message: b.name, confirmLabel: 'Supprimer', danger: true })
    if (ok) await run(Api.DeleteBackup(b.path), 'Sauvegarde supprimée.')
  }
</script>

<PageHeader title="Sauvegardes" subtitle="Une sauvegarde est une copie complète de la base, à faire régulièrement (et avant toute mise à jour).">
  {#snippet actions()}
    {#if info}<button class="btn btn-secondary" onclick={() => run(Api.OpenPath(info!.backupsDir))}><FolderOpen size={15} />Ouvrir le dossier</button>{/if}
    <button class="btn btn-primary" disabled={busy} onclick={create}><Plus size={16} />Sauvegarder maintenant</button>
  {/snippet}
</PageHeader>

<div class="card overflow-hidden">
  {#if backups.length}
    <table class="table">
      <thead><tr><th>Date</th><th>Fichier</th><th>Taille</th><th></th></tr></thead>
      <tbody>
        {#each backups as b, i (b.path)}
          <tr>
            <td class="font-medium tabular-nums">{fmtDateTime(b.date)} {#if i === 0}<span class="badge badge-success ml-2">la plus récente</span>{/if}</td>
            <td class="font-mono text-[12px] text-muted" data-selectable>{b.name}</td>
            <td class="text-muted">{b.size}</td>
            <td class="text-right whitespace-nowrap">
              <button class="btn btn-secondary btn-sm" disabled={busy} onclick={() => restore(b)}><RotateCcw size={14} />Restaurer</button>
              <button class="btn btn-danger-ghost btn-sm btn-icon" disabled={busy} onclick={() => remove(b)} aria-label="Supprimer"><Trash2 size={14} /></button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if loaded}
    <EmptyState icon={DatabaseBackup} title="Aucune sauvegarde" text="Créez une première sauvegarde : elle sera rangée dans le dossier backups/ à côté de l'application.">
      <button class="btn btn-primary" onclick={create}><Plus size={16} />Sauvegarder maintenant</button>
    </EmptyState>
  {/if}
</div>
{#if info}<p class="mt-3 text-[12.5px] text-muted" data-selectable>Dossier : {info.backupsDir}</p>{/if}
