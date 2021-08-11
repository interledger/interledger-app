import '../styles/globals.css'
import type { AppProps } from 'next/app'
import Head from 'next/head'

function MyApp({ Component, pageProps }: AppProps) {

  const metaContent = {
    title: "Fynbos",
    description: "Payments infrastructure for the future.",
  }

  return (
    <>
      <Head>
        <title>{metaContent.title}</title>
        <meta content={metaContent.description} name="description" />
        <meta content={metaContent.title} property="og:title" />
        <meta content={metaContent.description} property="og:description" />
        <meta content='/fynbos_SEO.png' property="og:image" />
        <meta property="og:type" content="website" />
        <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width" />
      </Head>
      <Component {...pageProps} />
    </> 
  )
}
export default MyApp
