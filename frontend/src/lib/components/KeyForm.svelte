<script lang="ts">
  // Création / modification d'une clé, avec les accès qu'elle ouvre.
  import { Api, run, type Access } from '../api'
  import { keyCategoryLabel } from '../format'
  import AccessPicker from './AccessPicker.svelte'
  import Modal from './Modal.svelte'

  let { open = $bindable(false), keyId = 0, onsaved }: { open?: boolean; keyId?: number; onsaved?: (id: number) => void } = $props()

  let form = $state({ number: '', description: '', category: 'simple', storageLocation: '', notes: '', total: 1, reserve: 0 })
  let accessIds = $state<number[]>([])
  let accesses = $state<Access[]>([])
  let categories = $state<string[]>(['simple', 'trousseau', 'badge', 'passe'])
  let saving = $state(false)
  let touched = $state(false)

  $effect(() => {
    if (open) void load()
  })

  async function load() {
    touched = false
    const [list, cats] = await Promise.all([run(Api.ListAccesses()), run(Api.GetKeyCategories())])
    accesses = list ?? []
    if (cats) categories = cats
    if (keyId) {
      const d = await run(Api.GetKeyDetail(keyId))
      if (!d) return
      const k = d.key
      form = { number: k.number, description: k.description, category: k.category || 'simple', storageLocation: k.storageLocation, notes: k.notes, total: k.total, reserve: k.reserve }
      accessIds = [...k.accessIds]
    } else {
      form = { number: '', description: '', category: 'simple', storageLocation: '', notes: '', total: 1, reserve: 0 }
      accessIds = []
    }
  }

  const errors = $derived({
    number: !form.number.trim() ? 'Le numéro est obligatoire.' : '',
    total: !(form.total >= 0) ? 'Le total ne peut pas être négatif.' : '',
    reserve: form.reserve < 0 ? 'La réserve ne peut pas être négative.' : form.reserve > form.total ? 'La réserve ne peut pas dépasser la quantité totale.' : '',
  })
  const valid = $derived(!errors.number && !errors.total && !errors.reserve)

  async function save() {
    touched = true
    if (!valid) return
    saving = true
    const id = await run(Api.SaveKey({ id: keyId, ...form, total: Number(form.total), reserve: Number(form.reserve), accessIds }), keyId ? 'Clé modifiée.' : 'Clé créée.')
    saving = false
    if (id !== undefined) {
      open = false
      onsaved?.(id)
    }
  }
</script>

<Modal bind:open title={keyId ? 'Modifier la clé' : 'Nouvelle clé'} subtitle="Les champs marqués * sont obligatoires." size="lg">
  <div class="grid grid-cols-[1fr_1fr] gap-x-6 gap-y-4">
    <div class="space-y-4">
      <div class="grid grid-cols-[1fr_1.3fr] gap-3">
        <label>
          <span class="label">Numéro *</span>
          <!-- svelte-ignore a11y_autofocus -->
          <input class="input font-mono {touched && errors.number ? 'input-error' : ''}" bind:value={form.number} placeholder="ex. K-012" autofocus />
          {#if touched && errors.number}<p class="error-text">{errors.number}</p>{/if}
        </label>
        <label>
          <span class="label">Catégorie</span>
          <select class="input" bind:value={form.category}>
            {#each categories as c}<option value={c}>{keyCategoryLabel[c] ?? c}</option>{/each}
          </select>
        </label>
      </div>
      <label class="block">
        <span class="label">Description</span>
        <input class="input" bind:value={form.description} placeholder="ex. Passe du bâtiment B" />
      </label>
      <div class="grid grid-cols-2 gap-3">
        <label>
          <span class="label">Exemplaires au total *</span>
          <input class="input {errors.total ? 'input-error' : ''}" type="number" min="0" bind:value={form.total} />
          {#if errors.total}<p class="error-text">{errors.total}</p>{:else if form.total === 0}<p class="hint">0 : toutes perdues ou cassées.</p>{/if}
        </label>
        <label>
          <span class="label">Dont en réserve</span>
          <input class="input {errors.reserve ? 'input-error' : ''}" type="number" min="0" bind:value={form.reserve} />
          {#if errors.reserve}<p class="error-text">{errors.reserve}</p>{:else}<p class="hint">Jamais prêtés (double de sécurité).</p>{/if}
        </label>
      </div>
      <label class="block">
        <span class="label">Emplacement de rangement</span>
        <input class="input" bind:value={form.storageLocation} placeholder="ex. Armoire à clés, crochet 12" />
      </label>
      <label class="block">
        <span class="label">Observations</span>
        <textarea class="input" rows="3" bind:value={form.notes}></textarea>
      </label>
    </div>
    <div class="flex min-h-0 flex-col">
      <span class="label">Accès ouverts par cette clé</span>
      <AccessPicker {accesses} bind:selected={accessIds} class="h-[360px]" />
    </div>
  </div>
  {#snippet footer()}
    <button class="btn btn-secondary" onclick={() => (open = false)}>Annuler</button>
    <button class="btn btn-primary" disabled={saving} onclick={save}>{keyId ? 'Enregistrer' : 'Créer la clé'}</button>
  {/snippet}
</Modal>
