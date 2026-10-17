<script lang="ts">
  // Création / modification d'un accès (porte, portail, zone).
  import { Api, run, type Building } from '../api'
  import Modal from './Modal.svelte'

  let { open = $bindable(false), accessId = 0, onsaved }: { open?: boolean; accessId?: number; onsaved?: (id: number) => void } = $props()

  let form = $state({ name: '', type: '', buildingId: 0, floor: '', category: '', notes: '' })
  let buildings = $state<Building[]>([])
  let saving = $state(false)
  let touched = $state(false)
  let newBuilding = $state('')

  $effect(() => {
    if (open) void load()
  })

  async function load() {
    touched = false
    newBuilding = ''
    buildings = (await run(Api.ListBuildings())) ?? []
    if (accessId) {
      const d = await run(Api.GetAccessDetail(accessId))
      if (!d) return
      const a = d.access
      form = { name: a.name, type: a.type, buildingId: a.buildingId, floor: a.floor, category: a.category, notes: a.notes }
    } else {
      form = { name: '', type: '', buildingId: buildings.length === 1 ? buildings[0].id : 0, floor: '', category: '', notes: '' }
    }
  }

  async function addBuilding() {
    const id = await run(Api.SaveBuilding(0, newBuilding), 'Bâtiment créé.')
    if (id === undefined) return
    buildings = (await run(Api.ListBuildings())) ?? []
    form.buildingId = id
    newBuilding = ''
  }

  const errors = $derived({
    name: !form.name.trim() ? 'La désignation est obligatoire.' : '',
    building: !form.buildingId ? 'Choisissez un bâtiment.' : '',
  })

  async function save() {
    touched = true
    if (errors.name || errors.building) return
    saving = true
    const id = await run(Api.SaveAccess({ id: accessId, ...form, buildingId: Number(form.buildingId) }), accessId ? 'Accès modifié.' : 'Accès créé.')
    saving = false
    if (id !== undefined) {
      open = false
      onsaved?.(id)
    }
  }
</script>

<Modal bind:open title={accessId ? "Modifier l'accès" : 'Nouvel accès'} subtitle="Une porte, un portail ou une zone fermée à clé.">
  <div class="space-y-4">
    <label class="block">
      <span class="label">Désignation *</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input {touched && errors.name ? 'input-error' : ''}" bind:value={form.name} placeholder="ex. Salle 104, Portail parking" autofocus />
      {#if touched && errors.name}<p class="error-text">{errors.name}</p>{/if}
    </label>
    <div>
      <span class="label">Bâtiment *</span>
      {#if buildings.length}
        <select class="input {touched && errors.building ? 'input-error' : ''}" bind:value={form.buildingId}>
          <option value={0} disabled>Choisir…</option>
          {#each buildings as b}<option value={b.id}>{b.name}</option>{/each}
        </select>
        {#if touched && errors.building}<p class="error-text">{errors.building}</p>{/if}
      {/if}
      <div class="mt-2 flex gap-2">
        <input class="input" bind:value={newBuilding} placeholder={buildings.length ? 'ou créer un nouveau bâtiment…' : 'Nom du premier bâtiment'} />
        <button class="btn btn-secondary" disabled={!newBuilding.trim()} onclick={addBuilding}>Créer</button>
      </div>
    </div>
    <div class="grid grid-cols-3 gap-3">
      <label>
        <span class="label">Étage / niveau</span>
        <input class="input" bind:value={form.floor} placeholder="RDC, R+1…" list="floors" />
        <datalist id="floors"><option value="Sous-sol"></option><option value="RDC"></option><option value="R+1"></option><option value="R+2"></option><option value="R+3"></option></datalist>
      </label>
      <label>
        <span class="label">Catégorie</span>
        <input class="input" bind:value={form.category} placeholder="Salle de classe…" list="access-categories" />
        <datalist id="access-categories">
          <option value="Salle de classe"></option><option value="Bureau"></option><option value="Local technique"></option>
          <option value="Laboratoire"></option><option value="Stockage"></option><option value="Extérieur"></option>
        </datalist>
      </label>
      <label>
        <span class="label">Type</span>
        <input class="input" bind:value={form.type} placeholder="Porte, portail…" />
      </label>
    </div>
    <label class="block">
      <span class="label">Observations</span>
      <textarea class="input" rows="3" bind:value={form.notes}></textarea>
    </label>
  </div>
  {#snippet footer()}
    <button class="btn btn-secondary" onclick={() => (open = false)}>Annuler</button>
    <button class="btn btn-primary" disabled={saving} onclick={save}>{accessId ? 'Enregistrer' : "Créer l'accès"}</button>
  {/snippet}
</Modal>
