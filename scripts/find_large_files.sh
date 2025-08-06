#!/bin/bash

# Script pour trouver les fichiers de plus de 100MB sur Ubuntu
# Usage: ./find_large_files.sh [chemin] [taille]
# Exemple: ./find_large_files.sh /home 100M

# Configuration par défaut
DEFAULT_PATH="/"
DEFAULT_SIZE="100M"

# Paramètres
SEARCH_PATH="${1:-$DEFAULT_PATH}"
FILE_SIZE="${2:-$DEFAULT_SIZE}"

echo "Recherche des fichiers de plus de $FILE_SIZE dans $SEARCH_PATH..."
echo "Date: $(date)"
echo "----------------------------------------"

# Trouver les gros fichiers et les trier par taille
find "$SEARCH_PATH" -type f -size +"$FILE_SIZE" -exec ls -lh {} \; 2>/dev/null | \
    awk '{print $5, $NF}' | \
    sort -hr | \
    head -20

echo "----------------------------------------"
echo "Top 20 des plus gros fichiers trouvés"