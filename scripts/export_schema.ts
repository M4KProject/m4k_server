#!/usr/bin/env -S deno run --allow-net --allow-env

import { pbAuth, pbFetch } from "./_helpers.ts";

await pbAuth();

// Récup collections
const schema = await (async () => {
  const data = await pbFetch({ url: '/api/collections?perPage=500' });

  const collections = data.items.filter((c: any) => !c.name.startsWith('_') && c.name !== 'users');

  return collections;
})();

// Enregistrer dans pb_schema.json
await Deno.writeTextFile('pb_schema.json', JSON.stringify(schema, null, 2));
console.log('✅ Schéma enregistré dans pb_schema.json');