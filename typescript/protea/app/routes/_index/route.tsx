import type { LoaderArgs } from '@remix-run/node'
import { defer } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts } from '~/components'
import type {
  FooterRecord,
  HomeRouteRecord
} from '~/generated/dato-cms-graphql'
import type { Identity } from '~/generated/protobuf-ts/backend/v1/backend'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import { getHomeRoute } from '~/lib/marketing.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'
import type { SnackbarType } from '~/lib/snackbar.server'
import { getSnackbar } from '~/lib/snackbar.server'
import type { PusherArgs } from '~/lib/usePusher'
import type { Transaction } from '~/lib/wallet.server'
import {
  getKycStatus,
  getLinkedAccounts,
  getLinkedIdentities,
  getTransactionsWithPending,
  getWalletPaymentPointer
} from '~/lib/wallet.server'
import { AppPage } from './app'
import { MarketingPage } from './marketing'

export enum KycStatus {
  Unknown = 0,
  InProgress = 1,
  DocumentsRequired = 2,
  Verified = 3,
  Suspended = 4
}

export async function loader({ request }: LoaderArgs) {
  const isUser = hasUserSession(request)

  let data = {
    isUser: isUser,
    homeRoute: {} as HomeRouteRecord,
    footer: {} as FooterRecord,
    pusherArgs: {} as PusherArgs,
    isSignupGated: IS_SIGNUP_GATED,
    firstName: '',
    paymentPointer: {
      url: '',
      asset: 'USD',
      assetScale: 2,
      alias: 'default',
      walletID: '',
      formatted: ''
    },
    hasCard: false,
    hasBank: false,
    transactions: [] as Transaction[],
    identities: {} as Record<string, Identity[]>,
    kycStatus: KycStatus.Unknown,
    snackbar: {
      message: ''
    } as SnackbarType
  }

  if (isUser) {
    const [
      session,
      paymentPointer,
      transactions,
      kycStatus,
      snackbar,
      pusherArgs,
      { bankAccounts, cardAccounts },
      identities
    ] = await Promise.all([
      getUserSession(request),
      getWalletPaymentPointer(request),
      getTransactionsWithPending(request, { pageSize: 3 }),
      getKycStatus(request),
      getSnackbar(request),
      getPusherArgs(request),
      getLinkedAccounts(request),
      getLinkedIdentities(request)
    ])

    data = {
      ...data,
      firstName: session.identity.traits.firstName,
      paymentPointer,
      transactions: transactions.transactions,
      kycStatus: kycStatus.kycStatus,
      snackbar,
      pusherArgs,
      hasCard: cardAccounts.length > 0,
      hasBank: bankAccounts.length > 0,
      identities
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
      actions: [{ type: 'shapes' }]
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
