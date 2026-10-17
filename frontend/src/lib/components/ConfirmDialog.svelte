<script lang="ts">
  // Confirmation des actions sensibles (suppression, restauration, retour…).
  import { TriangleAlert } from '@lucide/svelte'
  import { closeConfirm, confirmState } from '../state.svelte'

  let typed = $state('')
  const o = $derived(confirmState.options)
  const canConfirm = $derived(!o?.typeToConfirm || typed.trim().toUpperCase() === o.typeToConfirm)

  $effect(() => {
    if (confirmState.open) typed = ''
  })

  function onkey(e: KeyboardEvent) {
    if (!confirmState.open) return
    if (e.key === 'Escape') closeConfirm(false)
    if (e.key === 'Enter' && canConfirm) closeConfirm(true)
  }
</script>

<svelte:window onkeydown={onkey} />

{#if confirmState.open && o}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-6">
    <button class="absolute inset-0 cursor-default bg-black/40" aria-label="Annuler" tabindex="-1" onclick={() => closeConfirm(false)}></button>
    <div class="relative w-full max-w-md rounded-card border border-line bg-surface p-6 shadow-pop" role="alertdialog" aria-modal="true">
      <div class="flex gap-4">
        {#if o.danger}
          <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-danger-soft text-danger">
            <TriangleAlert size={20} />
          </div>
        {/if}
        <div class="min-w-0 flex-1">
          <h2 class="text-[16px] font-semibold">{o.title}</h2>
          <p class="mt-1.5 text-[13.5px] whitespace-pre-line text-muted">{o.message}</p>
          {#if o.typeToConfirm}
            <label class="mt-4 block">
              <span class="label">Tapez <strong>{o.typeToConfirm}</strong> pour confirmer</span>
              <!-- svelte-ignore a11y_autofocus -->
              <input class="input" bind:value={typed} autofocus />
            </label>
          {/if}
        </div>
      </div>
      <div class="mt-6 flex justify-end gap-2">
        <button class="btn btn-secondary" onclick={() => closeConfirm(false)}>Annuler</button>
        <button class="btn {o.danger ? 'btn-danger' : 'btn-primary'}" disabled={!canConfirm} onclick={() => closeConfirm(true)}>
          {o.confirmLabel ?? 'Confirmer'}
        </button>
      </div>
    </div>
  </div>
{/if}
