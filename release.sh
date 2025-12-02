#!/bin/bash

echo "--- Script de Création de Release ---"
echo

# Check if a version tag is provided as an argument
if [ -z "$1" ]; then
  echo -e "\033[1;31mErreur :\033[0m Vous devez fournir un numéro de version en argument."
  echo "Exemple: ./release.sh v1.0.0"
  exit 1
fi

VERSION=$1
echo "Version à créer : $VERSION"
echo

# --- Vérification de la synchronisation avec le dépôt distant ---
echo "[0/3] Vérification de la synchronisation avec le dépôt distant..."
git remote update > /dev/null 2>&1
LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u})

if [ "$LOCAL" != "$REMOTE" ]; then
    echo -e "\033[1;31mErreur :\033[0m Votre branche locale n'est pas à jour avec le dépôt distant (origin)."
    echo "Veuillez utiliser 'git pull' pour récupérer les changements ou 'git push' pour envoyer les vôtres."
    exit 1
fi
echo "Votre branche est synchronisée. Poursuite..."
echo


# Prompt for a release message
echo "Veuillez entrer un court message pour décrire cette version (appuyez sur Entrée pour valider) :"
read -r RELEASE_MESSAGE

if [ -z "$RELEASE_MESSAGE" ]; then
    RELEASE_MESSAGE="Release $VERSION"
fi

echo
# Create the annotated git tag (maintenant étape 1/3)
echo "[1/3] Création du tag git '$VERSION'..."
git tag -a "$VERSION" -m "$RELEASE_MESSAGE"

# Check if the tag was created successfully
if [ $? -ne 0 ]; then
    echo -e "\033[1;31mErreur :\033[0m La création du tag a échoué. Assurez-vous que :"
    echo "- Vous n'utilisez pas un numéro de version qui existe déjà."
    echo "- Vous avez bien validé (commit) toutes vos modifications."
    exit 1
fi

# Push the tag to the remote repository (maintenant étape 2/3)
echo "[2/3] Poussée du tag '$VERSION' vers le dépôt distant (origin)..."
git push origin "$VERSION"

if [ $? -eq 0 ]; then
    # Try to get the GitHub repository URL automatically
    REPO_URL=$(git config --get remote.origin.url | sed 's/.*:\/\///;s/\.git$//')
    echo
    echo -e "\033[1;32mSuccès ! Le tag a été poussé sur GitHub.\033[0m"
    echo "L'action de compilation va démarrer automatiquement."
    if [ -n "$REPO_URL" ]; then
        echo "Vous pouvez suivre sa progression ici : https://github.com/$REPO_URL/actions"
    fi
else
    echo -e "\033[1;31mErreur :\033[0m La poussée du tag a échoué. Vérifiez votre connexion et vos droits d'accès au dépôt."
fi
