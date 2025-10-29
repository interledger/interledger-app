import type {
  LinksFunction,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Links,
  Meta,
  Scripts,
  ScrollRestoration,
  isRouteErrorResponse,
  useLoaderData,
  useNavigation,
  useRouteError
} from '@remix-run/react'
import { captureRemixErrorBoundaryError, withSentry } from '@sentry/remix'
import clsx from 'clsx'
import { type ReactNode } from 'react'
import { Error, LiveReload } from '~/components'
import { Scaffold } from '~/components/Scaffold'
import { TotpChallengeGlobal } from '~/components/TotpChallengeGlobal'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import { getSnackbar } from '~/lib/snackbar.server'
import styles from '~/styles/app.css'
import { PendingConfirmationsLoader } from './components/PendingConfirmationsLoader'
import { getFeatures } from './data/wallet.server'
import { Features } from './generated/connect/backend/v1/backend_pb'
import { getPusherArgs } from './lib/pusher.server'
import { NON_FULL_SESSION_ROUTES, isTotpSet } from './lib/totp.server'
import { usePusher } from './lib/usePusher'

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

export async function loader({ request }: LoaderFunctionArgs) {
  const isUser = hasUserSession(request)
  const snackbar = await getSnackbar(request)

  let features = new Features()
  const url = new URL(request.url)
  const pathname = url.pathname

  if (isUser && !NON_FULL_SESSION_ROUTES.includes(pathname)) {
    const session = await getUserSession(request)
    const totpAvailable = await isTotpSet(session, request.headers)
    if (!totpAvailable) {
      return redirect('/totp/two-factor-authentication')
    }

    features = await getFeatures(request)
    if (features && !features.accountEnabled) {
      return redirect('/unavailable')
    }
  }

  const pusherArgs = await getPusherArgs(request)

  return json({
    isUser,
    features,
    snackbar,
    pusherArgs,
    env: {
      fynbosEnv: process.env.FYNBOS_ENV,
      sentryDsn: process.env.SENTRY_DSN,
      sentryRelease: process.env.SENTRY_RELEASE,
      segmentApiKey: process.env.SEGMENT_API_KEY || ''
    }
  })
}

function Page() {
  const { pusherArgs, env } = useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['cardReady'])

  return (
    <Document>
      <Scaffold />
      <PendingConfirmationsLoader walletId={pusherArgs.walletId} />
      <TotpChallengeGlobal />
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
