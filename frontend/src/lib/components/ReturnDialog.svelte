<script lang="ts">
  // Retour d'une ou plusieurs clés, avec l'état constaté.
  import { Api, run, type Loan } from '../api'
  import { fmtDate, fmtDays, plural } from '../format'
  import Modal from './Modal.svelte'

  let { open = $bindable(false), loans = [], ondone }: { open?: boolean; loans?: Loan[]; ondone?: () => void } = $props()

  let condition = $state('')
  let saving = $state(false)
  const presets = ['Bon état', 'Usée', 'Abîmée', 'Trousseau incomplet']

  $effect(() => {
    if (open) condition = ''
  })

  async function confirmReturn() {
    saving = true
    const ok = await run(
      Api.ReturnLoans(
        loans.map((l) => l.id),
        condition,
      ).then(() => true),
      loans.length > 1 ? `${loans.length} clés rendues.` : `Clé ${loans[0]?.keyNumber} rendue.`,
    )
    saving = false
    if (ok) {
      open = false
      ondone?.()
    }
  }
</script>

<Modal bind:open title="Enregistrer le retour" subtitle={loans[0] ? `${loans[0].borrowerName} rend ${plural(loans.length, 'clé')}.` : ''} size="sm">
  <ul class="mb-5 divide-y divide-line rounded-control border border-line">
    {#each loans as l (l.id)}
      <li class="flex items-center gap-3 px-3 py-2.5">
        <span class="chip chip-key">{l.keyNumber}</span>
        <span class="min-w-0 flex-1 truncate text-[13px]">{l.keyDescription}</span>
        <span class="text-[12px] text-muted">depuis le {fmtDate(l.loanDate)} · {fmtDays(l.days)}</span>
      </li>
    {/each}
  </ul>
  <label class="block">
    <span class="label">État constaté (facultatif)</span>
    <input class="input" bind:value={condition} placeholder="ex. Bon état" />
  </label>
  <div class="mt-2 flex flex-wrap gap-1.5">
    {#each presets as p}
      <button class="chip cursor-pointer hover:border-primary hover:text-primary" onclick={() => (condition = p)}>{p}</button>
    {/each}
  </div>
  {#snippet footer()}
    <button class="btn btn-secondary" onclick={() => (open = false)}>Annuler</button>
    <button class="btn btn-primary" disabled={saving || !loans.length} onclick={confirmReturn}>Confirmer le retour</button>
  {/snippet}
</Modal>
