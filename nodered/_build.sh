#!/bin/bash
set -e

deno run --allow-all ./_make.ts

cd lib/pocketbase
npm install
npm run build
cd ../../

source .env

docker compose up -d --build --remove-orphans

docker compose restart
