/** @type {import('next').NextConfig} */
const withMDX = require('@next/mdx')({
  extension: /\.blog.mdx?$/,
  options: {
    remarkPlugins: [],
    rehypePlugins: [require('@mapbox/rehype-prism'), require('rehype-slug')]
  }
})
module.exports = withMDX({
  pageExtensions: ['page.tsx', 'blog.mdx', 'api.ts'],
  reactStrictMode: true,
  publicRuntimeConfig: {
    // Will be available on both server and client
    kratosClient: process.env.KRATOS_CLIENT || 'http://fynbos.test',
    kratosServer: process.env.KRATOS_SERVER || 'http://kratos-public',
    apolloClient: process.env.APOLLO_CLIENT || 'http://fynbos.test/graphql',
    apolloServer: process.env.APOLLO_SERVER || 'http://backend/graphql'
  }
})
