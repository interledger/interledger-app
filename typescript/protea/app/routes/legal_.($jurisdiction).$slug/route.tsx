import { json } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'

import type { LoaderArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { StructuredText, toRemixMeta } from 'react-datocms'
import { fetchAndSanitizeHTML } from '~/lib/fetchAndSanitizeHTML'
import { getCurrentLegalPage } from '~/lib/marketing.server'

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.page.seoMeta),
    'twitter:url': `https://fynbos.app/legal${
      params.jurisdiction ? `/${params.jurisdiction}` : ''
    }/${params.slug}`,
    'og:url': `https://fynbos.app/legal${
      params.jurisdiction ? `/${params.jurisdiction}` : ''
    }/${params.slug}`
  }
}

export async function loader({ request, params, context }: LoaderArgs) {
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
    footer: (match) => match.data.footer
  }
}

export default function Page() {
  const { page, externalContent } = useLoaderData<typeof loader>()

  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible px-4 py-20 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        {externalContent && (
          <div
            style={{
              fontVariant: 'normal'
            }}
            className='prose prose-slate col-span-full dark:prose-invert prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider prose-pre:rounded-xl prose-pre:bg-slate-800 prose-pre:p-4 prose-pre:pb-6'
            dangerouslySetInnerHTML={{ __html: externalContent }}
          />
        )}
        {!externalContent && (
          <div
            style={{
              fontVariant: 'normal'
            }}
            className='prose prose-slate col-span-full dark:prose-invert prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider prose-pre:rounded-xl prose-pre:bg-slate-800 prose-pre:p-4 prose-pre:pb-6'
          >
            <h1>{page.title}</h1>
            <h3>Last updated: {page.updatedAt}</h3>
            {page && !externalContent && (
              <StructuredText data={page.body?.value} />
            )}
          </div>
        )}
      </section>
    </main>
  )
}
