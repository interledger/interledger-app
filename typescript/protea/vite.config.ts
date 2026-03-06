import { reactRouter } from "@react-router/dev/vite";
import { defineConfig } from "vite";
import { nodePolyfills } from "vite-plugin-node-polyfills";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  // The public path translates to Vite's base config.
  base: process.env.REMIX_PUBLIC_PATH
    ? `${process.env.REMIX_PUBLIC_PATH}/build/`
    : '/',

  server: {
    allowedHosts: ["interledger.test"],
    port: 3000,
    host: true,
    hmr: {
      port: 8002,
      clientPort: 443,
      protocol: 'wss',
      path: '/socket',
    },
  },

  build: {
    // The sourcemap gets deleted in the Dockerfile anyways
    sourcemap: process.env.NODE_ENV === 'production' ? 'hidden' : true,
  },
  plugins: [
    reactRouter(),
    tsconfigPaths(),
    nodePolyfills({
      include: ['os', 'constants', 'buffer', 'assert', 'process'],
      globals: { Buffer: true, process: true },
    }),
  ],
  optimizeDeps: {
    include: [
      // Explicitly include all major dependencies to avoid a mid-load optimization trigger.
      // If discovered lazily (like navigating directly to /signup during E2E), Vite re-optimizes 
      // chunks, causing a version mismatch and the "TypeError: Cannot read properties of null (reading 'useContext')" crash.
      'react',
      'react/jsx-runtime',
      'react/jsx-dev-runtime',
      'react-dom',
      'react-dom/client',
      'react-router',
      '@sentry/react-router',
      'react-datocms',
      'framer-motion',
      'zustand',
      '@headlessui/react',
      'libphonenumber-js',
      'clsx',
      'uuid',
      '@bufbuild/protobuf',
      '@bufbuild/connect',
      '@bufbuild/connect-node',
      'pusher-js',
      'node-rsa',
      'datocms-structured-text-utils',
      'pino',
      '@redis/client',
      'vite-plugin-node-polyfills/shims/buffer',
      'vite-plugin-node-polyfills/shims/global',
      'vite-plugin-node-polyfills/shims/process'
    ]
  },
  ssr: {
    // react-datocms is ESM-only (no CJS build). Vite's SSR output is CJS by default, and when
    // a package is externalized Vite emits require() for it — which Node can't execute against
    // a pure-ESM package. noExternal forces Vite to bundle react-datocms inline into the SSR
    // output and transform it to CJS, avoiding the runtime crash.
    noExternal: ['react-datocms'],
  },
});
