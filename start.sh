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
PORT=${1:-8000}
echo "[INFO] Démarrage du serveur FastAPI sur le port ${PORT}..."

# Vérifier si le port est déjà utilisé
BUSY_PIDS=$(lsof -ti tcp:${PORT} || true)
if [ -n "$BUSY_PIDS" ]; then
    echo "[WARN] Le port ${PORT} est déjà utilisé par le(s) PID: $BUSY_PIDS"
    read -p "Voulez-vous terminer ce(s) processus(s) et libérer le port ${PORT} ? [y/N] " yn
    case "$yn" in
        [Yy]*)
            echo "[INFO] Arrêt des processus: $BUSY_PIDS";
            kill $BUSY_PIDS || { echo "[ERREUR] Impossible de tuer $BUSY_PIDS"; exit 1; }
            sleep 1
            ;;
        *)
            echo "[INFO] Choisissez un autre port en lançant: sh start.sh <port>";
            exit 1
            ;;
    esac
fi

python3 -m uvicorn app.main:app --port ${PORT} --log-level warning &
SERVER_PID=$!

# Intercepter le signal de sortie (Ctrl+C) pour arrêter le serveur proprement
trap 'echo; echo "[INFO] Arrêt du serveur..."; kill $SERVER_PID' SIGINT SIGTERM

# Laisser un moment au serveur pour démarrer
sleep 2

# Vérifier que le serveur a bien démarré et écoute; si non, afficher une erreur
LISTENING=$(lsof -i tcp:${PORT} -sTCP:LISTEN -t || true)
if [ -z "$LISTENING" ]; then
    echo "[ERREUR] Le serveur n'a pas démarré correctement ou le port ${PORT} n'est pas à l'écoute. Vérifiez les logs du serveur.";
    # Remonter le code de sortie du serveur s'il est terminé
    wait $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Étape 3: Ouvrir la page dans le navigateur (spécifique à macOS, mais fonctionne aussi sur Linux avec xdg-utils)
URL="http://127.0.0.1:${PORT}"
echo "[INFO] Ouverture de ${URL} dans votre navigateur."
open ${URL}

# Étape 4: Attendre que le processus du serveur se termine.
# Cela garde le script en cours d'exécution pour que le trap fonctionne.
echo
echo "[OK] L'application est lancée."
echo ">>> Appuyez sur Ctrl+C dans ce terminal pour arrêter le serveur. <<<"
wait $SERVER_PID

