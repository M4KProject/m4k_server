const { execSync } = require('child_process');
const { readFileSync, writeFileSync, existsSync } = require('fs');
const { join } = require('path');

// Load .env file
const envPath = join('..', '.env');
if (!existsSync(envPath)) throw new Error(`File ${envPath} does not exist`);

const envContent = readFileSync(envPath, 'utf-8');
const env = {};

// Parse .env content
envContent.split('\n').forEach(line => {
    const trimmed = line.trim();
    if (trimmed && !trimmed.startsWith('#')) {
        const [key, ...valueParts] = trimmed.split('=');
        if (key && valueParts.length > 0) {
            env[key.trim()] = valueParts.join('=').trim().replace(/^["']|["']$/g, '');
        }
    }
});

if (!env.NODERED_PASSWORD) throw new Error('no env NODERED_PASSWORD');

const cmd = `docker run --rm -i --entrypoint="" nodered/node-red node-red-admin hash-pw <<< "${env.NODERED_PASSWORD}"`;

try {
    const stdout = execSync(cmd, { 
        encoding: 'utf-8',
        env: { ...process.env, ...env },
        shell: 'bash'
    });
    
    const hash = stdout.replace('Password:', '').trim();
    env.NODERED_PASSWORD_HASH = hash;
    
    // Stringify env back to .env format
    const newEnvContent = Object.entries(env)
        .map(([key, value]) => `${key}=${value}`)
        .join('\n');
    
    writeFileSync('.env', newEnvContent);
    
} catch (error) {
    throw new Error(`Failed to generate password hash: ${error}`);
}