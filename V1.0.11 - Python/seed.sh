#!/bin/bash

echo "--- Script de remplissage de la base de données ---"
echo
echo -e "\033[1;31mATTENTION :\033[0m Ce script va supprimer TOUTES les données actuellement dans la base"
echo "et les remplacer par un jeu de données de test."
echo
echo "Assurez-vous que l'application principale n'est PAS en cours d'exécution."
echo

read -p "Voulez-vous continuer ? (o/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Oo]$ ]]
then
    echo "Opération annulée."
    exit 1
fi

echo
echo "[INFO] Lancement du script de remplissage Python..."
python3 seed.py
echo
echo -e "\033[1;32m[SUCCÈS]\033[0m La base de données a été remplie."
echo "Vous pouvez maintenant démarrer l'application avec ./start.sh"
