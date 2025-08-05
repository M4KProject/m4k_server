#!/bin/bash
set -e # Arrêter le script à la première erreur

docker compose --env-file ../.env up -d --build --remove-orphans