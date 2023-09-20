import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
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
import type { FormattedLinkedAccount } from '~/data/wallet.server'
import {
  getFeatures,
  getKycStatus,
  getLinkedAccounts,
  getPublicWalletInfo
} from '~/data/wallet.server'
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
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { KycStatus } from '~/routes/_index/route'
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
  Discord,
  Sentinel // End of range value must be last, no need to public
}

export async function loader({ request, params }: LoaderArgs) {
  let account: FormattedLinkedAccount | undefined
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let features: Features | null = null
  let payment: PlainMessage<Payment> | ConnectError
  let phoneMask: string = ''
  let payStep: PayStep = PayStep.AMOUNT

  const { kycStatus } = await getKycStatus(request)
  if (kycStatus != KycStatus.Approved)
    return redirect(route('/personal-details'))

  features = await getFeatures(request)

  payment = await grpc.getPayment(request, { id: params.paymentId })

  console.log('payment', payment)

  if (isConnectError(payment)) throw payment.errorResponse

  // This payment is already confirmed
  if (payment.state > 1) throw redirect(route('/pay'))

  publicWalletInfo = await getPublicWalletInfo(
    request,
    payment.receiverWalletUrl
  )

  const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)
  sendAccounts = [...cardAccounts, ...bankAccounts].filter((acc) => acc.canSend)

  if (payment.senderAccount) {
    const accountId = payment.senderAccount
    account = sendAccounts.find((acc) => acc.id == accountId)
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

  // If we have an amount and account, and we don't have outstanding requirements, then we can skip the amount step
  if (
    payment.senderAmount &&
    payment.senderAccount &&
    payment.requiredActions.findIndex(
      (ra) =>
        ra == PaymentRequiredAction.SenderAccount ||
        ra == PaymentRequiredAction.SenderAmount ||
        ra == PaymentRequiredAction.ReceiverAmount
    ) == -1
  ) {
    payStep = PayStep.CONFIRM
  }

  return jsonWithCSRF(request, {
    features,
    payStep,
    account,
    sendAccounts,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: process.env.FYNBOS_ENV,
    payment,
    requiresOTP: payment.requiredActions.includes(PaymentRequiredAction.OTP)
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay', back: 'pay' }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay'
  }
}

export default function Page() {
  const { payStep, features, sendAccounts, payment } =
    useLoaderData<typeof loader>()
  const [step, setStep, setAmount, reset] = usePayStore((state) => [
    state.step,
    state.setStep,
    state.setAmount,
    state.reset
  ])

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  useEffect(() => {
    if (step == PayStep.UNKNOWN) {
      setAmount(String(Number(payment.senderAmount?.amount) / 100 ?? ''))
      setStep(payStep)
    }
  }, [payStep, setStep, step])

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

export async function action(args: ActionArgs) {
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

export async function confirmPaymentAction({ request, params }: ActionArgs) {
  const form = await request.formData()
  const serviceAgreement = form.get('serviceAgreement') as string
  const otp = String(form.get('otp') || '')

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
    otp: otp,
    ipAddress: clientIpAddress
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  response = await grpc.confirmPayment(request, {
    id: params.paymentId
  })
  if (isConnectError(response)) {
    if (
      response.code === Code.FailedPrecondition &&
      response.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' && violation.subject === 'threeDS'
      ) > -1
    ) {
      return redirect(`/pay/3ds?paymentId=${params.paymentId}&init=`)
    }
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}

export async function updatePaymentAction({ request, params }: ActionArgs) {
  const form = await request.formData()
  const amount = form.get('amount') as string
  const note = String(form.get('note') || '')
  const accountId = String(form.get('accountId') || '')
  const intent = form.get('intent') as string
  const amountToSubmit = Math.floor(parseFloat(amount) * 100)

  const errors = {
    form: '',
    amount: '',
    linkedAccount: '',
    note: ''
  }

  if (!amountToSubmit) {
    errors.amount = 'Amount is required.'
    return error(request, { errors, payment: null, intent })
  }

  const clientIpAddress = getClientIP(request)

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    note,
    senderAccount: accountId,
    senderAmount: {
      amount: BigInt(amountToSubmit),
      assetScale: 2,
      asset: 'USD'
    },
    ipAddress: clientIpAddress
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors, payment: null, intent })
    }
    return response.error(
      { errors, payment: null, intent },
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
