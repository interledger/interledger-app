import type { LoaderArgs, V2_MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getDiscordRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { datoMeta, mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderArgs) {
  const { discordRoute, footer } = await getDiscordRoute()
  return json({ discordRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match) => match.data.footer
  }
}

export const meta: V2_MetaFunction<typeof loader> = mergeMeta(
  ({ data }) => datoMeta(data?.discordRoute?.seoMeta),
  ({ location }) => [
    {
      name: 'og:url',
      content: `https://fynbos.app${location.pathname}`
    },
    {
      name: 'twitter:url',
      content: `https://fynbos.app${location.pathname}`
    }
  ]
)

export default function Page() {
  const { discordRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {discordRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
