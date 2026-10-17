<script lang="ts">
  // Message affiché quand une liste est vide, avec actions éventuelles.
  import type { Component, Snippet } from 'svelte'

  let {
    icon: Icon,
    title,
    text = '',
    children,
  }: { icon: Component<{ size?: number; class?: string }>; title: string; text?: string; children?: Snippet } = $props()
</script>

<div class="flex flex-col items-center justify-center px-6 py-14 text-center">
  <div class="mb-4 flex size-12 items-center justify-center rounded-full bg-primary-soft text-primary">
    <Icon size={22} />
  </div>
  <p class="text-[15px] font-semibold">{title}</p>
  {#if text}<p class="mt-1 max-w-md text-[13px] text-muted">{text}</p>{/if}
  {#if children}<div class="mt-5 flex flex-wrap justify-center gap-2">{@render children()}</div>{/if}
</div>
