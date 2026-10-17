<script lang="ts">
  // Création / modification d'un détenteur.
  import { Api, run } from '../api'
  import { borrowerStatusLabel } from '../format'
  import Modal from './Modal.svelte'

  let { open = $bindable(false), borrowerId = 0, onsaved }: { open?: boolean; borrowerId?: number; onsaved?: (id: number) => void } = $props()

  let form = $state({ name: '', email: '', status: 'permanent', phone: '' })
  let saving = $state(false)
  let touched = $state(false)
  const statuses = Object.keys(borrowerStatusLabel)

  $effect(() => {
    if (open) void load()
  })

  async function load() {
    touched = false
    if (borrowerId) {
      const d = await run(Api.GetBorrowerDetail(borrowerId))
      if (!d) return
      const b = d.borrower
      form = { name: b.name, email: b.email, status: b.status || 'permanent', phone: b.phone }
    } else {
      form = { name: '', email: '', status: 'permanent', phone: '' }
    }
  }

  const nameError = $derived(!form.name.trim() ? 'Le nom est obligatoire.' : '')
  const emailError = $derived(form.email.trim() && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim()) ? "L'adresse e-mail semble incorrecte." : '')

  async function save() {
    touched = true
    if (nameError || emailError) return
    saving = true
    const id = await run(Api.SaveBorrower({ id: borrowerId, ...form }), borrowerId ? 'Détenteur modifié.' : 'Détenteur créé.')
    saving = false
    if (id !== undefined) {
      open = false
      onsaved?.(id)
    }
  }
</script>

<Modal bind:open title={borrowerId ? 'Modifier le détenteur' : 'Nouveau détenteur'} size="sm">
  <div class="space-y-4">
    <label class="block">
      <span class="label">Nom et prénom *</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input {touched && nameError ? 'input-error' : ''}" bind:value={form.name} placeholder="ex. Dupont Jean" autofocus />
      {#if touched && nameError}<p class="error-text">{nameError}</p>{/if}
    </label>
    <div>
      <span class="label">Statut</span>
      <div class="segmented flex w-full">
        {#each statuses as s}
          <button class="flex-1" aria-pressed={form.status === s} onclick={() => (form.status = s)}>{borrowerStatusLabel[s]}</button>
        {/each}
      </div>
    </div>
    <label class="block">
      <span class="label">E-mail</span>
      <input class="input {emailError ? 'input-error' : ''}" type="email" bind:value={form.email} placeholder="prenom.nom@ac-amiens.fr" />
      {#if emailError}<p class="error-text">{emailError}</p>{/if}
    </label>
    <label class="block">
      <span class="label">Téléphone</span>
      <input class="input" type="tel" bind:value={form.phone} />
    </label>
  </div>
  {#snippet footer()}
    <button class="btn btn-secondary" onclick={() => (open = false)}>Annuler</button>
    <button class="btn btn-primary" disabled={saving} onclick={save}>{borrowerId ? 'Enregistrer' : 'Créer le détenteur'}</button>
  {/snippet}
</Modal>
