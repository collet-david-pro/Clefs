<script lang="ts">
  // Fenêtre modale : fond assombri, fermeture par Échap ou clic à l'extérieur.
  import type { Snippet } from 'svelte'
  import { X } from '@lucide/svelte'

  let {
    open = $bindable(false),
    title,
    subtitle = '',
    size = 'md',
    dismissable = true,
    children,
    footer,
  }: {
    open?: boolean
    title: string
    subtitle?: string
    size?: 'sm' | 'md' | 'lg' | 'xl'
    dismissable?: boolean
    children: Snippet
    footer?: Snippet
  } = $props()

  const widths = { sm: 'max-w-md', md: 'max-w-xl', lg: 'max-w-3xl', xl: 'max-w-5xl' }

  function close() {
    if (dismissable) open = false
  }
</script>

<svelte:window onkeydown={(e) => open && e.key === 'Escape' && close()} />

{#if open}
  <div class="fixed inset-0 z-40 flex items-center justify-center p-6">
    <button class="absolute inset-0 cursor-default bg-black/40 backdrop-blur-[2px]" aria-label="Fermer" tabindex="-1" onclick={close}></button>
    <div
      class="relative flex max-h-full w-full {widths[size]} flex-col overflow-hidden rounded-card border border-line bg-surface shadow-pop"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <header class="flex items-start justify-between gap-4 border-b border-line px-6 py-4">
        <div>
          <h2 class="text-[16px] font-semibold">{title}</h2>
          {#if subtitle}<p class="mt-0.5 text-[13px] text-muted">{subtitle}</p>{/if}
        </div>
        {#if dismissable}
          <button class="btn btn-ghost btn-icon -mr-2 -mt-1" onclick={close} aria-label="Fermer"><X size={18} /></button>
        {/if}
      </header>
      <div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
        {@render children()}
      </div>
      {#if footer}
        <footer class="flex items-center justify-end gap-2 border-t border-line bg-surface-2/50 px-6 py-3.5">
          {@render footer()}
        </footer>
      {/if}
    </div>
  </div>
{/if}
