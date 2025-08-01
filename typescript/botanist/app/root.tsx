import type { LinksFunction, V2_MetaFunction } from '@remix-run/node'
import {
  isRouteErrorResponse,
  Links,
  LiveReload,
  Meta,
  Scripts,
  ScrollRestoration,
  useRouteError
} from '@remix-run/react'
import { AdminLayout, Error } from '~/components'
import type { ReactNode } from 'react'
import styles from '~/styles/app.css'

const metaContent = {
  title: 'Interledger Wallet Admin',
  description: ''
}

export const meta: V2_MetaFunction = () => {
  return [
    { title: metaContent.title },
    { name: 'theme-color', content: '#FFE4E6' },
    { name: 'description', content: metaContent.description },
    { name: 'viewport', content: 'width=device-width,initial-scale=1' }
  ]
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
  title = 'Interledger Wallet Admin'
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
        <LiveReload port={443} />
      </body>
    </html>
  )
}

export default function Page() {
  return (
    <Document>
      <AdminLayout />
    </Document>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()
  if (isRouteErrorResponse(error)) {
    return (
      <Document title='An error occurred.'>
        <Error status={error.status} data={{ title: error.data.message }} />
      </Document>
    )
  }
  return (
    <Document title='An error occurred.'>
      <Error data={{ title: 'error.data.message' }} />
    </Document>
  )
}
