import type {
  LinksFunction,
  LoaderArgs,
  V2_MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts } from '~/components'
import { getHomeRoute } from '~/data/content.server'
import {
  getFeatures,
  getKycStatus,
  getTransactionsWithPending,
  getWalletInfo
} from '~/data/wallet.server'
import { hasUserSession } from '~/lib/kratos.server'
import { datoMeta, mergeMeta } from '~/lib/meta'
import { getPusherArgs } from '~/lib/pusher.server'
import { AppPage } from './app'
import styles from './home.css'
import { MarketingPage } from './marketing'

export enum KycStatus {
  Unknown = 0,
  Pending = 1,
  DocumentsRequired = 2,
  Approved = 3,
  Denied = 4,
  InReview = 5,
  Level1 = 6,
  Level2 = 7
}

export const links: LinksFunction = () => {
  return [{ rel: 'stylesheet', href: styles }]
}

export async function loader(args: LoaderArgs) {
  const isUser = hasUserSession(args.request)

  if (isUser) {
    return appLoader(args)
  } else {
    return marketingLoader()
  }
}

export async function appLoader({ request }: LoaderArgs) {
  const [walletInfo, transactions, kycStatus, pusherArgs, features] =
    await Promise.all([
      getWalletInfo(request),
      getTransactionsWithPending(request, {
        pageSize: 3
      }),
      getKycStatus(request),
      getPusherArgs(request),
      getFeatures(request)
    ])

  return json({
    isUser: true,
    walletInfo,
    transactions: transactions.transactions,
    kycStatus: kycStatus.kycStatus,
    pusherArgs,
    features
  })
}

export async function marketingLoader() {
  const { homeRoute, footer } = await getHomeRoute()

  return json({
    isUser: false,
    homeRoute,
    footer
  })
}

export const handle: ApplicationProps = {
  layout: (match) => (match.data.isUser ? Layouts.Wallet : Layouts.Marketing),
  scaffold: {
    header: {
      title: 'Home'
    },
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export const meta: V2_MetaFunction<typeof loader> = mergeMeta(
  ({ data }) => datoMeta(data?.homeRoute?.seoMeta),
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
  const { isUser } = useLoaderData<typeof loader>()
  if (isUser) return <AppPage />
  else return <MarketingPage />
}
