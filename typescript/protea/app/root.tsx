import type { LinksFunction, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  Links,
  LiveReload,
  Meta,
  Scripts,
  ScrollRestoration,
  useCatch,
  useLoaderData,
  useLocation
} from '@remix-run/react'
import { WalletLayout, LandingLayout, Error, FocusLayout } from '~/components'
import type { ReactNode } from 'react'
import styles from '~/styles/app.css'
import { hasUserSession } from './lib/kratos.server'

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
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico?v=2' },
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=Source+Code+Pro&family=Overpass+Mono&family=Inter:wght@400;500&family=Poppins:wght@400;500&display=swap'
    },
    {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@24,400,0,0'
    }
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
        <title>{title}</title>
        <meta charSet='utf-8' />
        <meta name='viewport' content='' />
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

export async function loader({ request }: LoaderArgs) {
  const isUser = await hasUserSession(request)
  return json({ isUser })
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()
  const location = useLocation()
  let isFocussed =
    location.pathname.startsWith('/signup') ||
    location.pathname.startsWith('/waitlist') ||
    location.pathname.startsWith('/contact') ||
    location.pathname.startsWith('/login') ||
    location.pathname.startsWith('/linked-account') ||
    location.pathname.startsWith('/personal-details') ||
    location.pathname.startsWith('/logout') ||
    location.pathname.startsWith('/recovery') ||
    location.pathname.startsWith('/settings/password') ||
    location.pathname.startsWith('/payment-pointer')

  return (
    <Document>
      {isFocussed && <FocusLayout />}
      {!isFocussed && isUser && <WalletLayout />}
      {!isFocussed && !isUser && <LandingLayout />}
    </Document>
  )
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
