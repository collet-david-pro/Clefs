<script lang="ts">
  // Bâtiments : création, renommage, suppression.
  import { Building2, Check, Pencil, Plus, Trash2, X } from '@lucide/svelte'
  import { Api, run, type Building } from '../lib/api'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import { plural } from '../lib/format'
  import { confirm, data, navigate } from '../lib/state.svelte'

  let buildings = $state<Building[]>([])
  let loaded = $state(false)
  let newName = $state('')
  let editing = $state(0)
  let editName = $state('')

  $effect(() => {
    void data.version
    void run(Api.ListBuildings()).then((b) => {
      buildings = b ?? []
      loaded = true
    })
  })

  async function create() {
    if (!newName.trim()) return
    if ((await run(Api.SaveBuilding(0, newName), 'Bâtiment créé.')) !== undefined) newName = ''
  }

  async function rename(b: Building) {
    if ((await run(Api.SaveBuilding(b.id, editName), 'Bâtiment renommé.')) !== undefined) editing = 0
  }

  async function remove(b: Building) {
    const ok = await confirm({
      title: `Supprimer « ${b.name} » ?`,
      message: b.accessCount
        ? `Ses ${plural(b.accessCount, 'accès', 'accès')} seront aussi supprimés. Les clés sont conservées.`
        : 'Ce bâtiment ne contient aucun accès.',
      confirmLabel: 'Supprimer',
      danger: true,
    })
    if (ok) await run(Api.DeleteBuilding(b.id), 'Bâtiment supprimé.')
  }
</script>

<PageHeader title="Bâtiments" subtitle="Les accès sont rangés par bâtiment." />

<div class="card mb-4 flex gap-2 p-4">
  <input class="input" placeholder="Nom du nouveau bâtiment (ex. Bâtiment B, Gymnase…)" bind:value={newName} onkeydown={(e) => e.key === 'Enter' && create()} />
  <button class="btn btn-primary" disabled={!newName.trim()} onclick={create}><Plus size={16} />Ajouter</button>
</div>

<div class="card overflow-hidden">
  {#if buildings.length}
    <ul class="divide-y divide-line">
      {#each buildings as b (b.id)}
        <li class="flex items-center gap-4 px-5 py-3">
          <Building2 size={18} class="text-muted" />
          {#if editing === b.id}
            <!-- svelte-ignore a11y_autofocus -->
            <input class="input max-w-sm" bind:value={editName} autofocus onkeydown={(e) => { if (e.key === 'Enter') rename(b); if (e.key === 'Escape') editing = 0 }} />
            <button class="btn btn-primary btn-sm btn-icon" onclick={() => rename(b)} aria-label="Valider"><Check size={15} /></button>
            <button class="btn btn-ghost btn-sm btn-icon" onclick={() => (editing = 0)} aria-label="Annuler"><X size={15} /></button>
            <div class="flex-1"></div>
          {:else}
            <span class="flex-1 font-medium">{b.name}</span>
            <button class="btn btn-ghost btn-sm" onclick={() => navigate('key-plan')}>{plural(b.accessCount, 'accès', 'accès')}</button>
            <button class="btn btn-ghost btn-sm btn-icon" onclick={() => { editing = b.id; editName = b.name }} aria-label="Renommer"><Pencil size={14} /></button>
            <button class="btn btn-danger-ghost btn-sm btn-icon" onclick={() => remove(b)} aria-label="Supprimer"><Trash2 size={14} /></button>
          {/if}
        </li>
      {/each}
    </ul>
  {:else if loaded}
    <EmptyState icon={Building2} title="Aucun bâtiment" text="Commencez par créer vos bâtiments, puis les accès de chacun et enfin les clés." />
  {/if}
</div>
