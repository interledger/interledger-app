import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts, MarketingPageWithSections } from '~/components'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { getWalletRoute } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  const { walletRoute, footer } = await getWalletRoute()
  return json({ walletRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.walletRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/about',
    'og:url': 'https://fynbos.app/about'
  }
}

export default function Page() {
  const { walletRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {walletRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
