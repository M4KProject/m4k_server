#!/usr/bin/env -S deno run --allow-net --allow-env --allow-read

import { pbAuth, pbFetch } from "./_helpers.ts";

// 1️⃣ Authentification
await pbAuth();

// 2️⃣ Lecture du fichier pb_schema.json
const schemaText = await Deno.readTextFile("pb_schema.json");
const schema = JSON.parse(schemaText);

if (!Array.isArray(schema)) {
  console.error("❌ Le fichier pb_schema.json n'est pas un tableau de collections.");
  Deno.exit(1);
}

// 3️⃣ Récupérer les collections existantes pour déterminer update vs create
const { items: existingCollections } = await pbFetch({
  url: "/api/collections?perPage=500",
  method: "GET",
});

const existingMap = new Map(existingCollections.map((c: any) => [c.name, c]));

for (const coll of schema) {
  const existing = existingMap.get(coll.name);

  if (existing) {
    console.log(`🔄 Mise à jour de la collection : ${coll.name}`);

    // On enlève id pour éviter un conflit dans le payload
    const { id, ...payload } = coll;

    try {
      await pbFetch({
        method: "PATCH",
        url: `/api/collections/${(existing as any).id}`,
        json: payload,
      });
      console.log(`✅ Collection mise à jour : ${coll.name}`);
    } catch (err) {
      console.error(`❌ Erreur mise à jour ${coll.name} :`, err);
    }
  } else {
    console.log(`➕ Création de la collection : ${coll.name}`);

    // On enlève id pour que PocketBase en génère un nouveau
    const { id, ...payload } = coll;

    try {
      await pbFetch({
        method: "POST",
        url: `/api/collections`,
        json: payload,
      });
      console.log(`✅ Collection créée : ${coll.name}`);
    } catch (err) {
      console.error(`❌ Erreur création ${coll.name} :`, err);
    }
  }
}
