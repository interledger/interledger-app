import { vitePlugin as remix } from "@remix-run/dev";
import { defineConfig } from "vite";
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
    },
  },

  build: {
    sourcemap: true,
  },

  plugins: [
    remix({
      appDirectory: "app",
      ignoredRouteFiles: [".*", "**/*.stories.tsx", "**/*.test.{ts,tsx}"],
      serverModuleFormat: "cjs",
      // assetsBuildDirectory is replaced by buildDirectory (defaults to "build").
      // With Vite, output automatically goes to buildDirectory/client and buildDirectory/server
      // so if you specifically need public/build, you might set buildDirectory: "public/build" however 
      // the structure will be different than classic Remix compiler. We'll leave the default "build".
    }),
    tsconfigPaths(),
  ],
  ssr: {
    noExternal: ['react-datocms'],
  },
});
