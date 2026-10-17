<script lang="ts">
  // À propos : version, fonctionnalités, technologies, auteur, licence.
  import { KeyRound } from '@lucide/svelte'
  import { Api } from '../lib/api'
  import { appSettings } from '../lib/state.svelte'

  let version = $state('')
  void Api.GetAppInfo().then((i) => (version = i.version))

  const features = [
    'Prêt par besoin : les portes à ouvrir donnent le trousseau minimal',
    'Bon de remise PDF à chaque prêt, rééditable à jour en un clic',
    "Alerte d'inventaire en cas de sur-prêt (signalé, jamais bloqué)",
    "Inventaire rapide : stock compté + clés prêtées, rapport PDF des écarts",
    'Fiches détaillées des clés, accès et détenteurs',
    'Recherche globale, historique filtrable, exports Excel',
    'Détection des clés en doublon',
    'Reprise des bases V1, V2 et V3 par un assistant guidé',
    'Sauvegardes atomiques et restauration en un clic',
    'Utilisation en réseau, plusieurs postes simultanés',
  ]
</script>

<div class="mx-auto max-w-2xl">
  <div class="card p-8 text-center">
    <div class="mx-auto flex size-14 items-center justify-center rounded-2xl bg-primary text-primary-fg"><KeyRound size={26} /></div>
    <h1 class="mt-4 text-[22px] font-semibold">{appSettings.appName}</h1>
    <p class="mt-1 text-muted">Version {version}</p>
    <p class="mt-3 text-[13.5px]">Gestion du parc de clés — {appSettings.establishment}{appSettings.address ? ', ' + appSettings.address : ''}.</p>
  </div>

  <div class="mt-5 grid grid-cols-2 gap-5">
    <section class="card p-6">
      <h2 class="section-title mb-3">Fonctionnalités</h2>
      <ul class="space-y-2 text-[13px]">
        {#each features as f}<li class="flex gap-2.5"><span class="mt-[7px] size-1.5 shrink-0 rounded-full bg-primary"></span>{f}</li>{/each}
      </ul>
    </section>
    <div class="space-y-5">
      <section class="card p-6">
        <h2 class="section-title mb-3">Technologies</h2>
        <ul class="space-y-1.5 text-[13px]">
          <li>Go et Wails (fenêtre native)</li>
          <li>Svelte, TypeScript et Tailwind CSS (interface)</li>
          <li>SQLite en mode WAL (pur Go)</li>
          <li>Exécutable unique, sans installation ni connexion internet</li>
        </ul>
      </section>
      <section class="card p-6">
        <h2 class="section-title mb-3">Conception</h2>
        <p class="text-[13px]" data-selectable>COLLET David<br />Secrétaire Général<br />Collège Victor Hugo — Chauny (02300)<br />david.collet@ac-amiens.fr</p>
        <p class="mt-3 text-[12.5px] text-muted">Code réalisé par IA (Claude, Anthropic) sous sa direction.</p>
      </section>
    </div>
  </div>

  <section class="card mt-5 p-6">
    <h2 class="section-title mb-3">Licence</h2>
    <p class="text-[13px] text-muted">
      Distribué sous licence MIT. Permission est accordée, gratuitement, à toute personne obtenant une copie de ce logiciel, de l'utiliser, le copier, le modifier, le
      fusionner, le publier, le distribuer et/ou le vendre, sous réserve de conserver la présente notice dans toutes les copies.
    </p>
    <p class="mt-4 text-center text-[12.5px] text-muted">© 2026 COLLET David — MIT License</p>
  </section>
</div>
