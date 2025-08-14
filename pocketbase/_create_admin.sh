#!/bin/bash

# Script pour créer l'admin PocketBase

cp ../.env .env
source .env

if [ -z "$PB_ADMIN_USERNAME" ] || [ -z "$PB_ADMIN_PASSWORD" ]; then
    echo "❌ Erreur: PB_ADMIN_USERNAME ou PB_ADMIN_PASSWORD non défini"
    echo "Vérifiez que le fichier .env existe et contient ces variables"
    exit 1
fi

# Vérifier si le conteneur PocketBase est en cours d'exécution
if ! docker ps | grep -q pocketbase; then
    echo "Erreur: le conteneur PocketBase n'est pas en cours d'exécution"
    echo "Lancez d'abord: ./_build.sh"
    exit 1
fi

# Créer ou mettre à jour le superuser avec upsert
UPSERT_OUTPUT=$(docker exec pocketbase /app/pocketbase superuser upsert "$PB_ADMIN_USERNAME" "$PB_ADMIN_PASSWORD" 2>&1)
UPSERT_RESULT=$?

if [ $UPSERT_RESULT -eq 0 ]; then
    echo "✅ Admin configuré avec succès!"
    echo "Email: $PB_ADMIN_EMAIL"
    
    # Tester l'authentification API pour vérifier que les credentials fonctionnent
    echo "🔍 Test de l'authentification API..."
    AUTH_TEST=$(curl -s -X POST "http://localhost:8090/api/collections/_superusers/auth-with-password" \
        -H "Content-Type: application/json" \
        -d "{\"identity\":\"$PB_ADMIN_EMAIL\",\"password\":\"$PB_ADMIN_PASSWORD\"}")
    
    if echo "$AUTH_TEST" | grep -q '"token"'; then
        echo "✅ Authentification API réussie!"
        echo "Vous pouvez maintenant vous connecter à http://localhost:8090/_/"
    else
        echo "⚠️ Problème d'authentification détecté, nettoyage et recréation..."
        docker exec pocketbase /app/pocketbase superuser delete "$PB_ADMIN_EMAIL" 2>/dev/null
        docker exec pocketbase /app/pocketbase superuser create "$PB_ADMIN_EMAIL" "$PB_ADMIN_PASSWORD"
        echo "✅ Admin recréé avec succès!"
    fi
    
    echo ""
    echo "💡 Si vous obtenez 'Invalid login credentials' sur l'interface web,"
    echo "   essayez de vider le cache de votre navigateur ou utilisez un onglet privé."
else
    echo "❌ Erreur lors de la configuration de l'admin:"
    echo "$UPSERT_OUTPUT"
    exit 1
fi