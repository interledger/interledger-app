import 'styles/main.css'
import App, { AppProps as NextAppProps, AppContext } from 'next/app'
import Head from 'next/head'
import { Session } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { AuthContext } from 'contexts/auth'
import { kratos } from 'lib/kratos'
import { Routes } from 'components'

interface AppProps extends NextAppProps {
  session: Session | undefined
}

function MyApp({ Component, pageProps, session }: AppProps) {
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
      <AuthContext.Provider value={{ session }}>
        <Component {...pageProps} />
      </AuthContext.Provider>
    </>
  )
}

// This turns off automatic static optimization unless a page
// explicitly specifies optimization. Every page is then SSR
// where we fetch the session data from Kratos.
MyApp.getInitialProps = async (appContext: AppContext) => {
  const appProps = await App.getInitialProps(appContext)

  // forward cookie to Kratos since this is now happening server side.
  const cookie = appContext.ctx.req?.headers.cookie
  const session = await kratos.toSession(undefined, cookie).then(res => res.data).catch((err: AxiosError) => {
    switch (err.response?.status) {
      case 403:
      // This is a legacy error code thrown. See code 422 for
      // more details.
      case 422:
        // This status code is returned when we are trying to
        // validate a session which has not yet completed
        // it's second factor
        redirect(Routes.login + '?aal=aal2', appContext)
        return undefined
      case 401:
        if (isProtected(appContext.ctx.pathname)) {
          redirect(Routes.login, appContext)
        }
        return undefined
    }

    // Something else happened!
    redirect('/error', appContext)
    return undefined
  })
  return { ...appProps, session }
}

// TODO: pattern for protected routes
const isProtected = (path: string): boolean => {
  return path.includes("/organisation")
}

const redirect = (path: string, appContext: AppContext): void => {
  if (appContext.ctx.res) {
    appContext.ctx.res?.writeHead(302, {
      Location: path,
    })
    appContext.ctx.res.end()
    return
  }

  appContext.router.push(path)
}

export default MyApp
