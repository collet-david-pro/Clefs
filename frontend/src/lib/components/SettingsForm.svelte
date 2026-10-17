<script lang="ts">
  // Paramètres de l'établissement, des bons de remise, des documents et des
  // prêts. Enregistrés dans la base : communs à tous les postes.
  import { Building2, FileText, ImagePlus, KeyRound, RotateCcw, Save, Trash2 } from '@lucide/svelte'
  import { Api, loadAppSettings, notifyDocument, run, type Settings } from '../api'
  import { confirm, notify } from '../state.svelte'

  let form = $state<Settings | null>(null)
  let saved = $state('')
  let saving = $state(false)
  let sessionAgent = $state('')

  void load()
  async function load() {
    const s = await run(Api.GetSettings())
    if (s) {
      form = s
      saved = JSON.stringify(s)
      // Sans agent choisi, GetAppInfo donne le nom de la session Windows.
      if (!s.agentName) sessionAgent = (await run(Api.GetAppInfo()))?.agent ?? ''
    }
  }

  const dirty = $derived(form !== null && JSON.stringify(form) !== saved)

  // Aperçu du texte d'engagement avec un exemple.
  const engagementPreview = $derived(
    (form?.engagement ?? '')
      .replaceAll('{detenteur}', 'Dupont Jean')
      .replaceAll('{cles}', 'les 2 clés mentionnées ci-dessus')
      .replaceAll('{la_les}', 'les'),
  )

  async function chooseLogo() {
    const url = await run(Api.ChooseLogo())
    if (url && form) form.logoDataUrl = url
  }

  async function save() {
    if (!form) return
    saving = true
    const s = await run(Api.SaveSettings(form))
    saving = false
    if (s) {
      form = s
      saved = JSON.stringify(s)
      await loadAppSettings()
      notify('success', 'Paramètres enregistrés.')
    }
  }

  async function preview() {
    if (!form) return
    const path = await run(Api.PreviewReceipt(form))
    if (path) notifyDocument("Aperçu d'un bon de remise créé (non enregistré dans l'historique).", path)
  }

  async function restoreDefaults() {
    const ok = await confirm({
      title: "Rétablir les valeurs d'origine ?",
      message: "Tous les champs ci-dessous reprennent leur valeur d'origine (logo retiré). Rien n'est enregistré avant de cliquer sur « Enregistrer ».",
      confirmLabel: 'Rétablir',
    })
    if (ok) form = await Api.GetDefaultSettings()
  }

  function cancel() {
    form = JSON.parse(saved)
  }
</script>

