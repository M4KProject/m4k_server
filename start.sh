#!/bin/bash
set -e # Arrêter le script à la première erreur

# Script runner pour tous les scripts dans scripts/

# Charger .env si il existe
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Couleurs
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Script Runner${NC}"
echo "================="

# Lister tous les fichiers dans scripts/ en excluant ceux qui commencent par "_"
# et trier par ordre alphabétique
scripts=($(find scripts -type f \( -name "*.ts" -o -name "*.sh" \) ! -name "_*" | sort))

# Si aucun script trouvé
if [ ${#scripts[@]} -eq 0 ]; then
  echo "❌ Aucun script trouvé dans scripts/"
  exit 1
fi

echo -e "${BLUE}Scripts disponibles:${NC}"
echo ""

# Afficher la liste des scripts avec numéros
for i in "${!scripts[@]}"; do
  script_file="${scripts[$i]}"
  script_name=$(basename "$script_file")
  echo "  $((i+1)). $script_name"
done

echo ""
echo -n -e "${YELLOW}Choisissez un script (1-${#scripts[@]}): ${NC}"
read choice

# Valider le choix
if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#scripts[@]} ]; then
  echo "❌ Choix invalide"
  exit 1
fi

# Récupérer le script sélectionné
selected_script="${scripts[$((choice-1))]}"
script_name=$(basename "$selected_script")

echo ""
echo -e "${GREEN}▶ Exécution de: $script_name${NC}"
echo "================================="

# Changer vers le répertoire scripts pour l'exécution
cd scripts

# Déterminer comment exécuter le script
script_basename=$(basename "$selected_script")
case "$script_basename" in
  *.sh)
    bash "$script_basename"
    ;;
  *.ts)
    deno run --allow-all "$script_basename"
    ;;
  *)
    echo "❌ Type de script non supporté"
    exit 1
    ;;
esac

echo ""
echo -e "${GREEN}✅ Terminé${NC}"
