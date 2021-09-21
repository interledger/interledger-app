/** @type {import('next').NextConfig} */
const withMDX = require('@next/mdx')({
  extension: /\.blog.mdx?$/,
  options: {
    remarkPlugins: [],
    rehypePlugins: [require('@mapbox/rehype-prism'), require('rehype-slug')]
  }
})
module.exports = withMDX({
  pageExtensions: ['page.tsx', 'blog.mdx'],
  reactStrictMode: true
})
