/**
 * @type {import('@remix-run/dev/config').AppConfig}
 */
module.exports = {
  appDirectory: 'app',
  serverBuildTarget: 'node-cjs',
  assetsBuildDirectory: 'public/build',
  publicPath: '/build/',
  devServerPort: 8002,
  ignoredRouteFiles: [
    '.*',
    '**/*.draft.mdx',
    '**/*.stories.tsx',
    '**/*.test.{ts,tsx}'
  ],
  sourcemap: true,
  mdx: async (filename) => {
    const [rehypePrism] = await Promise.all([
      import('@mapbox/rehype-prism').then((mod) => mod.default)
    ])

    return {
      rehypePlugins: [rehypePrism]
    }
  }
}
