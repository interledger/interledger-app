import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { getAllDocs } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  if (process.env.FYNBOS_ENV == 'prod')
    throw json(null, { status: 404, statusText: 'Not found' })
  // TODO if pathname == '/docs' redirect to '/docs/getting-started'
  const { allDocs } = await getAllDocs()
  console.log('allDocs', allDocs)
  return json({ allDocs })
}

export const handle: ApplicationProps = {
  layout: Layouts.Docs,
  scaffold: {
    header: {
      title: (match) => match.data.allDocs[0].title
    },
    nav: (match) => match.data.allDocs,
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
  return (
    <>
      <div className='h-16 w-full bg-blue-500'></div>
      {allDocs?.map((section) => (
        <h1>test</h1>
        // <MarketingPageWithSections
        //   key={section.id}
        //   section={section as SectionRecord}
        // />
      ))}
    </>
  )
}
