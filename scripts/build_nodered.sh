#!/bin/bash
set -e # Arrêter le script à la première erreur

cd ../nodered

docker compose --env-file ../.env up -d --build --remove-orphans