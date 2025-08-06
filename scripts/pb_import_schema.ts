#!/usr/bin/env -S deno run --allow-net --allow-env --allow-read

import { pbAuth, pbFetch } from "./_helpers.ts";

// 0️⃣ Charger les variables d'environnement depuis .env
try {
  const envText = await Deno.readTextFile("../.env");
  const envLines = envText.split('\n');
  for (const line of envLines) {
    const [key, value] = line.split('=');
    if (key && value) {
      Deno.env.set(key.trim(), value.trim());
    }
  }
  console.log("✅ Variables d'environnement chargées");
} catch (err) {
  console.warn("⚠️ Impossible de charger .env:", (err as Error).message);
}

// 1️⃣ Authentification
await pbAuth();

// 2️⃣ Lecture du fichier pb_schema.json
const schemaText = await Deno.readTextFile("../pocketbase/pb_schema.json");
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

const existingMap = new Map(existingCollections.map((c: any) => [c.name, c])) as Map<string, any>;

// 4️⃣ Créer d'abord les collections sans références (ordre de dépendances)
const collectionOrder = ['groups', 'medias', 'contents', 'devices', 'jobs', 'members', 'transcodes'];
const remaining = [...schema];

// Traiter les collections dans l'ordre recommandé
for (const collName of collectionOrder) {
  const collIndex = remaining.findIndex(c => c.name === collName);
  if (collIndex === -1) continue;
  
  const coll = remaining[collIndex];
  remaining.splice(collIndex, 1);
  
  await processCollection(coll, existingMap);
}

// Traiter les collections restantes
for (const coll of remaining) {
  await processCollection(coll, existingMap);
}

async function processCollection(coll: any, existingMap: Map<string, any>) {
  const existing = existingMap.get(coll.name);

  if (existing) {
    console.log(`🔄 Mise à jour de la collection : ${coll.name}`);

    // On enlève id pour éviter un conflit dans le payload
    const { id: _id, ...payload } = coll;

    try {
      await pbFetch({
        method: "PATCH",
        url: `/api/collections/${(existing as any).id}`,
        json: payload,
      });
      console.log(`✅ Collection mise à jour : ${coll.name}`);
    } catch (err) {
      console.error(`❌ Erreur mise à jour ${coll.name} :`, err);
      console.error('Payload:', JSON.stringify(payload, null, 2));
    }
  } else {
    console.log(`➕ Création de la collection : ${coll.name}`);

    // On enlève id et relations pour créer d'abord la structure de base
    const { id: _id, fields, ...payload } = coll;
    
    // Séparer les champs relation des autres
    const relationFields = fields?.filter((f: any) => f.type === 'relation') || [];
    const nonRelationFields = fields?.filter((f: any) => f.type !== 'relation') || [];
    
    const basePayload = { ...payload, fields: nonRelationFields };

    try {
      const created = await pbFetch({
        method: "POST",
        url: `/api/collections`,
        json: basePayload,
      });
      console.log(`✅ Collection créée : ${coll.name}`);
      
      // Ajouter les champs de relation après création si nécessaire
      if (relationFields.length > 0) {
        console.log(`🔗 Ajout des relations pour : ${coll.name}`);
        try {
          await pbFetch({
            method: "PATCH",
            url: `/api/collections/${created.id}`,
            json: { fields: [...nonRelationFields, ...relationFields] },
          });
          console.log(`✅ Relations ajoutées pour : ${coll.name}`);
        } catch (relErr) {
          console.error(`⚠️ Erreur ajout relations ${coll.name} :`, relErr);
        }
      }
      
    } catch (err) {
      console.error(`❌ Erreur création ${coll.name} :`, err);
      console.error('Payload:', JSON.stringify(basePayload, null, 2));
    }
  }
}
