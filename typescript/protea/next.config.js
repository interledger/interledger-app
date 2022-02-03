/** @type {import('next').NextConfig} */

const kratosClient = process.env.KRATOS_CLIENT || 'http://fynbos.test'
const kratosServer = process.env.KRATOS_SERVER || 'http://kratos-public'
const apolloClient = process.env.APOLLO_CLIENT || 'http://fynbos.test/graphql'
const apolloServer = process.env.APOLLO_SERVER || 'http://backend/graphql'

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
    kratosClient: kratosClient,
    kratosServer: kratosServer,
    apolloClient: apolloClient,
    apolloServer: apolloServer
  }
})
