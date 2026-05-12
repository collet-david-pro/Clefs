#!/bin/bash

# Script de démarrage pour Gestionnaire de Clés

echo "🔑 Gestionnaire de Clés - Démarrage"
echo "=================================="
echo ""

# Vérifier si Go est installé
if ! command -v go &> /dev/null
then
    echo "❌ Go n'est pas installé sur ce système"
    echo ""
    echo "Pour installer Go:"
    echo "  macOS:   brew install go"
    echo "  Linux:   sudo apt install golang-go"
    echo "  Windows: Téléchargez depuis https://go.dev/dl/"
    echo ""
    exit 1
fi

echo "✓ Go est installé ($(go version))"
echo ""

# Vérifier si les dépendances sont installées
if [ ! -f "go.sum" ]; then
    echo "📦 Installation des dépendances..."
    go mod download
    go mod tidy
    echo "✓ Dépendances installées"
    echo ""
fi

# Lancer l'application
echo "🚀 Lancement de l'application..."
echo ""
go run ./cmd/main.go
