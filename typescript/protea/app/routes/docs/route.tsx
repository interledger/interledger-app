import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Outlet, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { getAllDocs } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  if (process.env.FYNBOS_ENV == 'prod')
    throw json(null, { status: 404, statusText: 'Not found' })

  const { allDocs } = await getAllDocs()
  if (!allDocs || !allDocs[0].slug)
    throw json(null, { status: 404, statusText: 'Not found' })

  const slug = allDocs[0].slug

  const url = new URL(request.url)

  if (url.pathname == '/docs') {
    return redirect(route('/docs/:slug', { slug }))
  }

  return json({ allDocs })
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

export function meta({ data, params }: any) {
  return {
    // ...toRemixMeta(data.allDocs.seoMeta),
    'twitter:url': 'https://fynbos.app/docs',
    'og:url': 'https://fynbos.app/docs'
  }
}

export default function Page() {
  const { allDocs } = useLoaderData<typeof loader>()

  // Instantiate all sections here for the side nav
  // useEffect(() => {
  //   console.log('ALL docs client', allDocs)
  // }, [allDocs])
  return <Outlet />
}
