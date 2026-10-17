# Makefile du Gestionnaire de Clés (V4 — Go + Wails, Windows x64)
#
# Prérequis (poste de développement uniquement) : Go, Node.js + npm, CLI Wails v2
#   go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0

.PHONY: help install-deps dev build package captures bindings test check lint clean

APP_NAME=clefs

help:
	@echo "Gestionnaire de Clés — commandes disponibles :"
	@echo "  make install-deps  Installer les dépendances Go et npm"
	@echo "  make dev           Lancer en développement (rechargement à chaud)"
	@echo "  make build         Compiler l'exécutable Windows (build/bin/clefs.exe)"
	@echo "  make package       Compiler + archive build/release/clefs-windows-amd64.zip"
	@echo "  make captures      Régénérer les captures du README (docs/images)"
	@echo "  make bindings      Régénérer les liaisons JavaScript des services Go"
	@echo "  make test          Lancer les tests Go"
	@echo "  make check         Vérifier les types de l'interface (svelte-check)"
	@echo "  make lint          gofmt + go vet + recherche de code mort"
	@echo "  make clean         Supprimer les fichiers compilés"

install-deps:
	go mod download
	cd frontend && npm install

dev:
	wails dev

# -webview2 error : si le moteur WebView2 manque, message clair, jamais
# d'installation proposée (règle « rien à installer sur les postes »).
# Pas de -clean : il effacerait build/bin, où se trouve la base de test locale.
build: bindings
	wails build -skipbindings -trimpath -webview2 error

# Archive de distribution identique à la release GitHub (exe + infos.txt).
package:
	powershell -NoProfile -ExecutionPolicy Bypass -File build.ps1 -Package

# Captures d'écran du README, avec des données fictives (Edge sans fenêtre).
captures:
	powershell -NoProfile -ExecutionPolicy Bypass -File docs/captures/capture.ps1

# Génère frontend/src/lib/wailsjs via `go run` : l'exécutable temporaire de
# Wails (wailsbindings.exe) peut être bloqué par Smart App Control.
bindings:
	go run -tags bindings . > /dev/null

test:
	go test ./internal/...

check:
	cd frontend && npx svelte-check --tsconfig ./tsconfig.json

lint:
	gofmt -l main.go main_bindings.go internal
	go vet ./...
	go run golang.org/x/tools/cmd/deadcode@v0.47.0 -test ./...

# Ne supprime que les fichiers compilés : build/bin peut contenir une base de test.
clean:
	rm -rf build/bin/clefs.exe build/release frontend/dist
	go clean -cache
