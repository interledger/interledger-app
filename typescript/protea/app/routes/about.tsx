import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getAboutRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { datoMeta, mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const { aboutRoute, footer } = await getAboutRoute()
  return json({ aboutRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(
  ({ data, location }) => datoMeta(data?.aboutRoute?._seoMetaTags, location)
)

export default function Page() {
  const { aboutRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {aboutRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
