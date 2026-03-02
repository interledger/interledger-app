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
    sourcemap: true,
  },

  plugins: [
    reactRouter(),
    tsconfigPaths(),
    nodePolyfills({
      include: ['os', 'constants', 'buffer', 'assert', 'process'],
      globals: { Buffer: true, process: true },
    }),
  ],
  ssr: {
    noExternal: ['react-datocms'],
    external: ['pusher-js'],
  },
});
