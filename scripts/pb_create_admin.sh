#!/bin/bash

# Script pour créer l'admin PocketBase

# Charger les variables d'environnement depuis .env
if [ -f "../.env" ]; then
    export $(grep -v '^#' ../.env | xargs)
    echo "✅ Variables d'environnement chargées depuis .env"
elif [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo "✅ Variables d'environnement chargées depuis .env"
fi

if [ -z "$ADMIN_EMAIL" ] || [ -z "$ADMIN_PASSWORD" ]; then
    echo "❌ Erreur: ADMIN_EMAIL ou ADMIN_PASSWORD non défini"
    echo "Vérifiez que le fichier .env existe et contient ces variables"
    exit 1
fi

echo "Configuration de l'admin PocketBase..."
echo "Email: $ADMIN_EMAIL"

# Vérifier si le conteneur PocketBase est en cours d'exécution
if ! docker ps | grep -q pocketbase; then
    echo "Erreur: le conteneur PocketBase n'est pas en cours d'exécution"
    echo "Lancez d'abord: ./docker_build.sh"
    exit 1
fi

# Créer ou mettre à jour le superuser avec upsert
echo "Configuration de l'admin (création ou mise à jour)..."
UPSERT_OUTPUT=$(docker exec pocketbase /app/pocketbase superuser upsert "$ADMIN_EMAIL" "$ADMIN_PASSWORD" 2>&1)
UPSERT_RESULT=$?

if [ $UPSERT_RESULT -eq 0 ]; then
    echo "✅ Admin configuré avec succès!"
    echo "Email: $ADMIN_EMAIL"
    
    # Tester l'authentification API pour vérifier que les credentials fonctionnent
    echo "🔍 Test de l'authentification API..."
    AUTH_TEST=$(curl -s -X POST "http://localhost:8090/api/collections/_superusers/auth-with-password" \
        -H "Content-Type: application/json" \
        -d "{\"identity\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")
    
    if echo "$AUTH_TEST" | grep -q '"token"'; then
        echo "✅ Authentification API réussie!"
        echo "Vous pouvez maintenant vous connecter à http://localhost:8090/_/"
    else
        echo "⚠️ Problème d'authentification détecté, nettoyage et recréation..."
        docker exec pocketbase /app/pocketbase superuser delete "$ADMIN_EMAIL" 2>/dev/null
        docker exec pocketbase /app/pocketbase superuser create "$ADMIN_EMAIL" "$ADMIN_PASSWORD"
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