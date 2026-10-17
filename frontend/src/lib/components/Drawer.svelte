<script lang="ts">
  // Panneau latéral droit des fiches détaillées (clé, accès, détenteur).
  import type { Snippet } from 'svelte'
  import { X } from '@lucide/svelte'

  let {
    onclose,
    header,
    children,
    footer,
  }: { onclose: () => void; header: Snippet; children: Snippet; footer?: Snippet } = $props()
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && onclose()} />

<div class="fixed inset-0 z-30 flex justify-end">
  <button class="absolute inset-0 cursor-default bg-black/25" aria-label="Fermer la fiche" tabindex="-1" onclick={onclose}></button>
  <aside class="drawer relative flex h-full w-[540px] max-w-full flex-col border-l border-line bg-surface shadow-pop">
    <header class="flex items-start justify-between gap-3 border-b border-line px-6 py-5">
      <div class="min-w-0 flex-1">{@render header()}</div>
      <button class="btn btn-ghost btn-icon -mr-2" onclick={onclose} aria-label="Fermer"><X size={18} /></button>
    </header>
    <div class="min-h-0 flex-1 space-y-6 overflow-y-auto px-6 py-5">
      {@render children()}
    </div>
    {#if footer}
      <footer class="flex flex-wrap items-center gap-2 border-t border-line px-6 py-3.5">
        {@render footer()}
      </footer>
    {/if}
  </aside>
</div>

<style>
  .drawer {
    animation: slide-in 0.18s ease-out;
  }
  @keyframes slide-in {
    from {
      transform: translateX(24px);
      opacity: 0;
    }
  }
</style>
