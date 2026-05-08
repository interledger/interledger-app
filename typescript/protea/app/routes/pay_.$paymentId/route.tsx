import type { PlainMessage } from '@bufbuild/protobuf'
import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data, redirect } from 'react-router';
import { useLoaderData } from 'react-router';
import { useEffect } from 'react'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Alert,
  AlertBody,
  Card,
  CardContent,
  CardIcon,
  Icon,
  Layouts,
  Router
} from '~/components'
import type { FormattedLinkedAccount } from '~/data/accounts.server'
import { getLinkedAccountsForPayment } from '~/data/accounts.server'
import { getFeatures, getKycStatus } from '~/data/wallet.server'
import type {
  Features,
  Payment,
  PublicWalletInfo
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF } from '~/lib/csrf.server'
import type { ConnectError } from '~/lib/error.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { KycStatus, PaymentIdentityType, PaymentRequiredAction } from '~/lib/types'
import styles from '~/styles/flags.css?url'
import { Amount } from './Amount'
import { Confirm } from './Confirm'
import { confirmPaymentAction, updatePaymentAction } from './action.server';
import { envValue } from '~/env.server';

const IDENTITY_TYPE_TO_PLATFORM: Record<number, string> = {
  [PaymentIdentityType.Twitter]: 'twitter',
  [PaymentIdentityType.Slack]: 'slack',
  [PaymentIdentityType.WalletID]: 'domain',
  [PaymentIdentityType.WalletURL]: 'domain',
}

export async function loader({ request, params }: LoaderFunctionArgs) {
  let account: FormattedLinkedAccount
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PlainMessage<PublicWalletInfo>
  let features: Features | null = null
  let payment: PlainMessage<Payment> | ConnectError
  let phoneMask: string = ''

  const { kycStatus } = await getKycStatus(request)
  if (kycStatus != KycStatus.Approved)
    return redirect(href('/personal-details'))

  features = await getFeatures(request)

  payment = await grpc.getPayment(request, { id: params.paymentId })

  if (isConnectError(payment)) throw payment.errorResponse

  // This payment is already confirmed
  if (payment.state > 1) throw redirect(href('/pay'))

  const publicWalletInfoResponse = await grpc.getPublicWalletInfo(request, {
    walletAddress: payment.receiverWalletUrl
  })

  if (isConnectError(publicWalletInfoResponse)) {
    publicWalletInfo = {
      walletID: 'not-found',
      address: payment.receiverWalletUrl,
      shortAddress: payment.receiverIdentity,
      publicName: payment.receiverIdentity,
      identities: [
        {
          id: payment.receiverWalletUrl,
          wallet: '',
          platform: IDENTITY_TYPE_TO_PLATFORM[payment.receiverIdentityType] ?? '',
          identifier: payment.receiverIdentity,
          state: '',
          keyId: '',
          signature: '',
          signatureHash: '',
          proof: '',
          ctime: '',
          public: false,
          walletId: ''
        }
      ],
      canReceive: false
    }
  } else publicWalletInfo = publicWalletInfoResponse

  sendAccounts = await getLinkedAccountsForPayment(
    request,
    params.paymentId as string
  )

  if (payment.senderAccount) {
    const accountId = payment.senderAccount
    account = sendAccounts.find((acc) => acc.id == accountId)!
  } else {
    account = sendAccounts[0]
  }

  // Only load the phone mask if we require otp
  if (payment.requiredActions.includes(7)) {
    phoneMask = await getUserSession(request).then((v) => {
      const phone: string = v?.identity?.traits?.phone ?? ''
      const len = phone.length
      return phone.substring(len - 4, len).padStart(len, '*')
    })
  }

  return jsonWithCSRF(request, {
    features,
    account,
    sendAccounts,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: envValue("FYNBOS_ENV"),
    payment,
    requiresOTP: payment.requiredActions.includes(PaymentRequiredAction.OTP),
    PTIClientId: envValue("PTI_CLIENT_ID") || ''
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay', back: 'pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Pay'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const { features, sendAccounts } = useLoaderData<typeof loader>()
  const [step, reset] = usePayStore((state) => [state.step, state.reset])
  const [commandPaletteOpen, setCommandPaletteOpen] = useScaffoldStore(
    (state) => [state.commandPalletOpen, state.setCommandPalletOpen]
  )

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  useEffect(() => {
    if (commandPaletteOpen) {
      setCommandPaletteOpen(false)
    }
  }, [commandPaletteOpen, setCommandPaletteOpen])

  if (features && !features.sendEnabled)
    return (
      <>
        <Alert>
          <Icon>error</Icon>
          <AlertBody>
            Making payments in your state is currently unavailable. We're
            working to unlock all regions and will notify you when accessible.
            Thank you for your patience.
          </AlertBody>
        </Alert>
        <Card>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <CardIcon>
                <Icon>credit_card</Icon>
              </CardIcon>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>
                  Connect a card to receive payments.
                </p>
                <Router
                  prefetch='render'
                  className='text-sm font-medium text-primary'
                  to={href('/accounts')}
                >
                  Go to accounts page
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      </>
    )

  if (sendAccounts.length === 0)
    return (
      <Card>
        <CardContent>
          <div className='flex items-start space-x-4'>
            <CardIcon>
              <Icon>credit_card</Icon>
            </CardIcon>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                To send a payment, first connect a card.
              </p>
              <Router
                prefetch='render'
                className='text-sm font-medium text-primary'
                to={href('/accounts')}
              >
                Go to accounts page
              </Router>
            </div>
          </div>
        </CardContent>
      </Card>
    )

  return (
    <>
      {step === PayStep.AMOUNT && <Amount />}
      {step === PayStep.CONFIRM && <Confirm />}
    </>
  )
}

export async function action(args: ActionFunctionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string

  if (formName === 'updatePayment') {
    return updatePaymentAction(args)
  } else if (formName === 'confirmPayment') {
    return confirmPaymentAction(args)
  } else {
    throw data(
      { title: "Submitted a form that doesn't exist" },
      {
        status: 400
      }
    )
  }
}

