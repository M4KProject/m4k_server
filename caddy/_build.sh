#!/bin/bash
set -e

cp ../.env ./.env

docker compose up -d --build --remove-orphans

docker compose restart
