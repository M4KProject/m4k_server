#!/bin/bash

# Script pour build et run Node-RED
# Usage: ./nodered_build_run.sh [build|run|stop|restart]

NODERED_DIR="../nodered"
ACTION="${1:-run}"

echo "=== Node-RED Docker Management ==="
echo "Date: $(date)"
echo "Action: $ACTION"
echo "-----------------------------------"

# Vérifier si le répertoire Node-RED existe
if [ ! -d "$NODERED_DIR" ]; then
    echo "❌ Erreur: Le répertoire $NODERED_DIR n'existe pas"
    exit 1
fi

# Se déplacer dans le répertoire Node-RED
cd "$NODERED_DIR" || exit 1

# Créer le réseau web s'il n'existe pas
if ! docker network inspect web >/dev/null 2>&1; then
    echo "🌐 Création du réseau 'web'..."
    docker network create web
fi

case $ACTION in
    "build")
        echo "🔨 Build de l'image Node-RED..."
        docker-compose build
        echo "✅ Build terminé"
        ;;
    "run")
        echo "🚀 Démarrage de Node-RED..."
        docker-compose up -d
        echo "✅ Node-RED démarré"
        echo "🌐 Interface accessible sur: http://localhost:1880"
        ;;
    "stop")
        echo "🛑 Arrêt de Node-RED..."
        docker-compose down
        echo "✅ Node-RED arrêté"
        ;;
    "restart")
        echo "🔄 Redémarrage de Node-RED..."
        docker-compose down
        docker-compose up -d
        echo "✅ Node-RED redémarré"
        echo "🌐 Interface accessible sur: http://localhost:1880"
        ;;
    "logs")
        echo "📋 Affichage des logs..."
        docker-compose logs -f
        ;;
    *)
        echo "❌ Action non reconnue: $ACTION"
        echo "Actions disponibles: build, run, stop, restart, logs"
        exit 1
        ;;
esac

echo "✅ Script terminé"