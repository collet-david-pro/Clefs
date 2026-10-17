<script lang="ts">
  // Notifications non bloquantes, en bas à droite.
  import { CircleCheck, CircleAlert, Info, X } from '@lucide/svelte'
  import { dismiss, toasts } from '../state.svelte'

  const icons = { success: CircleCheck, error: CircleAlert, info: Info }
  const colors = { success: 'text-success', error: 'text-danger', info: 'text-info' }
</script>

<div class="pointer-events-none fixed right-5 bottom-5 z-[60] flex w-[380px] flex-col gap-2">
  {#each toasts as t (t.id)}
    {@const Icon = icons[t.kind]}
    <div class="toast pointer-events-auto flex items-start gap-3 rounded-card border border-line bg-surface px-4 py-3 shadow-pop" role="status">
      <Icon size={18} class="mt-0.5 shrink-0 {colors[t.kind]}" />
      <p class="min-w-0 flex-1 text-[13px] leading-snug break-words" data-selectable>{t.message}</p>
      {#if t.action}
        <button
          class="btn btn-sm btn-secondary -my-1"
          onclick={() => {
            t.action?.run()
            dismiss(t.id)
          }}>{t.action.label}</button
        >
      {/if}
      <button class="-mr-1 cursor-pointer text-muted hover:text-ink" onclick={() => dismiss(t.id)} aria-label="Fermer"><X size={15} /></button>
    </div>
  {/each}
</div>

<style>
  .toast {
    animation: rise 0.2s ease-out;
  }
  @keyframes rise {
    from {
      transform: translateY(8px);
      opacity: 0;
    }
  }
</style>
