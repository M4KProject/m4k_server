#!/bin/bash
set -e # Arrêter le script à la première erreur

cd ../pocketbase

docker compose --env-file ../.env up -d --build