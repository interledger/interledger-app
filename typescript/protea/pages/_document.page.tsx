import Document, {
  Html,
  Head,
  Main,
  NextScript,
  DocumentContext
} from 'next/document'

class MyDocument extends Document {
  static async getInitialProps(ctx: DocumentContext) {
    const initialProps = await Document.getInitialProps(ctx)
    return { ...initialProps }
  }

  render() {
    return (
      <Html>
        <Head>
          <link rel='icon' type='image/x-icon' href='/favicon.ico' />
          <meta property='og:type' content='website' />
          <link rel='preconnect' href='https://fonts.googleapis.com' />
          <link
            href='https://fonts.googleapis.com/css2?family=Source+Code+Pro&family=Overpass+Mono&family=Inter:wght@400;500&family=Poppins:wght@400;500&display=swap'
            rel='stylesheet'
          />
        </Head>
        <body className='antialiased font-body bg-white text-black selection:bg-primary/50'>
          <Main />
          <NextScript />
        </body>
      </Html>
    )
  }
}

export default MyDocument
