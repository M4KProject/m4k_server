const { build } = require('esbuild');

const config = {
  entryPoints: ['src/pb-auth.ts', 'src/pb-crud.ts'],
  bundle: true,
  platform: 'node',
  target: 'node16',
  format: 'cjs',
  outdir: 'dist',
  external: ['pocketbase', 'node-red'],
  sourcemap: true,
  minify: false
};

// Build
build(config).catch(() => process.exit(1));