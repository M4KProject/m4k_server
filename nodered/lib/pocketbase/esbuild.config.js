const { build } = require('esbuild');

const config = {
  entryPoints: ['src/pocketbase.ts'],
  bundle: true,
  platform: 'node',
  target: 'node16',
  format: 'cjs',
  outfile: 'dist/pocketbase.js',
  external: ['pocketbase', 'node-red'],
  sourcemap: true,
  minify: false
};

// Build
build(config).catch(() => process.exit(1));