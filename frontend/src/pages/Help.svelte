<script lang="ts">
  // Mode d'emploi intégré : guide pas à pas, du premier lancement à
  // l'utilisation quotidienne, avec un accès direct à chaque écran.
  import { ArrowRight, ChevronDown } from '@lucide/svelte'
  import PageHeader from '../lib/components/PageHeader.svelte'
  import { navigate, type Page } from '../lib/state.svelte'

  type Section = { title: string; intro?: string; steps: string[]; page?: Page; pageLabel?: string }

  const sections: Section[] = [
    {
      title: '1. Premier lancement',
      intro: "L'application est un seul fichier .exe : rien à installer.",
      steps: [
        "Placez l'exécutable dans un dossier qui lui est propre (par exemple C:\\Clefs-V4\\ ou un dossier du partage réseau).",
        'Au lancement, il crée à côté de lui : clefs.db (la base de données), backups\\ (les sauvegardes) et documents\\ (les PDF et exports Excel).',
        "Si Windows affiche une alerte (exécutable non signé) : « Informations complémentaires » puis « Exécuter quand même ».",
        "Sur une base vide, l'assistant d'import s'ouvre automatiquement.",
      ],
    },
    {
      title: '2. Reprendre les données d’une ancienne version',
      intro: 'Les versions 1, 2 et 3 sont reconnues automatiquement.',
      steps: [
        "Fermez l'ancienne application sur tous les postes.",
        "Dans l'assistant, choisissez le fichier clefs.db de l'ancien dossier (ou une sauvegarde .db).",
        "Vérifiez l'aperçu (nombre de clés, d'accès, de détenteurs, de prêts), puis cliquez sur « Importer ».",
        "L'ancien fichier n'est jamais modifié. Faites ensuite une première sauvegarde.",
      ],
      page: 'import',
      pageLabel: "Ouvrir l'assistant d'import",
    },
    {
      title: '3. Paramétrer l’établissement',
      steps: [
        "Paramètres → « Établissement et documents » : nom et adresse de l'établissement, nom de l'application, logo (PNG ou JPEG).",
        "Bons de remise : titre, texte d'engagement, libellés des signatures, signataire, mention en bas de page. Dans le texte d'engagement, {detenteur}, {cles} et {la_les} sont remplacés automatiquement.",
        'Prêts : type de prêt et durée de retour proposés par défaut, nom de l’agent inscrit sur les prêts (sinon, le nom de la session Windows).',
        "« Aperçu d'un bon » montre le rendu avant d'enregistrer. Ces réglages sont communs à tous les postes.",
      ],
      page: 'settings',
      pageLabel: 'Ouvrir les paramètres',
    },
    {
      title: '4. Créer le référentiel',
      intro: 'Sans import, créez les éléments dans cet ordre :',
      steps: [
        'Les bâtiments.',
        'Les accès : chaque porte, portail ou zone fermée à clé, avec son bâtiment (étage et catégorie facultatifs).',
        "Les clés : numéro, nombre d'exemplaires, réserve (doubles jamais prêtés), rangement, et les accès que chaque clé ouvre.",
        'Les détenteurs : un par un, ou en important une liste Excel (page Détenteurs → Importer, modèle fourni).',
      ],
      page: 'buildings',
      pageLabel: 'Ouvrir les bâtiments',
    },
    {
      title: '5. Prêter des clés',
      intro: 'Le prêt part du besoin : quelles portes la personne doit-elle ouvrir ?',
      steps: [
        '« Nouveau prêt » (ou Ctrl + N), puis choisissez la personne (ou créez-la sur place).',
        'Cochez les portes à ouvrir : le trousseau le plus réduit est calculé automatiquement.',
        'Retirez une clé proposée (×) ou ajoutez-en une à la main si besoin.',
        'Indiquez le type de prêt et, si besoin, la date de retour prévue, puis validez.',
        'Le bon de remise PDF (une page A4) est créé : ouvrez-le, imprimez-le et faites-le signer.',
        "Prêter plus d'exemplaires que le stock reste possible : la clé est alors signalée « erreur d'inventaire ».",
      ],
      page: 'new-loan',
      pageLabel: 'Faire un prêt',
    },
    {
      title: '6. Enregistrer un retour',
      steps: [
        'Depuis Prêts en cours, la fiche du détenteur ou la fiche de la clé : bouton « Retour ».',
        'Plusieurs clés d’une même personne se rendent en une fois (« Tout rendre »).',
        "Notez si besoin l'état constaté (bon état, abîmée…) : il reste dans l'historique.",
      ],
      page: 'loans',
      pageLabel: 'Ouvrir les prêts en cours',
    },
    {
      title: '7. Faire un inventaire',
      intro: 'Pour remettre le stock à jour après une clé retrouvée, cassée ou perdue.',
      steps: [
        "Onglet Inventaire : pour chaque clé, saisissez le nombre d'exemplaires présents dans l'armoire, réserve comprise.",
        'Le nouveau total est calculé seul : exemplaires comptés + exemplaires prêtés. Les écarts s’affichent en vert ou en rouge.',
        'Entrée passe à la clé suivante ; « Pré-remplir avec le stock attendu » permet de ne corriger que les écarts. Les clés non saisies ne changent pas.',
        "À la validation, un rapport d'inventaire PDF liste les écarts. L'inventaire corrige aussi les « erreurs d'inventaire ».",
      ],
      page: 'inventory',
      pageLabel: "Ouvrir l'inventaire",
    },
    {
      title: '8. Retrouver une information',
      steps: [
        'La recherche en haut de l’écran (Ctrl + K) trouve une clé, une porte ou une personne en quelques lettres.',
        'Un clic sur une ligne ouvre sa fiche : stock, accès, détenteurs, historique et actions.',
        'Tableau de bord : les points à traiter (retards, erreurs d’inventaire, doublons) et les derniers mouvements.',
        'Qui a quoi ?, Plan de clés (bâtiment par bâtiment), Doublons et Historique (filtrable) complètent la consultation.',
      ],
      page: 'dashboard',
      pageLabel: 'Ouvrir le tableau de bord',
    },
    {
      title: '9. Documents et exports',
      steps: [
        'Bon de remise : créé à chaque prêt, toujours sur une seule page A4 ; réédition d’un bon récapitulatif à jour depuis la fiche du détenteur.',
        'Prêts en cours → « Éditer tous les bons (PDF) » : un seul fichier avec un sommaire puis un bon par détenteur, pour l’archivage.',
        'Bilan du stock (page Clés), plan de clés (page Plan de clés) et rapport d’inventaire, au format PDF.',
        'Exports Excel : clés, détenteurs, prêts en cours et historique filtré.',
        'Tout est rangé dans le dossier documents\\ (Paramètres → Données → Ouvrir les documents).',
      ],
    },
    {
      title: '10. Sauvegardes et sécurité',
      steps: [
        '« Sauvegarder maintenant » crée une copie complète dans backups\\. Faites-en régulièrement, et toujours avant une mise à jour.',
        "Restaurer une sauvegarde remplace les données actuelles ; une copie de sécurité de l'état actuel est faite avant.",
        'Une clé prêtée, ou une personne qui détient des clés, ne peut pas être supprimée tant que les retours ne sont pas enregistrés.',
        'Les actions sensibles demandent toujours une confirmation.',
      ],
      page: 'backups',
      pageLabel: 'Ouvrir les sauvegardes',
    },
    {
      title: '11. Utilisation en réseau',
      steps: [
        "Placez le dossier de l'application sur un partage réseau : plusieurs postes peuvent l'utiliser en même temps.",
        'Les paramètres sont communs à tous les postes ; le thème clair ou sombre est propre à chaque poste.',
        'Pour fermer l’application : bouton « Quitter » en bas du menu, ou la croix de la fenêtre.',
      ],
    },
  ]

  const faq: [string, string][] = [
    ["Que signifie « erreur d'inventaire » ?", "Plus d'exemplaires sont prêtés que le stock déclaré ne le permet. Le prêt n'est pas bloqué, mais il faut vérifier : un inventaire remet le stock à jour."],
    ['Pourquoi une clé n’est-elle pas proposée au prêt ?', "Elle n'a plus d'exemplaire disponible, ou elle n'est associée à aucune des portes cochées. Vous pouvez toujours l'ajouter à la main."],
    ['Faut-il installer quelque chose sur les postes ?', "Non. L'affichage utilise Microsoft Edge WebView2, fourni avec Windows 10 et 11."],
    ['Où sont mes données ?', 'Dans clefs.db, à côté de l’exécutable. Paramètres → Données indique les chemins exacts.'],
  ]

  let open = $state<Record<number, boolean>>({ 0: true })
  let openFaq = $state<Record<number, boolean>>({})
