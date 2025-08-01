import type {
  LinksFunction,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import {
  Link,
  Links,
  Meta,
  Scripts,
  ScrollRestoration,
  isRouteErrorResponse,
  useLoaderData,
  useLocation,
  useNavigation,
  useRouteError
} from '@remix-run/react'
import { captureRemixErrorBoundaryError, withSentry } from '@sentry/remix'
import clsx from 'clsx'
import type { ReactNode } from 'react'
import { AnchorRouter, Error, LiveReload } from '~/components'
import { Scaffold } from '~/components/Scaffold'
import { hasUserSession } from '~/lib/kratos.server'
import { getSnackbar } from '~/lib/snackbar.server'
import { useSegment } from '~/lib/useSegment'
import styles from '~/styles/app.css'
import { getFeatures } from './data/wallet.server'
import { Features } from './generated/connect/backend/v1/backend_pb'

const metaContent = {
  title: 'Interledger Wallet',
  description:
    'Unlock the potential of Open Payments and Web Monetization through the Interledger Wallet and help drive the evolution of digital financial services.'
}

export const meta: MetaFunction = () => [
  { title: metaContent.title },
  {
    property: 'og:title',
    content: metaContent.title
  },
  {
    name: 'description',
    content: metaContent.description
  },
  {
    property: 'og:description',
    content: metaContent.description
  },
  {
    property: 'og:locale',
    content: 'en'
  },
  {
    property: 'og:type',
    content: 'website'
  },
  {
    property: 'og:site_name',
    content: metaContent.title
  }
]

export const links: LinksFunction = () => {
  return [
    { rel: 'stylesheet', href: styles },
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico?v=2' }
  ]
}

type DocumentProps = {
  children: ReactNode
  theme?: 'theme-dark' | 'theme-light' | 'theme-system'
}

function Document({ children, theme = 'theme-system' }: DocumentProps) {
  const navigation = useNavigation()
  return (
    <html lang='en'>
      <head>
        <meta charSet='utf-8' />
        <meta name='viewport' content='width=device-width,initial-scale=1' />
        <meta name='theme-color' content='#FFE4E6' />
        <Meta />
        <Links />
      </head>
      <body
        className={clsx(
          theme,
          'bg-page font-sans text-base font-normal text-strong antialiased selection:bg-brand/50',
          navigation.state == 'submitting' && 'cursor-progress'
        )}
      >
        {children}
        <ScrollRestoration />
        <Scripts />
        <LiveReload port={443} />
      </body>
    </html>
  )
}

const validatePathsList = [
  '/',
  '/pay',
  '/deposit',
  '/withdraw',
  '/accounts',
  '/payments',
  '/personal-details'
]

export const shouldRevalidate: ShouldRevalidateFunction = ({
  defaultShouldRevalidate,
  nextUrl
}) => {
  /**
   * NOTE: We always revalidate when routing to validatePathsList.
   * To ensure the layout is in sync on client side navigation and to validate if account is disabled.
   * This needs to be done for any route that returns a function in its layout handle.
   */
  if (validatePathsList.includes(nextUrl.pathname)) return true
  // TODO: possible also revalidate if an action has been submitted so that we can show global snackbars even on error
  // Could also just return json instead throwing an error
  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderFunctionArgs) {
  const isUser = hasUserSession(request)
  const snackbar = await getSnackbar(request)
  let features = new Features()

  const url = new URL(request.url)
  const pathname = url.pathname

  // if wallet is in a region that is not enabled redirect
  if (isUser && validatePathsList.includes(pathname)) {
    features = await getFeatures(request)

    if (features && !features.accountEnabled) {
      return redirect('/unavailable')
    }
  }

  return json({
    isUser,
    features,
    snackbar,
    env: {
      fynbosEnv: process.env.FYNBOS_ENV,
      sentryDsn: process.env.SENTRY_DSN,
      sentryRelease: process.env.SENTRY_RELEASE,
      segmentApiKey: process.env.SEGMENT_API_KEY || ''
    }
  })
}

function Page() {
  const location = useLocation()
  const { env } = useLoaderData<typeof loader>()
  useSegment(env.segmentApiKey)

  if (location.pathname == '/temp-cloudflare-error') return <CloudFlareError />

  return (
    <Document>
      <Scaffold />
      <script
        dangerouslySetInnerHTML={{
          __html: `window.ENV = ${JSON.stringify(env)}`
        }}
      />
    </Document>
  )
}
export default withSentry(Page)

export function ErrorBoundary() {
  const error = useRouteError()
  captureRemixErrorBoundaryError(error)

  if (isRouteErrorResponse(error)) {
    return (
      <Document>
        <Error
          status={error.status}
          statusText={error.statusText}
          data={error.data}
        />
      </Document>
    )
  }

  return (
    <Document>
      <Error data={{ title: (error as Error).message }} />
    </Document>
  )
}

function CloudFlareError() {
  return (
    <Document>
      <main className='w-full overflow-hidden'>
        <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-8 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
          <div className='relative col-span-full -mx-8 h-44 sm:col-span-6 sm:col-start-2 lg:col-start-4 lg:mx-0'>
            <div className='absolute right-[10.5rem] top-0 h-14 w-14 rounded-bl-full bg-slate-700' />
            <div className='absolute right-28 top-0 h-14 w-14 rounded-tr-full bg-slate-400' />
            <div className='absolute right-14 top-0 h-14 w-14 rounded-full bg-slate-300' />
            <div className='absolute right-0 top-0 h-14 w-14 bg-slate-100' />
            <div className='absolute right-56 top-14 h-14 w-14 rounded-full bg-slate-300' />
            <div className='absolute right-28 top-14 h-14 w-14 rounded-b-full bg-slate-100' />
            <div className='absolute right-14 top-14 h-14 w-14 rounded-full bg-rose-500' />
            <div className='absolute right-0 top-14 h-14 w-14 rounded-tl-full bg-slate-500' />
            <div className='absolute right-0 top-28 h-14 w-14 rounded-full bg-slate-300' />
            <div className='absolute right-14 top-[10.5rem] h-14 w-14 rounded-full bg-slate-600' />
            <div className='absolute right-0 top-[10.5rem] h-14 w-14 rounded-b-full bg-slate-100' />
            {/* Desktop only */}
            <div className='absolute -right-14 top-0 hidden h-14 w-14 rounded-full bg-slate-600 lg:block' />
            <div className='absolute -right-28 top-0 hidden h-14 w-14 rounded-t-full bg-slate-100 lg:block' />
            <div className='absolute -right-14 top-14 hidden h-14 w-14 rounded-full bg-slate-300 lg:block' />
            <div className='absolute -right-14 top-28 hidden h-14 w-14 rounded-full bg-slate-200 lg:block' />
            <div className='absolute -right-28 top-28 hidden h-14 w-14 rounded-br-full bg-slate-300 lg:block' />
            <div className='absolute -right-14 top-[10.5rem] hidden h-14 w-14 rounded-full bg-slate-300 lg:block' />
            <div className='absolute -right-28 top-[10.5rem] hidden h-14 w-14 bg-slate-100 lg:block' />
          </div>
          <div className='col-span-full flex flex-grow flex-col items-start justify-center sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <div className='h-32' />
            <div className='sm:mt-12'>
              <div>
                <h1 className='font-display text-4xl font-medium text-medium'>
                  Unexpected error
                </h1>
                <p className='mt-3 text-weak'>
                  An unexpected error has occurred and our engineers are tending
                  to the issue.
                </p>
                <p className='mt-3 text-weak'>Please refresh your browser.</p>
                <p className='mt-3 text-weak'>
                  If the problem persists, send an email to{' '}
                  <AnchorRouter
                    className='text-primary'
                    to='mailto:support@interledger.app'
                  >
                    support@interledger.app
                  </AnchorRouter>{' '}
                  outlining what you were trying to do.
                </p>
              </div>
              <div className='mt-10'>
                <Link to={'/'}>
                  <span className='text-primary'>Go back home</span>
                </Link>
              </div>
            </div>
          </div>
        </section>
      </main>
      <footer className='fixed bottom-0 flex w-full justify-center p-4'>
        <div className='flex text-xs text-weak'>
          <p>::CLOUDFLARE_ERROR_500S_BOX::</p>
        </div>
      </footer>
    </Document>
  )
}
