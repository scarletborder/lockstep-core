import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts'],
  dts: true,
  sourcemap: true,
  clean: true,
  format: ['cjs', 'esm', 'iife'],
  globalName: 'LockstepCoreClient',
  minify: false,
  target: 'es2019',
  external: [],
  splitting: false,
  legacyOutput: false,
});
