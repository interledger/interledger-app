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
  useRouteError,
  type ShouldRevalidateFunction
} from '@remix-run/react'
import { captureRemixErrorBoundaryError, withSentry } from '@sentry/remix'
import clsx from 'clsx'
import { type ReactNode } from 'react'
import {
  Card,
  CardContent,
  CardCopy,
  CardHeader,
  CardTitle,
  Error,
  GridColumn,
  InterledgerLogo,
  LiveReload,
  WalletGrid
} from '~/components'
import { Scaffold } from '~/components/Scaffold'
import { TotpChallengeGlobal } from '~/components/TotpChallengeGlobal'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import { getSnackbar } from '~/lib/snackbar.server'
import styles from '~/styles/app.css'
import { PendingConfirmationsLoader } from './components/PendingConfirmationsLoader'
import { getFeatures } from './data/wallet.server'
import { Features } from './generated/connect/backend/v1/backend_pb'
import { isConnectError } from './lib/error.server'
import { grpc } from './lib/grpc.server'
import { getPusherArgs } from './lib/pusher.server'
import { NON_FULL_SESSION_ROUTES, isTotpSet } from './lib/totp.server'
import { usePusher } from './lib/usePusher'

export const shouldRevalidate: ShouldRevalidateFunction = ({
  actionResult,
  defaultShouldRevalidate
}) => {
  if (actionResult && 'shouldRevalidate' in actionResult) {
    return actionResult.shouldRevalidate === true
  }
  return defaultShouldRevalidate
}

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
  env?: Record<string, unknown>
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
          navigation.state === 'submitting' && 'cursor-progress'
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
  let isDisabled = false
  let walletAddress = ''
  const url = new URL(request.url)
  const pathname = url.pathname

  if (isUser && !NON_FULL_SESSION_ROUTES.includes(pathname)) {
    const session = await getUserSession(request)
    const totpAvailable = await isTotpSet(session, request.headers)
    if (!totpAvailable) {
      return redirect('/totp/two-factor-authentication')
    }

    features = await getFeatures(request)
    if (
      features &&
      !features.accountEnabled &&
      url.pathname !== '/wallet-address'
    ) {
      const wallet = await grpc.getWalletInfo(request, {})
      if (!isConnectError(wallet)) {
        walletAddress = wallet.url
      }
      isDisabled = true
    }
  }

  const pusherArgs = await getPusherArgs(request)

  return json({
    isDisabled,
    walletAddress,
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
  const { pusherArgs, env, isDisabled, walletAddress } =
    useLoaderData<typeof loader>()
  usePusher(pusherArgs, ['cardReady'])

  return (
    <Document>
      {isDisabled ? (
        <Unavailable walletAddress={walletAddress} />
      ) : (
        <>
          <Scaffold />
          <PendingConfirmationsLoader walletId={pusherArgs.walletId} />
          <TotpChallengeGlobal />
        </>
      )}
      <script
        dangerouslySetInnerHTML={{
          __html: `window.ENV = ${JSON.stringify(env)}`
        }}
      />
      <TailwindIndicator />
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

function Unavailable({ walletAddress }: { walletAddress: string }) {
  return (
    <main className='mb-32 mt-32 w-full px-4'>
      <WalletGrid>
        <GridColumn className='col-span-8 col-start-1 space-y-12 xl:col-start-3'>
          <InterledgerLogo className='max-w-sm self-center' />
          <Card className='flex !flex-row'>
            <CardContent className='ml-2 text-lg'>
              The application is not yet available in your location, but do not
              worry we are working tirelessly to solve it as fast as possible!{' '}
              <br />
              We will notify you by email once the application becomes fully
              functional in your region.
            </CardContent>
          </Card>
          {walletAddress && (
            <Card>
              <CardHeader>
                <CardTitle>Your wallet address has been reserved!</CardTitle>
              </CardHeader>
              <CardCopy
                copyContent={walletAddress}
                shareData={{
                  title: 'Wallet address',
                  text: 'You can pay me using my wallet address.',
                  url: walletAddress
                }}
                success='Wallet address copied to clipboard.'
                copyError="Couldn't copy to clipboard."
                shareError="Couldn't share wallet address."
              >
                {walletAddress}
              </CardCopy>
            </Card>
          )}
        </GridColumn>
      </WalletGrid>
    </main>
  )
}

function TailwindIndicator() {
  if (process.env.NODE_ENV === 'production') return null

  return (
    <div className='fixed bottom-1 left-1 z-50 flex h-6 w-6 items-center justify-center rounded-full bg-gray-800 p-3 font-mono text-xs text-white'>
      <div className='block sm:hidden'>xs</div>
      <div className='hidden sm:block md:hidden'>sm</div>
      <div className='hidden md:block lg:hidden'>md</div>
      <div className='hidden lg:block xl:hidden'>lg</div>
      <div className='hidden xl:block 2xl:hidden'>xl</div>
      <div className='hidden 2xl:block'>2xl</div>
    </div>
  )
}
