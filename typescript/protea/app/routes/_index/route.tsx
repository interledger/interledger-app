import type { LoaderArgs } from '@remix-run/node'
import { defer } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts, WalletShapes } from '~/components'
import type {
  FooterRecord,
  HomeRouteRecord
} from '~/generated/dato-cms-graphql'
import type {
  Features,
  Transaction,
  WalletInfo
} from '~/generated/protobuf-ts/backend/v1/backend'
import { hasUserSession } from '~/lib/kratos.server'
import { getHomeRoute } from '~/lib/marketing.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'
import type { PusherArgs } from '~/lib/usePusher'
import {
  getFeatures,
  getKycStatus,
  getTransactionsWithPending,
  getWalletInfo
} from '~/lib/wallet.server'
import { AppPage } from './app'
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

export async function loader({ request }: LoaderArgs) {
  const isUser = hasUserSession(request)

  let data = {
    isUser: isUser,
    homeRoute: {} as HomeRouteRecord,
    footer: {} as FooterRecord,
    pusherArgs: {} as PusherArgs,
    isSignupGated: IS_SIGNUP_GATED,
    walletInfo: {} as WalletInfo,
    transactions: [] as Transaction[],
    kycStatus: KycStatus.Unknown,
    features: {} as Features
  }

  if (isUser) {
    const [walletInfo, transactions, kycStatus, pusherArgs, features] =
      await Promise.all([
        getWalletInfo(request),
        getTransactionsWithPending(request, { pageSize: 3 }),
        getKycStatus(request),
        getPusherArgs(request),
        getFeatures(request)
      ])

    data = {
      ...data,
      walletInfo,
      transactions: transactions.transactions,
      kycStatus: kycStatus.kycStatus,
      pusherArgs,
      features
    }
  } else {
    const { homeRoute, footer } = await getHomeRoute()
    data.homeRoute = homeRoute as HomeRouteRecord
    data.footer = footer as FooterRecord
  }
  return defer(data)
}

export const handle: ApplicationProps = {
  layout: (match) => (match.data.isUser ? Layouts.Wallet : Layouts.Marketing),
  scaffold: {
    header: {
      title: 'Home',
      actions: <WalletShapes />
    },
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.homeRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/',
    'og:url': 'https://fynbos.app/'
  }
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()
  if (isUser) return <AppPage />
  else return <MarketingPage />
}
