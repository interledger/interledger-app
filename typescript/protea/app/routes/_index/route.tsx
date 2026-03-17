import type { LinksFunction, LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts } from '~/components'
import {
  getFeatures,
  getKycStatus,
  getTransactionsWithPending,
  getWalletInfo
} from '~/data/wallet.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { hasUserSession } from '~/lib/kratos/session.server'
import { getPusherArgs } from '~/lib/pusher.server'
import flagStyles from '~/styles/flags.css'
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

export const links: LinksFunction = () => {
  return [{ rel: 'stylesheet', href: flagStyles }]
}

export async function loader(args: LoaderFunctionArgs) {
  const isUser = hasUserSession(args.request)

  if (isUser) {
    return appLoader(args)
  } else {
    return marketingLoader()
  }
}

export async function appLoader({ request }: LoaderFunctionArgs) {
  const [
    walletInfo,
    transactions,
    kycStatus,
    pusherArgs,
    features,
    balanceResponse
  ] = await Promise.all([
    getWalletInfo(request),
    getTransactionsWithPending(request, {
      pageSize: 3
    }),
    getKycStatus(request),
    getPusherArgs(request),
    getFeatures(request),
    grpc.getBalances(request, {})
  ])
  if (isConnectError(balanceResponse)) throw balanceResponse.error

  return json({
    isUser: true,
    walletInfo,
    transactions: transactions.transactions,
    kycStatus: kycStatus.kycStatus,
    pusherArgs,
    features,
    balances: balanceResponse.balances
  })
}

export async function marketingLoader() {
  return json({
    isUser: false
  })
}

export const handle: ApplicationProps = {
  layout: (match: UIMatch<typeof loader>) =>
    match.data.isUser ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: {
      title: 'Home'
    },
    fab: Fab.Pay
  }
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()
  if (isUser) return <AppPage />
  else return <MarketingPage />
}
