<script lang="ts">
  // Fiche d'un accès : clés qui l'ouvrent et personnes qui peuvent l'ouvrir.
  import { Pencil, Trash2 } from '@lucide/svelte'
  import { Api, run, type AccessDetail } from '../api'
  import { closeDetail, confirm, data, openDetail } from '../state.svelte'
  import Drawer from './Drawer.svelte'
  import AccessForm from './AccessForm.svelte'
  import StockBar from './StockBar.svelte'

  let { id }: { id: number } = $props()

  let d = $state<AccessDetail | null>(null)
  let editOpen = $state(false)

  $effect(() => {
    void data.version
    void load(id)
  })

  async function load(accessId: number) {
    d = (await run(Api.GetAccessDetail(accessId))) ?? null
  }

  async function remove() {
    if (!d) return
    const ok = await confirm({
      title: `Supprimer l'accès « ${d.access.name} » ?`,
      message: "Les clés sont conservées : elles n'ouvriront simplement plus cet accès.",
      confirmLabel: 'Supprimer',
      danger: true,
    })
    if (ok && (await run(Api.DeleteAccess(d.access.id).then(() => true), 'Accès supprimé.'))) closeDetail()
  }
</script>

<Drawer onclose={closeDetail}>
  {#snippet header()}
    {#if d}
      <p class="text-[12.5px] font-medium text-muted">{d.access.buildingName}{d.access.floor ? ` · ${d.access.floor}` : ''}</p>
      <h2 class="mt-1 truncate text-[18px] font-semibold">{d.access.name}</h2>
      <div class="mt-2 flex flex-wrap gap-1.5">
        {#if d.access.category}<span class="badge badge-neutral">{d.access.category}</span>{/if}
        {#if d.access.type}<span class="badge badge-neutral">{d.access.type}</span>{/if}
      </div>
    {:else}
      <h2 class="text-[18px] font-semibold">Chargement…</h2>
    {/if}
  {/snippet}

  {#if d}
    <section>
      <h3 class="section-title mb-2">Clés qui ouvrent cet accès ({d.keys.length})</h3>
      {#if d.keys.length}
        <ul class="divide-y divide-line rounded-control border border-line">
          {#each d.keys as k (k.id)}
            <li>
              <button class="flex w-full cursor-pointer items-center gap-3 px-3 py-2.5 text-left hover:bg-surface-2" onclick={() => openDetail('key', k.id)}>
                <span class="chip chip-key">{k.number}</span>
                <span class="min-w-0 flex-1 truncate text-[13px]">{k.description}</span>
                <StockBar key={k} compact />
              </button>
            </li>
          {/each}
        </ul>
      {:else}
        <div class="alert alert-warning">Aucune clé n'ouvre cet accès : il ne pourra pas être proposé lors d'un prêt.</div>
      {/if}
    </section>

    <section>
      <h3 class="section-title mb-2">Peuvent l'ouvrir actuellement ({d.holders.length})</h3>
      {#if d.holders.length}
        <ul class="space-y-1.5">
          {#each d.holders as h (h.borrowerId + h.keyNumber)}
            <li class="flex items-center gap-2 text-[13px]">
              <button class="min-w-0 flex-1 cursor-pointer truncate text-left font-medium hover:text-primary" onclick={() => openDetail('borrower', h.borrowerId)}>{h.borrowerName}</button>
              <span class="chip chip-key">{h.keyNumber}</span>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-[13px] text-muted">Personne ne détient de clé de cet accès.</p>
      {/if}
    </section>

    {#if d.access.notes}
      <section>
        <h3 class="section-title mb-2">Observations</h3>
        <p class="text-[13px] whitespace-pre-line" data-selectable>{d.access.notes}</p>
      </section>
    {/if}
  {/if}

  {#snippet footer()}
    {#if d}
      <button class="btn btn-secondary" onclick={() => (editOpen = true)}><Pencil size={14} />Modifier</button>
      <button class="btn btn-danger-ghost ml-auto" onclick={remove}><Trash2 size={14} />Supprimer</button>
    {/if}
  {/snippet}
</Drawer>

<AccessForm bind:open={editOpen} accessId={id} />
