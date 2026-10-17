// Maquette des services Go (window.go.app.App) avec des données FICTIVES,
// pour produire les captures de docs/images sans lancer l'application
// (voir docs/captures/capture.ps1).
(function () {
  const shot = new URLSearchParams(location.search).get('shot') || 'dashboard'
  try {
    localStorage.setItem('theme', shot.endsWith('-dark') ? 'dark' : 'light')
  } catch (e) {}

  const day = 86400000
  const now = Date.now()
  const iso = (d) => new Date(now - d * day).toISOString()

  const buildings = [
    { id: 1, name: 'Bâtiment A' },
    { id: 2, name: 'Bâtiment B — Sciences' },
    { id: 3, name: 'Gymnase' },
    { id: 4, name: 'Extérieur' },
  ]
  const accessDefs = [
    [1, 'Salle 101', 1, 'RDC', 'Salle de classe'], [2, 'Salle 102', 1, 'RDC', 'Salle de classe'],
    [3, 'Salle 103', 1, 'R+1', 'Salle de classe'], [4, 'Salle 104', 1, 'R+1', 'Salle de classe'],
    [5, 'CDI', 1, 'RDC', 'Documentation'], [6, 'Salle des professeurs', 1, 'RDC', 'Bureau'],
    [7, 'Bureau de la direction', 1, 'R+1', 'Bureau'], [8, 'Local reprographie', 1, 'RDC', 'Local technique'],
    [9, 'Laboratoire 1', 2, 'RDC', 'Laboratoire'], [10, 'Laboratoire 2', 2, 'RDC', 'Laboratoire'],
    [11, 'Réserve sciences', 2, 'RDC', 'Stockage'], [12, 'Salle 201', 2, 'R+1', 'Salle de classe'],
    [13, 'Salle 202', 2, 'R+1', 'Salle de classe'], [14, 'Gymnase', 3, 'RDC', 'Sport'],
    [15, 'Vestiaires', 3, 'RDC', 'Sport'], [16, 'Local matériel EPS', 3, 'RDC', 'Stockage'],
    [17, 'Portail principal', 4, '', 'Extérieur'], [18, 'Portail parking', 4, '', 'Extérieur'],
  ]
  const keyDefs = [
    // id, numéro, description, total, réserve, rangement, catégorie, accès
    [1, 'PG', 'Passe général', 3, 1, 'Coffre de la direction', 'passe', [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]],
    [2, 'PA', 'Passe du bâtiment A', 4, 1, 'Armoire à clés, crochet 1', 'passe', [1, 2, 3, 4, 5, 6, 8]],
    [3, 'K-101', 'Salle 101', 3, 1, 'Armoire à clés, crochet 4', 'simple', [1]],
    [4, 'K-102', 'Salle 102', 3, 1, 'Armoire à clés, crochet 5', 'simple', [2]],
    [5, 'K-103', 'Salle 103', 2, 0, 'Armoire à clés, crochet 6', 'simple', [3]],
    [6, 'K-104', 'Salle 104', 2, 0, 'Armoire à clés, crochet 7', 'simple', [4]],
    [7, 'CDI', 'Centre de documentation', 2, 1, 'Armoire à clés, crochet 8', 'simple', [5]],
    [8, 'SC', 'Laboratoires et réserve de sciences', 4, 1, 'Armoire à clés, crochet 10', 'trousseau', [9, 10, 11]],
    [9, 'K-201', 'Salle 201', 2, 0, 'Armoire à clés, crochet 11', 'simple', [12]],
    [10, 'K-202', 'Salle 202', 2, 0, 'Armoire à clés, crochet 12', 'simple', [13]],
    [11, 'GYM', 'Gymnase et vestiaires', 2, 0, 'Loge', 'trousseau', [14, 15]],
    [12, 'EPS', 'Local matériel EPS', 2, 0, 'Loge', 'simple', [16]],
    [13, 'POR', 'Portails', 4, 1, 'Loge', 'badge', [17, 18]],
    [14, 'REP', 'Local reprographie', 2, 0, 'Armoire à clés, crochet 2', 'simple', [8]],
  ]
  const borrowerDefs = [
    [1, 'Martin Sophie', 'permanent', 'sophie.martin@exemple.fr', '03 23 00 00 01'],
    [2, 'Bernard Lucas', 'permanent', 'lucas.bernard@exemple.fr', ''],
    [3, 'Petit Camille', 'contractuel', 'camille.petit@exemple.fr', ''],
    [4, 'Durand Hugo', 'permanent', 'hugo.durand@exemple.fr', ''],
    [5, 'Leroy Emma', 'intervenant', 'emma.leroy@exemple.fr', ''],
    [6, 'Moreau Nathan', 'permanent', '', ''],
    [7, 'Nettoyage Plus (société)', 'entreprise', 'contact@nettoyage-plus.exemple', '03 23 00 00 99'],
    [8, 'Fournier Léa', 'contractuel', 'lea.fournier@exemple.fr', ''],
    [9, 'Girard Paul', 'permanent', '', ''],
    [10, 'Lambert Chloé', 'permanent', 'chloe.lambert@exemple.fr', ''],
  ]
  const loanDefs = [
    // id, clé, détenteur, il y a (jours), retour prévu (jours, négatif = passé), type
    [1, 8, 1, 120, null, 'permanent'], [2, 3, 1, 120, null, 'permanent'], [3, 7, 1, 40, null, 'permanent'],
    [4, 2, 2, 200, null, 'permanent'], [5, 11, 2, 200, null, 'permanent'],
    [6, 4, 3, 30, 10, 'ponctuel'], [7, 9, 4, 15, null, 'permanent'], [8, 11, 5, 9, -2, 'ponctuel'],
    [9, 13, 7, 300, null, 'permanent'], [10, 14, 7, 300, null, 'permanent'], [11, 1, 7, 300, null, 'permanent'],
    [12, 5, 8, 6, 20, 'ponctuel'], [13, 11, 9, 4, null, 'ponctuel'], [14, 12, 9, 4, null, 'ponctuel'],
    [15, 10, 10, 2, null, 'permanent'], [16, 8, 10, 2, null, 'permanent'], [17, 1, 4, 50, null, 'permanent'],
  ]

  const bName = Object.fromEntries(buildings.map((b) => [b.id, b.name]))
  const accesses = accessDefs.map(([id, name, b, floor, category]) => ({
    id, name, type: '', buildingId: b, buildingName: bName[b], floor, category, notes: '',
    keyIds: keyDefs.filter((k) => k[7].includes(id)).map((k) => k[0]),
  }))
  const borrowers0 = borrowerDefs.map(([id, name, status, email, phone]) => ({ id, name, status, email, phone }))
  const loans = loanDefs.map(([id, keyId, bId, ago, planned, loanType]) => {
    const k = keyDefs.find((x) => x[0] === keyId)
    const b = borrowers0.find((x) => x.id === bId)
    const plannedDate = planned === null ? '' : new Date(now + planned * day).toISOString()
    return {
      id, keyId, keyNumber: k[1], keyDescription: k[2], borrowerId: bId, borrowerName: b.name,
      loanDate: iso(ago), returnDate: '', plannedReturnDate: plannedDate, loanType, condition: '',
      createdBy: 'secretariat', status: planned !== null && planned < 0 ? 'overdue' : 'active', days: ago,
    }
  })
  const keys = keyDefs.map(([id, number, description, total, reserve, storageLocation, category, accessIds]) => {
    const kl = loans.filter((l) => l.keyId === id)
    return {
      id, number, description, category, storageLocation, notes: '', total, reserve,
      loaned: kl.length, available: total - reserve - kl.length, holders: kl.map((l) => l.borrowerName), accessIds,
    }
  })
  const borrowers = borrowers0.map((b) => ({ ...b, activeLoans: loans.filter((l) => l.borrowerId === b.id).length }))
  const history = [
    ...loans,
    { ...loans[5], id: 90, keyId: 6, keyNumber: 'K-104', keyDescription: 'Salle 104', loanDate: iso(60), returnDate: iso(50), status: 'returned', days: 10, condition: 'Bon état' },
  ].sort((a, b) => (a.loanDate < b.loanDate ? 1 : -1))

  const settings = {
    appName: 'Gestionnaire de Clés', establishment: 'Collège Victor Hugo', address: 'Chauny', logoDataUrl: '',
    footerNote: 'Document à conserver', receiptTitle: 'Bon de remise de clés',
    engagement: "Je soussigné(e), {detenteur}, reconnais avoir reçu {cles}. Je m'engage à en prendre soin et à {la_les} restituer à la fin de leur utilisation ou à première demande, et à signaler sans délai toute perte. En cas de perte ou de dégradation, ma responsabilité pourra être engagée.",
    signatureBorrower: 'Le détenteur', signatureEstablishment: "Pour l'établissement", signatory: 'Le gestionnaire',
    defaultLoanType: 'ponctuel', defaultReturnDays: 0, agentName: 'Secrétariat',
  }

  const accessById = (id) => accesses.find((a) => a.id === id)
  const keyById = (id) => keys.find((k) => k.id === id)
  const anomalies = keys.filter((k) => k.available < 0).map((k) => ({ keyId: k.id, keyNumber: k.number, total: k.total, reserve: k.reserve, loaned: k.loaned, available: k.available }))

  const api = {
    GetAppInfo: () => ({ version: '4.0.0', dataDir: 'C:\\Clefs', dbPath: 'C:\\Clefs\\clefs.db', documentsDir: 'C:\\Clefs\\documents', backupsDir: 'C:\\Clefs\\backups', agent: 'Secrétariat', isEmpty: false }),
    GetSettings: () => settings,
    GetDefaultSettings: () => settings,
    GetKeyCategories: () => ['simple', 'trousseau', 'badge', 'passe'],
    GetBorrowerStatuses: () => ['permanent', 'contractuel', 'intervenant', 'entreprise'],
    ListKeys: () => keys,
    ListAccesses: () => accesses,
    ListBorrowers: () => borrowers,
    ListBuildings: () => buildings.map((b) => ({ ...b, accessCount: accesses.filter((a) => a.buildingId === b.id).length })),
    ListActiveLoans: () => loans,
    GetHistory: () => history,
    ListBackups: () => [0, 1, 7].map((d, i) => ({ path: `C:\\Clefs\\backups\\clefs_backup_${i}.db`, name: `clefs_backup_2026101${7 - i}_1730${i}0.db`, size: '96.0 KB', date: iso(d) })),
    GetDashboard: () => ({
      keyTypes: keys.length, keyCopies: keys.reduce((n, k) => n + k.total, 0), loanedCopies: loans.length,
      availableCopies: keys.reduce((n, k) => n + Math.max(k.available, 0), 0), borrowers: borrowers.length,
      borrowersWithKeys: borrowers.filter((b) => b.activeLoans).length, accesses: accesses.length,
      overdue: loans.filter((l) => l.status === 'overdue'), anomalies,
      redundancies: [{ borrowerId: 7, borrowerName: 'Nettoyage Plus (société)', keyNumbers: ['PG', 'REP'], accesses: ['Local reprographie'] }],
      recentLoans: history.slice(0, 8),
    }),
    GetBorrowerDetail: (id) => {
      const b = borrowers.find((x) => x.id === id)
      const bl = loans.filter((l) => l.borrowerId === id)
      const covered = [...new Set(bl.flatMap((l) => keyById(l.keyId).accessIds))].map(accessById)
      return { borrower: b, activeLoans: bl, coveredAccesses: covered, redundancies: [], history: history.filter((l) => l.borrowerId === id) }
    },
    GetKeyDetail: (id) => {
      const k = keyById(id)
      return { key: k, accesses: k.accessIds.map(accessById), activeLoans: loans.filter((l) => l.keyId === id), history: history.filter((l) => l.keyId === id) }
    },
    SuggestKeys: (req) => {
      const wanted = new Set(req.accessIds)
      const candidates = keys.filter((k) => k.available > 0 && !req.excludedKeyIds.includes(k.id) && k.accessIds.some((a) => wanted.has(a)))
      const remaining = new Set(wanted)
      const chosen = []
      while (remaining.size) {
        let best = null, score = 0
        for (const k of candidates) {
          const s = k.accessIds.filter((a) => remaining.has(a)).length
          const bonus = k.accessIds.length - s
          if (s > score || (s === score && best && bonus < best.accessIds.length - score)) { best = k; score = s }
        }
        if (!best) break
        chosen.push(best)
        best.accessIds.forEach((a) => remaining.delete(a))
      }
      return {
        keys: chosen.map((k) => ({ key: k, suggested: true, manual: false, covers: k.accessIds.filter((a) => wanted.has(a)).map((a) => accessById(a).name) })),
        uncoverable: [...remaining].map(accessById), redundancies: [],
      }
    },
    GetKeyPlan: () => buildings.map((b) => ({ id: b.id, name: b.name, accesses: accesses.filter((a) => a.buildingId === b.id).map((a) => ({ access: a, keys: a.keyIds.map(keyById) })) })),
    GetWhoHasWhat: () => borrowers.filter((b) => b.activeLoans).map((b) => ({ borrower: b, loans: loans.filter((l) => l.borrowerId === b.id), accesses: [], redundancies: [] })),
    GetRedundancies: () => [],
    ChooseDatabaseFile: () => 'C:\\Clefs-V3\\clefs.db',
    InspectDatabaseFile: (path) => ({ path, kind: 'v3', label: 'Version 3 (base ou sauvegarde 3.x)', buildings: 7, accesses: 96, keys: 53, borrowers: 62, loans: 162, activeLoans: 153 }),
  }
  const App = new Proxy(api, {
    get: (target, name) => async (...args) => (name in target ? structuredClone(target[name](...args)) : null),
  })
  window.go = { app: { App } }

  // --- Mise en scène de chaque capture ---
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
  const byText = (sel, text) => [...document.querySelectorAll(sel)].find((e) => e.textContent.trim().includes(text))
  const nav = (label) => byText('aside button', label)?.click()
  const typeIn = (input, value) => {
    input.value = value
    input.dispatchEvent(new Event('input', { bubbles: true }))
  }

  const scenes = {
    'dashboard': async () => {},
    'dashboard-dark': async () => {},
    'new-loan': async () => {
      nav('Nouveau prêt')
      await sleep(400)
      const input = document.querySelector('input[placeholder="Tapez un nom…"]')
      input.dispatchEvent(new FocusEvent('focus'))
      await sleep(200)
      byText('ul button', 'Lambert Chloé')?.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
      await sleep(300)
      for (const name of ['Laboratoire 1', 'Laboratoire 2', 'Salle 201', 'Portail principal']) {
        byText('label', name)?.querySelector('input')?.click()
        await sleep(150)
      }
      await sleep(500)
    },
    'keys': async () => {
      nav('Clés')
    },
    'borrower': async () => {
      nav('Détenteurs')
      await sleep(400)
      byText('tbody tr', 'Martin Sophie')?.click()
    },
    'inventory': async () => {
      nav('Inventaire')
      await sleep(400)
      const counts = { 'PG': 2, 'PA': 3, 'K-101': 3, 'K-102': 1, 'K-103': 2, 'GYM': 0 }
      for (const [num, v] of Object.entries(counts)) {
        const input = document.querySelector(`input[aria-label^="Exemplaires de ${num} "]`)
        if (input) typeIn(input, String(v))
        await sleep(60)
      }
    },
    'settings': async () => {
      nav('Paramètres')
    },
    'import': async () => {
      nav('Importer une base')
      await sleep(300)
      byText('main button', 'Continuer')?.click()
      await sleep(300)
      byText('main button', 'Parcourir')?.click()
    },
    'key-plan': async () => {
      nav('Plan de clés')
    },
  }

  window.addEventListener('load', async () => {
    await sleep(700)
    await (scenes[shot.replace(/-dark$/, '')] || scenes[shot] || (async () => {}))()
  })
})()
