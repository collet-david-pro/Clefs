<script lang="ts">
  // Jauge de stock d'une clé : sorties / réserve / disponibles.
  import type { Key } from '../api'

  let { key, compact = false }: { key: Key; compact?: boolean } = $props()

  const total = $derived(Math.max(key.total, key.loaned + key.reserve, 1))
  const pct = (n: number) => `${Math.max(0, (n / total) * 100)}%`
  const tone = $derived(key.available < 0 ? 'text-danger' : key.available === 0 ? 'text-warning' : 'text-success')
</script>

<div class="flex items-center gap-3 {compact ? 'min-w-36' : ''}">
  <div class="flex h-1.5 flex-1 overflow-hidden rounded-full bg-surface-2" title="{key.loaned} sortie(s), {key.reserve} en réserve, {key.available} disponible(s) sur {key.total}">
    <div class="bg-primary" style:width={pct(key.loaned)}></div>
    <div class="bg-muted/40" style:width={pct(key.reserve)}></div>
  </div>
  <span class="w-16 text-right text-[12.5px] font-semibold tabular-nums {tone}">
    {key.available}<span class="font-normal text-muted"> / {key.total}</span>
  </span>
</div>
