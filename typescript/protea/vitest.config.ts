/// <reference types="vitest" />
/// <reference types="vite/client" />

import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  test: {
    globals: true,
    environment: 'jsdom',
    // setupFiles: ['./test/setup-test-env.ts'],
    include: ['./**/*.test.{ts,tsx}'],
    exclude: ['.*\\/node_modules\\/.*', '.*\\/build\\/.*'],
    env: {
      SESSION_SECRET: 'foobar'
    }
  }
})
