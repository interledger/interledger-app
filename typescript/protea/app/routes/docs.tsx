import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'

import type { LoaderDocsNav } from '~/components/Scaffold/Docs/useDocsStore'
import { getAllDocs } from '~/data/content.server'

export async function loader({ request }: LoaderArgs) {
  if (process.env.FYNBOS_ENV == 'prod')
    throw json(null, { status: 404, statusText: 'Not found' })

  const { allDocs } = await getAllDocs()
  if (!allDocs || (allDocs.length > 0 && !allDocs[0].slug))
    throw json(null, { status: 404, statusText: 'Not found' })

  const slug = allDocs[0].slug as string

  const url = new URL(request.url)

  if (url.pathname == '/docs' || url.pathname == '/docs/') {
    throw redirect(route('/docs/:slug', { slug }))
  }

  return json({ allDocs: allDocs as LoaderDocsNav[] })
}

export const handle: ApplicationProps = {
  layout: Layouts.Docs,
  scaffold: {
    header: {
      title: 'Docs'
    },
    footer: (match) => match.data.footer
  }
}

export function meta() {
  return {
    'twitter:url': 'https://fynbos.app/docs',
    'og:url': 'https://fynbos.app/docs'
  }
}

export default function Page() {
  return <Outlet />
}
