# Compilation locale du Gestionnaire de Clés sous Windows (sans make).
#
#   .\build.ps1            liaisons JS + exécutable build\bin\clefs.exe
#   .\build.ps1 -Test      lance aussi les tests Go et la vérification de l'interface
#   .\build.ps1 -Package   crée aussi build\release\clefs-windows-amd64.zip
#                          (même contenu que la release GitHub : exe + infos.txt)
#   .\build.ps1 -Run       lance l'application après compilation
#
# Les options se combinent : .\build.ps1 -Test -Package
# Utilise Go, Node.js et Wails du PATH, ou à défaut une installation portable
# dans %USERPROFILE%\dev-tools (go\, node\) et %USERPROFILE%\go\bin (wails).
param([switch]$Test, [switch]$Package, [switch]$Run)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$portable = Join-Path $env:USERPROFILE 'dev-tools'
foreach ($dir in @((Join-Path $portable 'go\bin'), (Join-Path $portable 'node'), (Join-Path $env:USERPROFILE 'go\bin'))) {
    if ((Test-Path $dir) -and ($env:Path -notlike "*$dir*")) { $env:Path = "$dir;$env:Path" }
}

foreach ($tool in 'go', 'npm', 'wails') {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Outil introuvable : $tool (voir NOTICE.md § 9)" }
}

$version = (Select-String -Path 'internal\app\app.go' -Pattern '^const Version = "(.*)"').Matches[0].Groups[1].Value
Write-Host "Gestionnaire de Clés $version" -ForegroundColor Cyan

Write-Host '1/3 Liaisons JavaScript des services Go' -ForegroundColor Cyan
# main.go intègre frontend/dist : sur un dépôt fraîchement cloné, un fichier
# vide permet de compiler le paquet principal avant la construction de l'interface.
if (-not (Test-Path 'frontend\dist')) { New-Item -ItemType File -Force 'frontend\dist\.gitkeep' | Out-Null }
# Via `go run` : l'exécutable temporaire de Wails (wailsbindings.exe) peut être
# bloqué par Smart App Control.
go run -tags bindings . | Out-Null
if ($LASTEXITCODE -ne 0) {
    # Smart App Control bloque parfois aussi `go run` : on garde alors les
    # liaisons existantes (commitées). Après l'ajout d'une méthode Go, relancer
    # jusqu'à ce que la génération passe ; -Test détecte des liaisons périmées.
    Write-Warning 'Liaisons non régénérées (exécution bloquée ?) : les liaisons existantes sont utilisées.'
}

if ($Test) {
    Write-Host '   Tests Go' -ForegroundColor Cyan
    go test ./internal/...
    if ($LASTEXITCODE -ne 0) { throw 'Tests Go en échec' }
    Push-Location frontend
    if (-not (Test-Path node_modules)) { npm install }
    npx svelte-check --tsconfig ./tsconfig.json
    $ok = $LASTEXITCODE -eq 0
    Pop-Location
    if (-not $ok) { throw "Vérification de l'interface en échec" }
}

Write-Host '2/3 Compilation (interface + exécutable)' -ForegroundColor Cyan
# Pas de -clean : il effacerait build\bin, où se trouve la base de test locale.
wails build -skipbindings -trimpath -webview2 error
if ($LASTEXITCODE -ne 0) { throw "Échec de la compilation (l'application est-elle ouverte ?)" }
$exe = Join-Path $PSScriptRoot 'build\bin\clefs.exe'

if ($Package) {
    # Même archive que la CI (.github/workflows/release.yml).
    $release = Join-Path $PSScriptRoot 'build\release'
    $stage = Join-Path $release 'contenu'
    Remove-Item $release -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force $stage | Out-Null
    Copy-Item $exe (Join-Path $stage 'clefs-windows-amd64.exe')
    Copy-Item 'infos.txt' $stage
    $zip = Join-Path $release 'clefs-windows-amd64.zip'
    Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $zip
    Remove-Item $stage -Recurse -Force
    Write-Host "   Archive : build\release\clefs-windows-amd64.zip (version $version)" -ForegroundColor Cyan
}

Write-Host '3/3 Terminé : build\bin\clefs.exe' -ForegroundColor Green
if ($Run) { Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) }
