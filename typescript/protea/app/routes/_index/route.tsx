import type { LoaderArgs } from '@remix-run/node'
import { defer } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts } from '~/components'
import type { FooterRecord, HomepageRecord } from '~/generated/dato-cms-graphql'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import { getHomePage } from '~/lib/marketing.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'
import type { SnackbarType } from '~/lib/snackbar.server'
import { getSnackbar } from '~/lib/snackbar.server'
import type { PusherArgs } from '~/lib/usePusher'
import type { Transaction } from '~/lib/wallet.server'
import {
  getKycStatus,
  getLinkedAccounts,
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
    homepage: {} as HomepageRecord,
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
    balance: '' as unknown as Promise<string>,
    transactions: [] as Transaction[],
    kycStatus: KycStatus.Unknown,
    canTopUp: false,
    canWithdraw: false,
    nextStep: {
      title: '',
      icon: '',
      action: { to: '', text: '' },
      show: false
    },
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
      linkedAccounts,
      snackbar,
      pusherArgs
    ] = await Promise.all([
      getUserSession(request),
      getWalletPaymentPointer(request),
      getTransactionsWithPending(request, { pageSize: 3 }),
      getKycStatus(request),
      getLinkedAccounts(request),
      getSnackbar(request),
      getPusherArgs(request)
    ])

    data = {
      ...data,
      firstName: session.identity.traits.firstName,
      paymentPointer,
      transactions: transactions.transactions,
      kycStatus: kycStatus.kycStatus,
      canTopUp: linkedAccounts.canTopUp,
      canWithdraw: linkedAccounts.canWithdraw,
      snackbar,
      pusherArgs
    }

    /**
     * Next Step state machine
     * 1. Activate PP - KYCStatus.Unknown
     * 2. Add debit - KYCStatus.Verified + !hasTransactions + !canTopUp
     * 3. Add bank - KYCStatus.Verified + hasTransactions + !canWithdraw
     */
    if (data.kycStatus == KycStatus.Unknown) {
      data.nextStep = {
        title:
          'Your payment pointer is reserved, we just need a few more details to activate it.',
        icon: 'attach_money',
        action: {
          to: route('/personal-details'),
          text: 'Activate payment pointer'
        },
        show: true
      }
    } else if (
      data.kycStatus == KycStatus.Verified &&
      transactions.transactions.length == 0 &&
      !data.canTopUp
    ) {
      data.nextStep = {
        title:
          'Add a debit card to easily send payments or top up your cash balance.',
        icon: 'add_card',
        action: {
          to: route('/link-account/card'),
          text: 'Add a debit card'
        },
        show: true
      }
    } else if (
      data.kycStatus == KycStatus.Verified &&
      transactions.transactions.length > 0 &&
      !data.canWithdraw
    ) {
      data.nextStep = {
        title:
          'Add a bank account to securely withdraw from your cash balance at any time.',
        icon: 'account_balance',
        action: {
          to: route('/link-account/bank'),
          text: 'Add bank account'
        },
        show: true
      }
    }
  } else {
    const { homepage, footer } = await getHomePage()
    data.homepage = homepage as HomepageRecord
    data.footer = footer as FooterRecord
  }
  return defer(data)
}

export const handle: ApplicationProps = {
  layout: (match) => (match.data.isUser ? Layouts.Wallet : Layouts.Marketing),
  scaffold: {
    header: {},
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.homepage.seoMeta),
    'twitter:url': 'https://fynbos.app/',
    'og:url': 'https://fynbos.app/'
  }
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()
  if (isUser) return <AppPage />
  else return <MarketingPage />
}
