#!/bin/bash
set -e # Arrêter le script à la première erreur

cd ../caddy

docker compose --env-file ../.env up -d --build --remove-orphans
docker compose restart