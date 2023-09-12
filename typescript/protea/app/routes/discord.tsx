import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getDiscordRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'

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

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.discordRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/discord',
    'og:url': 'https://fynbos.app/discord'
  }
}

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
