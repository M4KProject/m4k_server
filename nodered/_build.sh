#!/bin/bash
set -e

deno run --allow-all ./_make.ts

source .env

docker compose up -d --build --remove-orphans

docker compose restart
