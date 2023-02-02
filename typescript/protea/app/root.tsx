import type { LinksFunction, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import {
  Links,
  LiveReload,
  Meta,
  Scripts,
  ScrollRestoration,
  useCatch,
  useLoaderData,
  useMatches
} from '@remix-run/react'
import {
  WalletLayout,
  LandingLayout,
  Error,
  FocusLayout,
  Layouts
} from '~/components'
import type { ReactNode } from 'react'
import styles from '~/styles/app.css'
import { hasUserSession } from './lib/kratos.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'

const metaContent = {
  title: 'Fynbos',
  description: 'Building the better way to pay.'
}

export const meta: MetaFunction = () => {
  return {
    title: metaContent.title,
    'theme-color': '#FFE4E6',
    description: metaContent.description,
    viewport: 'width=device-width,initial-scale=1',

    // Open Graph / Facebook
    'og:title': metaContent.title,
    'og:type': 'website',
    'og:url': 'https://fynbos.app/',
    'og:description': metaContent.description,
    'og:image': '/fynbos.png',

    // Twitter
    'twitter:site': '@fynbosdev',
    'twitter:card': 'summary_large_image',
    'twitter:url': 'https://fynbos.app/',
    'twitter:title': metaContent.title,
    'twitter:description': metaContent.description,
    'twitter:image': '/fynbos_SEO.png'
  }
}

export const links: LinksFunction = () => {
  return [
    { rel: 'stylesheet', href: styles },
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico?v=2' }
  ]
}

function Document({
  children,
  title = 'Fynbos'
}: {
  children: ReactNode
  title?: string
}) {
  return (
    <html lang='en'>
      <head>
        <Meta />
        <Links />
      </head>
      <body className='theme-blue bg-app font-sans text-base font-normal text-strong antialiased selection:bg-brand/50'>
        {children}
        <ScrollRestoration />
        <Scripts />
        <LiveReload />
      </body>
    </html>
  )
}

export const shouldRevalidate: ShouldRevalidateFunction = ({
  defaultShouldRevalidate,
  nextUrl
}) => {
  /**
   * NOTE: We always revalidate when routing to /.
   * To ensure the layout is in sync on client side navigation.
   * This needs to be done for any route that returns a function in its layout handle.
   */
  if (nextUrl.pathname == '/') return true
  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderArgs) {
  const isUser = hasUserSession(request)
  return json({ isUser, isSignupGated: IS_SIGNUP_GATED })
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()
  const matches = useMatches()

  const layoutHandle = matches[matches.length - 1]?.handle?.layout

  let layout, layoutComponent

  if (typeof layoutHandle === 'function') layout = layoutHandle(isUser)
  else layout = layoutHandle

  if (layout == Layouts.FocusLayout) layoutComponent = <FocusLayout />
  else if (layout == Layouts.WalletLayout) layoutComponent = <WalletLayout />
  else layoutComponent = <LandingLayout />

  return <Document>{layoutComponent}</Document>
}

export function ErrorBoundary({ error }: { error: Error }) {
  return (
    <Document title='An error occurred.'>
      <Error data={{ title: error.message }} />
    </Document>
  )
}

export function CatchBoundary() {
  const caught = useCatch()
  return (
    <Document title='An error occurred.'>
      <Error
        status={caught.status}
        statusText={caught.statusText}
        data={caught.data}
      />
    </Document>
  )
}
