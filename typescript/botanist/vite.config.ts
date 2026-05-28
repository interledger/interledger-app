import { reactRouter } from '@react-router/dev/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  server: {
    allowedHosts: ['admin.mgnt.interledger.test'],
    port: 3000,
    host: true,
    hmr: {
      port: 8002,
      clientPort: 443,
      protocol: 'wss',
      path: '/socket'
    }
  },
  build: {
    sourcemap: true
  },
  plugins: [react(), reactRouter(), tsconfigPaths()]
})
