#!/bin/bash

# Script pour créer une nouvelle release du Gestionnaire de Clés
# Usage: ./create-release.sh [version]
# Exemple: ./create-release.sh 2.0.0

set -e  # Arrêter le script en cas d'erreur

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Gestionnaire de Clés - Create Release${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Vérifier si une version est fournie
if [ -z "$1" ]; then
    echo -e "${RED}❌ Erreur: Aucune version spécifiée${NC}"
    echo -e "${YELLOW}Usage: ./create-release.sh [version]${NC}"
    echo -e "${YELLOW}Exemple: ./create-release.sh 2.0.0${NC}"
    exit 1
fi

VERSION=$1
TAG="v${VERSION}"

echo -e "${BLUE}📦 Version à créer: ${GREEN}${TAG}${NC}"
echo ""

# Vérifier si le tag existe déjà
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo -e "${RED}❌ Erreur: Le tag ${TAG} existe déjà${NC}"
    echo -e "${YELLOW}💡 Conseil: Utilisez une version différente ou supprimez le tag existant avec:${NC}"
    echo -e "${YELLOW}   git tag -d ${TAG}${NC}"
    echo -e "${YELLOW}   git push origin :refs/tags/${TAG}${NC}"
    exit 1
fi

# Vérifier si le répertoire est un dépôt git
if [ ! -d .git ]; then
    echo -e "${RED}❌ Erreur: Ce répertoire n'est pas un dépôt git${NC}"
    exit 1
fi

# Vérifier s'il y a des modifications non commitées
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}⚠️  Attention: Il y a des modifications non commitées${NC}"
    echo -e "${YELLOW}Voulez-vous continuer quand même? (y/n)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo -e "${RED}❌ Opération annulée${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}📝 Étapes à effectuer:${NC}"
echo -e "  1. Créer le tag ${GREEN}${TAG}${NC}"
echo -e "  2. Pousser le tag vers GitHub"
echo -e "  3. GitHub Actions va automatiquement:"
echo -e "     - Builder l'application pour Windows x64"
echo -e "     - Builder l'application pour macOS Intel (amd64)"
echo -e "     - Builder l'application pour macOS Apple Silicon (arm64)"
echo -e "     - Créer une release avec tous les fichiers .zip"
echo ""
echo -e "${YELLOW}Voulez-vous continuer? (y/n)${NC}"
read -r response

if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo -e "${RED}❌ Opération annulée${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}🏷️  Création du tag ${TAG}...${NC}"

# Demander un message pour le tag
echo -e "${YELLOW}Entrez un message pour cette release (ou appuyez sur Entrée pour un message par défaut):${NC}"
read -r tag_message

if [ -z "$tag_message" ]; then
    tag_message="Release ${TAG}"
fi

# Créer le tag annoté
git tag -a "$TAG" -m "$tag_message"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Tag ${TAG} créé avec succès${NC}"
else
    echo -e "${RED}❌ Erreur lors de la création du tag${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}🚀 Push du tag vers GitHub...${NC}"

# Pousser le tag
git push origin "$TAG"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Tag poussé avec succès vers GitHub${NC}"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✨ Release ${TAG} créée avec succès!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}📊 Prochaines étapes:${NC}"
    echo -e "  1. GitHub Actions va automatiquement builder l'application pour:"
    echo -e "     • Windows x64 (compatible x86)"
    echo -e "     • macOS Intel (amd64)"
    echo -e "     • macOS Apple Silicon (arm64)"
    echo -e "  2. Surveillez la progression sur: ${YELLOW}https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions${NC}"
    echo -e "  3. Une fois terminé, la release sera disponible sur: ${YELLOW}https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/releases${NC}"
    echo ""
    echo -e "${BLUE}⏱️  Le build prend généralement 10-15 minutes (3 plateformes)${NC}"
else
    echo -e "${RED}❌ Erreur lors du push du tag${NC}"
    echo -e "${YELLOW}💡 Le tag a été créé localement mais n'a pas pu être poussé${NC}"
    echo -e "${YELLOW}   Vous pouvez réessayer avec: git push origin ${TAG}${NC}"
    exit 1
fi
