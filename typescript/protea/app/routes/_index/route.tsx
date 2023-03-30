import type { LoaderArgs } from '@remix-run/node'
import { defer } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Layouts } from '~/components'
import { getUserSession, hasUserSession } from '~/lib/kratos.server'
import type { Transaction } from '~/lib/wallet.server'
import {
  getKycStatus,
  getWalletPaymentPointer,
  getLinkedAccounts,
  getTransactionsWithPending
} from '~/lib/wallet.server'
import type { SnackbarType } from '~/lib/snackbar.server'
import { getSnackbar } from '~/lib/snackbar.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'
import type { PusherArgs } from '~/lib/usePusher'
import { getPusherArgs } from '~/lib/pusher.server'
import { MarketingPage } from './marketing-page'
import { AppPage, KycStatus } from './app-page'

export async function loader({ request }: LoaderArgs) {
  const isUser = hasUserSession(request)

  let data = {
    isUser: isUser,
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
          to: route('/'),
          // to: route('/linked-account/:type/widget', { type: 'card' }),
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
          to: route('/'),
          // to: route('/linked-account/:type/widget', { type: 'bank' }),
          text: 'Add bank account'
        },
        show: true
      }
    }
  }
  return defer(data)
}

export const handle = {
  layout: (isUser: boolean) =>
    isUser ? Layouts.WalletLayout : Layouts.LandingLayout
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()

  if (isUser) return <AppPage />
  else return <MarketingPage />
}
