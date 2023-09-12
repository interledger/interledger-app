import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getWalletRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'

export async function loader({ request }: LoaderArgs) {
  const { walletRoute, footer } = await getWalletRoute()
  return json({ walletRoute, footer })
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
    ...toRemixMeta(data.walletRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/wallet',
    'og:url': 'https://fynbos.app/wallet'
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
