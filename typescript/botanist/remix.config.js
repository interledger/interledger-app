/**
 * @type {import('@remix-run/dev/config').AppConfig}
 */
module.exports = {
  appDirectory: 'app',
  serverBuildTarget: 'node-cjs',
  assetsBuildDirectory: 'public/build',
  publicPath: '/build/',
  devServerPort: 8003,
  ignoredRouteFiles: ['.*', '**/*.draft.mdx'],
  sourcemap: true
}
