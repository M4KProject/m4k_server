import { load, stringify } from "https://deno.land/std@0.224.0/dotenv/mod.ts";

const env = await load({ envPath: '../.env' });

if (!env.NODERED_PASSWORD) throw new Error('no env NODERED_PASSWORD');

const cmd = `docker run --rm -i --entrypoint="" nodered/node-red node-red-admin hash-pw <<< "$NODERED_PASSWORD"`;
const command = new Deno.Command('bash', { args: ['-c', cmd], stdout: 'piped', env });
const { success, stdout } = await command.output();

if (!success) throw new Error('Failed to generate password hash');

const hash = new TextDecoder().decode(stdout).replace('Password:', '').trim();

env.NODERED_PASSWORD_HASH = hash;

const newEnv = stringify(env);

await Deno.writeTextFile('.env', newEnv);
