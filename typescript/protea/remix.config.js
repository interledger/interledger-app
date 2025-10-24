/** @type {import('@remix-run/dev').AppConfig} */
module.exports = {
  tailwind: true,
  postcss: true,
  dev: {
    port: 8002
  },
  appDirectory: 'app',
  assetsBuildDirectory: 'public/build',
  serverModuleFormat: 'cjs',
  publicPath: `${process.env.REMIX_PUBLIC_PATH || ''}/build/`,
  ignoredRouteFiles: ['.*', '**/*.stories.tsx', '**/*.test.{ts,tsx}'],
  sourcemap: true,
  browserNodeBuiltinsPolyfill: {
    modules: {
      os: true,
      crypto: true,
      constants: true,
      buffer: true,
      assert: true,
      process: true,
    },
    globals: {
      Buffer: true,
      process: true
    },
  }
}
