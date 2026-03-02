import { KycStatus } from '~/lib/types'
import type { LinksFunction, LoaderFunctionArgs } from 'react-router';
import { data } from 'react-router';
import type { UIMatch } from 'react-router';
import { useLoaderData } from 'react-router';
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
import { hasUserSession } from '~/lib/kratos.server'
import { getPusherArgs } from '~/lib/pusher.server'
import flagStyles from '~/styles/flags.css?url'
import { AppPage } from './app'
import { MarketingPage } from './marketing'



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

async function appLoader({ request }: LoaderFunctionArgs) {
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

  return data({
    isUser: true,
    walletInfo,
    transactions: transactions.transactions,
    kycStatus: kycStatus.kycStatus,
    pusherArgs,
    features,
    balances: balanceResponse.balances
  })
}

async function marketingLoader() {
  return data({
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
