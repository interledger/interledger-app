import 'styles/main.css'
import type { AppProps } from 'next/app'
import Head from 'next/head'

export default function MyApp({ Component, pageProps }: AppProps) {
  const metaContent = {
    title: 'Fynbos',
    description: 'Connecting the Internet economy.'
  }

  return (
    <>
      <Head>
        <title>{metaContent.title}</title>
        <meta content={metaContent.description} name='description' />
        <meta name='title' content={metaContent.title} />
        <meta name='description' content={metaContent.description} />

        {/* Open Graph / Facebook */}
        <meta property='og:title' content={metaContent.title} />
        <meta property='og:type' content='website' />
        <meta property='og:url' content='https://fynbos.dev/' />
        <meta property='og:description' content={metaContent.description} />
        <meta property='og:image' content='/fynbos.png' />

        {/*Twitter*/}
        <meta property='twitter:card' content='summary_large_image' />
        <meta property='twitter:url' content='https://fynbos.dev/' />
        <meta property='twitter:title' content='Fynbos' />
        <meta
          property='twitter:description'
          content='Connecting the Internet economy.'
        />
        <meta property='twitter:image' content='/fynbos_SEO.png' />
        <meta
          name='viewport'
          content='minimum-scale=1, initial-scale=1, width=device-width'
        />
      </Head>
      <Component {...pageProps} />
    </>
  )
}