</script>

<PageHeader title="Aide" subtitle="Guide pas à pas : du premier lancement à l'utilisation quotidienne.">
  {#snippet actions()}
    <button class="btn btn-ghost" onclick={() => (open = Object.fromEntries(sections.map((_, i) => [i, true])))}>Tout déplier</button>
  {/snippet}
</PageHeader>

<div class="max-w-3xl space-y-3">
  {#each sections as s, i}
    <section class="card overflow-hidden">
      <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-4 text-left" onclick={() => (open[i] = !open[i])}>
        <h2 class="flex-1 text-[15px] font-semibold">{s.title}</h2>
        <ChevronDown size={17} class="text-muted transition-transform {open[i] ? 'rotate-180' : ''}" />
      </button>
      {#if open[i]}
        <div class="border-t border-line px-5 py-4">
          {#if s.intro}<p class="mb-3 text-[13.5px] text-muted">{s.intro}</p>{/if}
          <ol class="space-y-2.5 text-[13.5px]">
            {#each s.steps as step, j}
              <li class="flex gap-3">
                <span class="flex size-5 shrink-0 items-center justify-center rounded-full bg-primary-soft text-[11px] font-semibold text-primary">{j + 1}</span>
                <span>{step}</span>
              </li>
            {/each}
          </ol>
          {#if s.page}
            <button class="btn btn-secondary btn-sm mt-4" onclick={() => navigate(s.page!)}>{s.pageLabel}<ArrowRight size={14} /></button>
          {/if}
        </div>
      {/if}
    </section>
  {/each}

  <section class="card overflow-hidden">
    <h2 class="px-5 pt-4 pb-2 text-[15px] font-semibold">Questions fréquentes</h2>
    {#each faq as [q, a], i}
      <div class="border-t border-line">
        <button class="flex w-full cursor-pointer items-center gap-3 px-5 py-3 text-left" onclick={() => (openFaq[i] = !openFaq[i])}>
          <span class="flex-1 text-[13.5px] font-medium">{q}</span>
          <ChevronDown size={15} class="text-muted transition-transform {openFaq[i] ? 'rotate-180' : ''}" />
        </button>
        {#if openFaq[i]}<p class="px-5 pb-3 text-[13.5px] text-muted">{a}</p>{/if}
      </div>
    {/each}
  </section>

  <section class="card p-5">
    <h2 class="text-[15px] font-semibold">Raccourcis clavier</h2>
    <div class="mt-3 grid grid-cols-2 gap-2 text-[13.5px]">
      <div class="flex items-center justify-between"><span>Rechercher</span><span class="kbd">Ctrl K</span></div>
      <div class="flex items-center justify-between"><span>Nouveau prêt</span><span class="kbd">Ctrl N</span></div>
      <div class="flex items-center justify-between"><span>Fermer une fenêtre ou une fiche</span><span class="kbd">Échap</span></div>
      <div class="flex items-center justify-between"><span>Inventaire : clé suivante</span><span class="kbd">Entrée</span></div>
    </div>
  </section>
</div>
