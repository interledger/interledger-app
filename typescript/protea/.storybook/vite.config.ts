import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

// Storybook-specific Vite config — React Router plugin intentionally excluded.
// The plugin requires an app entry point and SSR context that Storybook doesn't
// provide, causing it to throw during dep optimisation (before viteFinal runs).
// Storybook is pointed to this file via framework.options.builder.viteConfigPath.
export default defineConfig({
  plugins: [
    tsconfigPaths(),
    nodePolyfills({
      include: ['os', 'constants', 'buffer', 'assert', 'process'],
      globals: { Buffer: true, process: true },
    }),
  ],
})
