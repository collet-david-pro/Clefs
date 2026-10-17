# Régénère les captures d'écran de la documentation (docs/images/app-*.png)
# à partir de l'interface compilée et de données FICTIVES (mock.js), sans
# lancer l'application ni afficher de fenêtre : Edge est utilisé en mode
# headless, servi par un petit serveur local (serve.mjs).
#
#   .\docs\captures\capture.ps1
#
# Prérequis : Node.js (PATH ou %USERPROFILE%\dev-tools\node) et Microsoft Edge.
# Pour changer une scène (page affichée, actions simulées), voir la fin de mock.js.
$ErrorActionPreference = 'Stop'
$here = $PSScriptRoot
$repo = Resolve-Path (Join-Path $here '..\..')
$images = Join-Path $repo 'docs\images'

$nodeDir = Join-Path $env:USERPROFILE 'dev-tools\node'
if (-not (Get-Command node -ErrorAction SilentlyContinue) -and (Test-Path $nodeDir)) { $env:Path = "$nodeDir;$env:Path" }
$node = (Get-Command node).Source
$edge = @("${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe", "$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe") | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $edge) { throw 'Microsoft Edge introuvable.' }

Write-Host "Compilation de l'interface…" -ForegroundColor Cyan
Push-Location (Join-Path $repo 'frontend')
if (-not (Test-Path node_modules)) { npm install | Out-Null }
npm run build | Out-Null
Pop-Location

# Site de démonstration : interface compilée + maquette des services.
$work = Join-Path ([IO.Path]::GetTempPath()) 'clefs-captures'
Remove-Item $work -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force (Join-Path $work 'site') | Out-Null
Copy-Item (Join-Path $repo 'frontend\dist\*') (Join-Path $work 'site') -Recurse
Copy-Item (Join-Path $here 'mock.js') (Join-Path $work 'site\mock.js')
Copy-Item (Join-Path $here 'serve.mjs') (Join-Path $work 'serve.mjs')
$index = Join-Path $work 'site\index.html'
(Get-Content $index -Raw) -replace '<script type="module"', "<script src=`"/mock.js`"></script>`n    <script type=`"module`"" | Set-Content $index -Encoding utf8

# Scène (mock.js) → fichier de docs/images (tableau de paires : une clé
# « keys » masquerait la propriété Keys d'une table de hachage).
$shots = @(
  @('dashboard', 'app-tableau-de-bord'),
  @('dashboard-dark', 'app-tableau-de-bord-sombre'),
  @('new-loan', 'app-nouveau-pret'),
  @('keys', 'app-cles'),
  @('borrower', 'app-fiche-detenteur'),
  @('inventory', 'app-inventaire'),
  @('settings', 'app-parametres'),
  @('import', 'app-import'),
  @('key-plan', 'app-plan-de-cles')
)

$server = Start-Process -FilePath $node -ArgumentList "`"$(Join-Path $work 'serve.mjs')`"" -PassThru -WindowStyle Hidden
try {
  Start-Sleep -Seconds 1
  foreach ($pair in $shots) {
    $scene = $pair[0]
    $png = Join-Path $images ($pair[1] + '.png')
    Remove-Item $png -ErrorAction SilentlyContinue
    Start-Process -FilePath $edge -Wait -WindowStyle Hidden -ArgumentList @(
      '--headless=new', '--disable-gpu', '--hide-scrollbars', '--no-first-run', "--user-data-dir=$work\edge-profile",
      '--window-size=1360,860', '--force-device-scale-factor=1', '--virtual-time-budget=6000',
      "--screenshot=$png", "http://localhost:4729/?shot=$scene")
    Write-Host ("{0,-16} {1}" -f $scene, $(if (Test-Path $png) { 'ok' } else { 'ÉCHEC' }))
  }
} finally {
  Stop-Process -Id $server.Id -ErrorAction SilentlyContinue
}
Write-Host "Captures enregistrées dans $images" -ForegroundColor Green
