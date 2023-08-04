/**
 * @type {import('@remix-run/dev/config').AppConfig}
 */
module.exports = {
  tailwind: true,
  postcss: true,
  future: {
    v2_errorBoundary: true,
    // v2_meta: true,
    v2_routeConvention: true,
    v2_normalizeFormMethod: true
  },
  appDirectory: 'app',
  assetsBuildDirectory: 'public/build',
  serverModuleFormat: 'cjs',
  publicPath: process.env.FYNBOS_ENV == 'prod' || process.env.FYNBOS_ENV == 'dev' ? 'https://cdn.fynbos.app/protea/build/' : '/build/',
  ignoredRouteFiles: ['.*', '**/*.stories.tsx', '**/*.test.{ts,tsx}'],
  sourcemap: true
}
