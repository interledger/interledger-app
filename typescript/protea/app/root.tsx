import {
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useCatch
} from 'remix'
import React from 'react'
import type { MetaFunction, LinksFunction } from 'remix'
import styles from '~/styles/app.css'
import { Error } from '~/components'

const metaContent = {
  title: 'Fynbos',
  description: 'Connecting the Internet economy.'
}

export const meta: MetaFunction = () => {
  return {
    title: metaContent.title,
    'theme-color': '#FDE2E6',
    description: metaContent.description,
    viewport: 'width=device-width,initial-scale=1',

    // Open Graph / Facebook
    'og:title': metaContent.title,
    'og:type': 'website',
    'og:url': 'https://fynbos.dev/',
    'og:description': metaContent.description,
    'og:image': '/fynbos.png',

    // Twitter
    'twitter:card': 'summary_large_image',
    'twitter:url': 'https://fynbos.dev/',
    'twitter:title': metaContent.title,
    'twitter:description': metaContent.description,
    'twitter:image': '/fynbos_SEO.png'
  }
}

export const links: LinksFunction = () => {
  return [
    { rel: 'stylesheet', href: styles },
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' },
    { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=Source+Code+Pro&family=Overpass+Mono&family=Inter:wght@400;500&family=Poppins:wght@400;500&display=swap'
    }
  ]
}

function Document({
  children,
  title = 'Fynbos'
}: {
  children: React.ReactNode
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
      <body className='font-body bg-white text-black antialiased selection:bg-primary/50'>
        {children}
        <ScrollRestoration />
        <Scripts />
        <LiveReload />
      </body>
    </html>
  )
}

export default function App() {
  return (
    <Document>
      <Outlet />
    </Document>
  )
}

export function ErrorBoundary({ error }: { error: Error }) {
  return (
    <Document title='An error occurred.'>
      <Error reason={error.message} />
    </Document>
  )
}

export function CatchBoundary() {
  const caught = useCatch()
  return (
    <Document title='An error occurred.'>
      <Error statusCode={caught.status} reason={caught.statusText} />
    </Document>
  )
}