{#if form}
  <div class="space-y-5 pb-20">
    <section class="card p-6">
      <h2 class="flex items-center gap-2 text-[15px] font-semibold"><Building2 size={17} class="text-primary" />Établissement</h2>
      <p class="mt-1 text-[13px] text-muted">Affiché dans le menu et en en-tête de tous les documents PDF.</p>
      <div class="mt-4 grid grid-cols-[1fr_200px] gap-6">
        <div class="space-y-4">
          <label class="block">
            <span class="label">Nom de l'établissement</span>
            <input class="input" bind:value={form.establishment} placeholder="ex. Collège Victor Hugo" />
          </label>
          <label class="block">
            <span class="label">Adresse ou ville</span>
            <input class="input" bind:value={form.address} placeholder="ex. 1 rue de l'École, 02300 Chauny" />
            <p class="hint">Facultatif : affichée à la suite du nom dans l'en-tête des documents.</p>
          </label>
          <label class="block">
            <span class="label">Nom de l'application</span>
            <input class="input" bind:value={form.appName} />
            <p class="hint">Titre de la fenêtre, du menu et des documents.</p>
          </label>
        </div>
        <div>
          <span class="label">Logo</span>
          <div class="flex h-28 items-center justify-center rounded-control border border-dashed border-line bg-surface-2 p-3">
            {#if form.logoDataUrl}
              <img src={form.logoDataUrl} alt="Logo de l'établissement" class="max-h-full max-w-full object-contain" />
            {:else}
              <div class="flex flex-col items-center gap-1.5 text-[12px] text-muted">
                <div class="flex size-9 items-center justify-center rounded-[10px] bg-primary text-primary-fg"><KeyRound size={18} /></div>
                Logo par défaut
              </div>
            {/if}
          </div>
          <div class="mt-2 flex gap-1.5">
            <button class="btn btn-secondary btn-sm flex-1" onclick={chooseLogo}><ImagePlus size={14} />Choisir…</button>
            {#if form.logoDataUrl}
              <button class="btn btn-danger-ghost btn-sm btn-icon" onclick={() => (form!.logoDataUrl = '')} aria-label="Retirer le logo"><Trash2 size={14} /></button>
            {/if}
          </div>
          <p class="hint">PNG ou JPEG, 1 Mo au plus.</p>
        </div>
      </div>
    </section>

    <section class="card p-6">
      <h2 class="flex items-center gap-2 text-[15px] font-semibold"><FileText size={17} class="text-primary" />Bons de remise</h2>
      <div class="mt-4 space-y-4">
        <label class="block">
          <span class="label">Titre du bon</span>
          <input class="input" bind:value={form.receiptTitle} />
        </label>
        <div>
          <div class="flex items-end justify-between">
            <span class="label">Texte d'engagement</span>
            <button class="mb-1.5 cursor-pointer text-[12px] font-medium text-primary hover:underline" onclick={async () => (form!.engagement = (await Api.GetDefaultSettings()).engagement)}>
              Texte d'origine
            </button>
          </div>
          <textarea class="input" rows="5" bind:value={form.engagement}></textarea>
          <p class="hint">
            Mots remplacés automatiquement :
            <code class="rounded bg-surface-2 px-1">{'{detenteur}'}</code> le nom du détenteur ·
            <code class="rounded bg-surface-2 px-1">{'{cles}'}</code> « la clé mentionnée ci-dessus » ou « les 3 clés mentionnées ci-dessus » ·
            <code class="rounded bg-surface-2 px-1">{'{la_les}'}</code> « la » ou « les ».
          </p>
          <div class="mt-2 rounded-control bg-surface-2 px-3 py-2 text-[12.5px]">
            <span class="font-medium text-muted">Exemple : </span>{engagementPreview}
          </div>
        </div>
        <div class="grid grid-cols-3 gap-3">
          <label>
            <span class="label">Cadre de signature du détenteur</span>
            <input class="input" bind:value={form.signatureBorrower} />
          </label>
          <label>
            <span class="label">Cadre de signature de l'établissement</span>
            <input class="input" bind:value={form.signatureEstablishment} />
          </label>
          <label>
            <span class="label">Signataire (facultatif)</span>
            <input class="input" bind:value={form.signatory} placeholder="ex. M. Dupont, gestionnaire" />
          </label>
        </div>
        <label class="block">
          <span class="label">Mention en pied de page (tous les documents)</span>
          <input class="input" bind:value={form.footerNote} placeholder="ex. Document à conserver — ne pas jeter" />
        </label>
      </div>
    </section>

    <section class="card p-6">
      <h2 class="flex items-center gap-2 text-[15px] font-semibold"><KeyRound size={17} class="text-primary" />Prêts</h2>
      <div class="mt-4 grid grid-cols-3 gap-4">
        <div>
          <span class="label">Type de prêt proposé</span>
          <div class="segmented flex">
            <button class="flex-1" aria-pressed={form.defaultLoanType !== 'permanent'} onclick={() => (form!.defaultLoanType = 'ponctuel')}>Ponctuel</button>
            <button class="flex-1" aria-pressed={form.defaultLoanType === 'permanent'} onclick={() => (form!.defaultLoanType = 'permanent')}>Permanent</button>
          </div>
        </div>
        <label>
          <span class="label">Retour prévu proposé (jours)</span>
          <input class="input" type="number" min="0" max="365" bind:value={form.defaultReturnDays} />
          <p class="hint">0 : pas de date proposée.</p>
        </label>
        <label>
          <span class="label">Nom de l'agent</span>
          <input class="input" bind:value={form.agentName} placeholder={sessionAgent ? `Session Windows : ${sessionAgent}` : 'Nom de la session Windows'} />
          <p class="hint">Inscrit sur les prêts et les bons. Vide : nom de la session Windows.</p>
        </label>
      </div>
    </section>
  </div>

  <!-- Barre d'enregistrement, toujours visible -->
  <div class="pointer-events-none sticky bottom-4 z-10 -mt-16 flex justify-center">
    <div class="pointer-events-auto flex items-center gap-2 rounded-card border border-line bg-surface px-4 py-2.5 shadow-pop">
      <span class="mr-2 text-[13px] {dirty ? 'font-medium text-warning' : 'text-muted'}">{dirty ? 'Modifications non enregistrées' : 'Paramètres à jour'}</span>
      <button class="btn btn-ghost btn-sm" onclick={restoreDefaults}><RotateCcw size={14} />Valeurs d'origine</button>
      <button class="btn btn-secondary btn-sm" onclick={preview}><FileText size={14} />Aperçu d'un bon</button>
      {#if dirty}<button class="btn btn-secondary btn-sm" onclick={cancel}>Annuler</button>{/if}
      <button class="btn btn-primary btn-sm" disabled={!dirty || saving} onclick={save}><Save size={14} />Enregistrer</button>
    </div>
  </div>
{/if}
