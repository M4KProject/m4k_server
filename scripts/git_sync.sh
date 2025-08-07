#!/bin/bash

# Script pour git pull avec vérifications
# Usage: ./git_pull.sh

cd ../

echo "=== Git Pull Script ==="
echo "Date: $(date)"
echo "Répertoire: $(pwd)"
echo "------------------------"

# Vérifier si on est dans un repo git
if [ ! -d ".git" ]; then
    echo "❌ Erreur: Ce répertoire n'est pas un dépôt git"
    exit 1
fi

# Afficher le statut actuel
echo "📊 Statut avant pull:"
git status --short

# Afficher la branche courante
echo "🌿 Branche courante: $(git branch --show-current)"

# Vérifier s'il y a des modifications non commitées
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  Attention: Des modifications non commitées sont présentes"
    read -p "Continuer le pull ? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Pull annulé"
        exit 1
    fi
fi

# Effectuer le pull
echo "⬇️  Exécution de git pull..."
if git pull; then
    echo "✅ Pull réussi"
else
    echo "❌ Erreur lors du pull"
    exit 1
fi

# Afficher le statut final
echo "📊 Statut après pull:"
git status --short

echo "✅ Script terminé avec succès"