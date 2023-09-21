/** @type {import('@remix-run/dev').AppConfig} */
module.exports = {
  tailwind: true,
  postcss: true,
  future: {
    v2_errorBoundary: true,
    // v2_meta: true,
    v2_dev: true,
    v2_routeConvention: true,
    v2_normalizeFormMethod: true
  },
  serverDependenciesToBundle: 'all',
  appDirectory: 'app',
  assetsBuildDirectory: 'public/build',
  serverModuleFormat: 'cjs',
  publicPath: `${process.env.REMIX_PUBLIC_PATH || ''}/build/`,
  ignoredRouteFiles: ['.*', '**/*.stories.tsx', '**/*.test.{ts,tsx}'],
  sourcemap: true
}
