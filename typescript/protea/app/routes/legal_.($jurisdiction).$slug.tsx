import { json } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'

import type { LoaderFunctionArgs } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { StructuredText } from 'react-datocms'
import { Prose } from '~/components/Content'
import { getCurrentLegalPage } from '~/data/content.server'
import { fetchAndSanitizeHTML } from '~/lib/fetchAndSanitizeHTML.server'

export async function loader({ params }: LoaderFunctionArgs) {
  const { legalPage, footer } = await getCurrentLegalPage({
    filter: {
      slug: { eq: params.slug },
      jurisdiction: { eq: params.jurisdiction ?? 'global' }
    }
  })

  if (legalPage === null)
    throw json(null, { status: 404, statusText: 'Not Found' })

  // TODO decide if we want to route to the global version if there isn't a jurisdiction specific one.

  let externalContent
  if (legalPage?.external) {
    externalContent = await fetchAndSanitizeHTML(legalPage.external)
  }

  return json({
    footer,
    page: {
      ...legalPage,
      updatedAt: DateTime.fromISO(legalPage?._publishedAt).toFormat(
        'dd MMM yyyy'
      )
    },
    externalContent
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer
  }
}

export default function Page() {
  const { page, externalContent } = useLoaderData<typeof loader>()

  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 py-20 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        {externalContent && (
          <Prose
            style={{
              fontVariant: 'normal'
            }}
            className='col-span-full'
          >
            <div dangerouslySetInnerHTML={{ __html: externalContent }} />
          </Prose>
        )}
        {!externalContent && (
          <Prose
            style={{
              fontVariant: 'normal'
            }}
            className='col-span-full'
          >
            <h1>{page.title}</h1>
            <h3>Last updated: {page.updatedAt}</h3>
            {page && !externalContent && (
              <StructuredText data={page.body?.value} />
            )}
          </Prose>
        )}
      </section>
    </main>
  )
}
