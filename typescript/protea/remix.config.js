/** @type {import('@remix-run/dev').AppConfig} */
module.exports = {
  tailwind: true,
  postcss: true,
  dev: {
    port: 8002
    // Can't use these unless we break out the express server
    // tlsKey: 'key.pem', // relative to cwd
    // tlsCert: 'cert.pem' // relative to cwd
  },
  // serverDependenciesToBundle: 'all',
  appDirectory: 'app',
  assetsBuildDirectory: 'public/build',
  serverModuleFormat: 'cjs',
  publicPath: `${process.env.REMIX_PUBLIC_PATH || ''}/build/`,
  ignoredRouteFiles: ['.*', '**/*.stories.tsx', '**/*.test.{ts,tsx}'],
  sourcemap: true,
  browserNodeBuiltinsPolyfill: { modules: { os: true } }
}
