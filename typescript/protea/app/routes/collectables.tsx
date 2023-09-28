import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getCollectablesRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { datoMeta, mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const { collectablesRoute, footer } = await getCollectablesRoute()
  return json({ collectablesRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(
  ({ data, location }) =>
    datoMeta(data?.collectablesRoute?._seoMetaTags, location)
)

export default function Page() {
  const { collectablesRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {collectablesRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
