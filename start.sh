#!/bin/bash

echo "--- Lancement de l'application de gestion de clés ---"
echo

# Étape 1: Installation/Mise à jour des dépendances
echo "[INFO] Installation des dépendances Python depuis requirements.txt..."
python3 -m pip install -r requirements.txt --quiet --disable-pip-version-check
if [ $? -ne 0 ]; then
    echo "[ERREUR] L'installation des dépendances a échoué. Arrêt du script."
    exit 1
fi
echo "[OK] Les dépendances sont à jour."
echo

# Étape 2: Démarrage du serveur web en arrière-plan
echo "[INFO] Démarrage du serveur FastAPI sur le port 8000..."
python3 -m uvicorn app.main:app --port 8000 --log-level warning &
SERVER_PID=$!

# Intercepter le signal de sortie (Ctrl+C) pour arrêter le serveur proprement
trap 'echo; echo "[INFO] Arrêt du serveur..."; kill $SERVER_PID' SIGINT SIGTERM

# Laisser un moment au serveur pour démarrer
sleep 2

# Étape 3: Ouvrir la page dans le navigateur (spécifique à macOS, mais fonctionne aussi sur Linux avec xdg-utils)
echo "[INFO] Ouverture de http://127.0.0.1:8000 dans votre navigateur."
open http://127.0.0.1:8000

# Étape 4: Attendre que le processus du serveur se termine.
# Cela garde le script en cours d'exécution pour que le trap fonctionne.
echo
echo "[OK] L'application est lancée."
echo ">>> Appuyez sur Ctrl+C dans ce terminal pour arrêter le serveur. <<<"
wait $SERVER_PID

