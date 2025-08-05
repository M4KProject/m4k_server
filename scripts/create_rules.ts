#!/usr/bin/env -S deno run --allow-net --allow-env

import { pbAuth, pbFetch } from "./_helpers.ts";

// Transforme la règle comme la fonction Go
const toRule = (rule: string): string => {
  rule = " " + rule + " ";
  rule = rule.replace(/ auth /g, " @request.auth.id ");
  rule = rule.replace(/ group_members\./g, " group.members_via_group.");
  rule = rule.replace(/ members\./g, " members_via_group.");
  rule = rule.replace(/\t/g, " ");
  rule = rule.replace(/\n/g, " ");
  rule = rule.replace(/  +/g, " ");
  rule = rule.replace(/\( /g, "(");
  rule = rule.replace(/ \)/g, ")");
  return rule.trim();
}

const computeRules = (name: string) => {
  let viewRule =
    `auth != "" && group_members.user ?= auth && group_members.role ?>= 10`;
  let editRule =
    `auth != "" && group_members.user ?= auth && group_members.role ?>= 20`;

  if (name === "devices") {
    viewRule =
      `auth != "" && (user = auth || group_members.user ?= auth && group_members.role ?>= 10)`;
    editRule =
      `auth != "" && (user = auth || group_members.user ?= auth && group_members.role ?>= 20)`;
  } else if (name === "members") {
    viewRule =
      `auth != "" && (group.user = auth || group_members.user ?= auth && group_members.role ?>= 10)`;
    editRule =
      `auth != "" && (group.user = auth || group_members.user ?= auth && group_members.role ?>= 30)`;
  } else if (name === "groups") {
    viewRule =
      `auth != "" && (user = auth || members.user ?= auth && members.role ?>= 10)`;
    editRule =
      `auth != "" && (user = auth || members.user ?= auth && members.role ?>= 30)`;
  }

  return { viewRule, editRule };
}

await pbAuth();

// 1️⃣ Récupération des collections
const { items: collections } = await pbFetch({
  url: "/api/collections?perPage=500",
  method: "GET",
});

for (const coll of collections) {
  if (coll.name.startsWith("_") || coll.name === "users") continue;

  console.log(`Collection Rule : ${coll.name}`);

  const { viewRule, editRule } = computeRules(coll.name);

  const payload: any = {
    listRule: toRule(viewRule),
    viewRule: toRule(viewRule),
    createRule: toRule(editRule),
    updateRule: toRule(editRule),
    deleteRule: toRule(editRule),
  };

  // Spécial pour groups → createRule unique
  if (coll.name === "groups") {
    payload.createRule = toRule(`auth != "" && user = auth`);
  }

  try {
    await pbFetch({
      method: "PATCH",
      url: `/api/collections/${coll.id}`,
      json: payload,
    });
    console.log(`✅ Rules applied for ${coll.name}`);
  } catch (err) {
    console.error(`❌ Error updating ${coll.name}:`, err);
  }
}

console.log("🎉 All collection rules initialized !");