#!/bin/bash
set -e

node _make.js

cd lib/pocketbase
npm install
npm run build
cd ../../

source .env

docker compose up -d --build --remove-orphans

docker compose restart
