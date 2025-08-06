#!/usr/bin/env -S deno run --allow-net --allow-env --allow-read

import { pbAuth } from "./_helpers.ts";

try {
  const result = await pbAuth();
  console.log("✅ Authentication successful", result);
} catch (error) {
  console.error("❌ Authentication failed:", error);
}