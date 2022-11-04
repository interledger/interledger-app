/**
 * @type {import('@remix-run/dev/config').AppConfig}
 */

const publicPath = process.env.NODE_ENV === 'production' ? 'https://cdn.fynbos.app/protea/public/' : '/build/'

module.exports = {
  appDirectory: 'app',
  serverBuildTarget: 'node-cjs',
  assetsBuildDirectory: 'public/build',
  publicPath: publicPath,
  devServerPort: 8002,
  ignoredRouteFiles: ['.*', '**/*.draft.mdx'],
  sourcemap: true
}
