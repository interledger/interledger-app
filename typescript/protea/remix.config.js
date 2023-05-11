/**
 * @type {import('@remix-run/dev/config').AppConfig}
 */
module.exports = {
  tailwind: true,
  future: {
    v2_errorBoundary: true,
    // v2_meta: true,
    v2_routeConvention: true,
    v2_normalizeFormMethod: true
  },
  appDirectory: 'app',
  assetsBuildDirectory: 'public/build',
  serverModuleFormat: 'cjs',
  publicPath: '/build/',
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
