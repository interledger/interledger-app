import 'styles/main.css'
import { AppProps as NextAppProps } from 'next/app'
import Head from 'next/head'
import { ApolloProvider } from '@apollo/client'
import { apolloClient } from 'lib/apollo'

function MyApp({ Component, pageProps }: NextAppProps) {
  const metaContent = {
    title: 'Fynbos',
    description: 'Connecting the Internet economy.'
  }

  return (
    <>
      <Head>
        <title>{metaContent.title}</title>
        <meta name="theme-color" content="#FDE2E6" />
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
      <ApolloProvider client={apolloClient}>
        <Component {...pageProps} />
      </ApolloProvider>
    </>
  )
}

export default MyApp
