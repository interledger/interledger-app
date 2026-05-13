import type { LinksFunction, LoaderFunctionArgs, MetaFunction } from 'react-router';
import {
  Links,
  Meta,
  Scripts,
  ScrollRestoration,
  isRouteErrorResponse,
  useLoaderData,
  useNavigation,
  useRouteError,
  data,
  type ShouldRevalidateFunction,
} from 'react-router';
import { captureException } from '@sentry/react-router'
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
  WalletGrid
} from '~/components'
import { Scaffold } from '~/components/Scaffold'
import { TotpChallengeGlobal } from '~/components/TotpChallengeGlobal'
import { hasUserSession } from '~/lib/kratos/session.server'
import { getSnackbar } from '~/lib/snackbar.server'
import styles from '~/styles/app.css?url'
import { PendingConfirmationsLoader } from './components/PendingConfirmationsLoader'
import { getFeatures } from './data/wallet.server'
import { Features } from './generated/connect/backend/v1/backend_pb'
import { isConnectError } from './lib/error.server'
import { grpc } from './lib/grpc.server'
import { getPusherArgs } from './lib/pusher.server'
import { emailVerificationGuard, recoveryLinkSessionInvalidationGuard, withAAL2Guard } from './lib/totp.server'
import { usePusher } from './lib/usePusher'
import { PtiConfigProvider } from './lib/pti-context'
import { Route } from './+types/root';
import { envValue } from './env.server'

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
      </body>
    </html>
  )
}

export async function loader({ request }: LoaderFunctionArgs) {
  const isUser = hasUserSession(request)
  const { snackbar, headers } = await getSnackbar(request)
  const pusherArgs = await getPusherArgs(request)

  const url = new URL(request.url)
  const pathname = url.pathname
  let features = new Features()
  let isDisabled = false
  let walletAddress = ''
  const env = {
    fynbosEnv: envValue("FYNBOS_ENV"),
    sentryDsn: envValue("SENTRY_DSN"),
    sentryRelease: envValue("SENTRY_RELEASE"),
    targetHost: envValue("TARGET_HOST"),
    supportEmail: envValue("SUPPORT_EMAIL")
  }

  if (!isUser) {
    return data({
      isDisabled,
      walletAddress,
      isUser: false,
      features,
      snackbar,
      pusherArgs,
      env
    }, { headers })
  }

  await recoveryLinkSessionInvalidationGuard(pathname, request)
  await emailVerificationGuard(pathname, request)
  await withAAL2Guard(pathname, request, async () => {
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
  })

  return data({
    isDisabled,
    walletAddress,
    isUser,
    features,
    snackbar,
    pusherArgs,
    env
  }, { headers })
}

export type RootLoaderData = Route.ComponentProps['loaderData']

function Page() {
  const { pusherArgs, env, isDisabled, walletAddress } =
    useLoaderData<RootLoaderData>()
  usePusher(pusherArgs, ['cardReady'])

  return (
    <PtiConfigProvider>
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
      </Document>
    </PtiConfigProvider>
  )
}
export default Page

export function ErrorBoundary() {
  const error = useRouteError()
  captureException(error)

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
