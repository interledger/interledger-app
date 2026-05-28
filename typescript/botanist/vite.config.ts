import { reactRouter } from '@react-router/dev/vite'
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
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-dom/client',
      'react-router',
      '@headlessui/react',
      'clsx',
      'luxon',
      '@grpc/grpc-js',
      '@protobuf-ts/grpc-transport',
      '@protobuf-ts/runtime',
      '@protobuf-ts/runtime-rpc'
    ]
  },
  plugins: [reactRouter(), tsconfigPaths()]
})
