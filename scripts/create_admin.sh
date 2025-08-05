#!/bin/bash

# Script pour créer l'admin PocketBase

if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "Erreur: ADMIN_EMAIL ou ADMIN_PASSWORD non défini dans .env"
    exit 1
fi

echo "Création de l'admin PocketBase..."
echo "Email: $ADMIN_EMAIL"

# Vérifier si le conteneur PocketBase est en cours d'exécution
if ! docker ps | grep -q pocketbase; then
    echo "Erreur: le conteneur PocketBase n'est pas en cours d'exécution"
    echo "Lancez d'abord: ./docker_build.sh"
    exit 1
fi

# Créer le superuser
docker exec pocketbase /app/pocketbase superuser create "$ADMIN_EMAIL" "$ADMIN_PASSWORD"

if [ $? -eq 0 ]; then
    echo "Admin créé avec succès!"
    echo "Vous pouvez maintenant vous connecter à http://localhost:8090/_/"
else
    echo "Erreur lors de la création de l'admin (il existe peut-être déjà)"
fi