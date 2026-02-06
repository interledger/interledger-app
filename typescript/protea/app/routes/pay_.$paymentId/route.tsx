import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'

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
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import type { ConnectError } from '~/lib/error.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { KycStatus } from '~/routes/_index/route'
import styles from '~/styles/flags.css'
import { Amount } from './Amount'
import { Confirm } from './Confirm'

export enum PaymentRequiredAction {
  Unknown,
  ThreeDS,
  SenderIdentifier,
  SenderAccount,
  ReceiverIdentifier,
  SenderAmount,
  ReceiverAmount,
  OTP,
  IPAddress
}
export enum PaymentIdentityType {
  Unknown,
  Twitter,
  WalletID,
  WalletURL,
  Slack,
  Sentinel // End of range value must be last, no need to public
}

function receiverIdentityTypeToPlatform(receiverIdentityType?: number): string {
  switch (receiverIdentityType) {
    case PaymentIdentityType.Twitter:
      return 'twitter'
    case PaymentIdentityType.Slack:
      return 'slack'
    case PaymentIdentityType.WalletID:
    case PaymentIdentityType.WalletURL:
      return 'wallet'
    default:
      return 'wallet'
  }
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
    return redirect(route('/personal-details'))

  features = await getFeatures(request)

  payment = await grpc.getPayment(request, { id: params.paymentId })

  if (isConnectError(payment)) throw payment.errorResponse

  // This payment is already confirmed
  if (payment.state > 1) throw redirect(route('/pay'))

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
          platform: receiverIdentityTypeToPlatform(
            payment.receiverIdentityType
          ),
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
      const len = v.identity.traits.phone.length
      return v.identity.traits.phone.substring(len - 4, len).padStart(len, '*')
    })
  }

  return jsonWithCSRF(request, {
    features,
    account,
    sendAccounts,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: process.env.FYNBOS_ENV,
    payment,
    requiresOTP: payment.requiredActions.includes(PaymentRequiredAction.OTP),
    PTIClientId: process.env.PTI_CLIENT_ID || ''
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
                  to={route('/accounts')}
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
                to={route('/accounts')}
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
    throw json(
      { title: "Submitted a form that doesn't exist" },
      {
        status: 400
      }
    )
  }
}

export async function confirmPaymentAction({
  request,
  params
}: ActionFunctionArgs) {
  const form = await request.formData()
  const serviceAgreement = form.get('serviceAgreement') as string
  // const otp = String(form.get('otp') || '')

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    otp: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    errors.serviceAgreement = 'You are required to authorize to continue.'
    return error(request, { errors })
  }

  const clientIpAddress = getClientIP(request)

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    ipAddress: clientIpAddress
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  response = await grpc.confirmPayment(request, {
    id: params.paymentId
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}

function stringToBigInt(amount: string) {
  if (amount == '') return BigInt(0)
  const dotIndex = amount.lastIndexOf('.')
  if (dotIndex > -1) {
    const amounts = amount.split('.')
    return BigInt(amounts[0] + amounts[1].slice(0, 2).padEnd(2, '0'))
  }
  return BigInt(parseFloat(amount) * 100)
}

export async function updatePaymentAction({
  request,
  params
}: ActionFunctionArgs) {
  const form = await request.formData()
  const send = String(form.get('send') || '')
  const receive = String(form.get('receive') || '')
  const note = String(form.get('note') || '')
  const accountId = String(form.get('accountId') || '')
  const sendCurrency = String(form.get('sendCurrency') || '')
  const receiveCurrency = String(form.get('receiveCurrency') || '')
  const intent = form.get('intent') as string

  const sendToSubmit = stringToBigInt(send)
  const receiveToSubmit = stringToBigInt(receive)

  const errors = {
    form: '',
    amount: '',
    linkedAccount: '',
    note: ''
  }

  if (intent == 'submit' && sendToSubmit == 0n) {
    errors.amount = 'Amount is required.'
    return error(request, { errors, payment: null, intent: '' })
  }

  const clientIpAddress = getClientIP(request)

  let senderAmount, receiverAmount
  if (send != '') {
    senderAmount = {
      amount: sendToSubmit,
      assetScale: 2,
      asset: sendCurrency
    }
  }
  if (receive != '') {
    receiverAmount = {
      amount: receiveToSubmit,
      assetScale: 2,
      asset: receiveCurrency
    }
  }

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    note,
    senderAccount: accountId,
    senderAmount,
    receiverAmount,
    ipAddress: clientIpAddress
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors, payment: null, intent: '' })
    }
    if (
      response.code == Code.FailedPrecondition &&
      response.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' &&
          violation.subject === 'insufficientFunds'
      ) > -1
    ) {
      return response.error({
        errors: { ...errors, amount: 'You have insufficient funds available.' },
        payment: null,
        intent: ''
      })
    }
    return response.error(
      { errors, payment: null, intent: '' },
      {},
      { action: 'Contact support' }
    )
  }

  return json({
    payment: response,
    intent,
    errors
  })
}
