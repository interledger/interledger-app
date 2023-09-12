import { Code } from '@bufbuild/connect'
import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSearchParams } from '@remix-run/react'
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
  getLinkedAccount,
  getLinkedAccounts,
  getPublicWalletInfo
} from '~/data/wallet.server'
import type {
  Features,
  Payment,
  PublicWalletInfo,
  SearchResult
} from '~/generated/connect/backend/v1/backend_pb'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { KycStatus } from '~/routes/_index/route'
import { Amount } from '~/routes/pay/Amount'
import { Confirm } from '~/routes/pay/Confirm'
import { Search } from '~/routes/pay/Search'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const paymentId = url.searchParams.get('paymentId')
  if (paymentId) {
    return loadPayment(request, paymentId)
  }

  let results: PlainMessage<SearchResult>[] = []
  let address: PlainMessage<SearchResult> | null = null
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''
  let features: Features | null = null
  let payment: Payment | null = null

  if (url.search == '') {
    const { kycStatus } = await getKycStatus(request)
    if (kycStatus != KycStatus.Approved)
      return redirect(route('/personal-details'))

    features = await getFeatures(request)

    const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)
    sendAccounts = [...cardAccounts, ...bankAccounts].filter(
      (acc) => acc.canSend
    )
  }

  const accounts = url.searchParams.get('accounts')
  if (accounts) {
    const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)

    sendAccounts = [...cardAccounts, ...bankAccounts].filter(
      (acc) => acc.canSend
    )

    if (sendAccounts.length === 0) {
      return redirectWithSnackbar(request, route('/accounts'), {
        message: 'You need a connected account to make a payment.',
        icon: 'close'
      })
    }
  }

  const term = url.searchParams.get('term')
  if (term) {
    const response = await grpc.searchWallets(request, { term })

    if (!isConnectError(response)) results = response.results
  }

  const walletAddressParam = url.searchParams.get('address')
  if (walletAddressParam) {
    // TODO: refactor this to return a SearchResult
    const response = await grpc.getPaymentAddress(request, {
      address: walletAddressParam
    })

    if (!isConnectError(response)) {
      address = {
        walletID: '',
        walletUrl: response.walletUrl,
        canSend: response.canSendToAddress,
        identifier: response.handle,
        identifierType: response.type,
        subResults: []
      }
    }
  }

  const walletUrl = url.searchParams.get('walletUrl')
  if (walletUrl) {
    publicWalletInfo = await getPublicWalletInfo(request, walletUrl)
  }

  const phone = url.searchParams.get('phone')
  if (phone) {
    phoneMask = await getUserSession(request).then((v) => {
      const len = v.identity.traits.phone.length
      return v.identity.traits.phone.substring(len - 4, len).padStart(len, '*')
    })
  }

  return jsonWithCSRF(request, {
    features,
    results,
    address,
    sendAccounts,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: process.env.FYNBOS_ENV,
    payment
  })
}

async function loadPayment(request: Request, id: string) {
  let results: SearchResult[] = []
  let address: SearchResult | null = null
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''
  let features: Features | null = null

  let payment = await grpc.getPayment(request, { id })

  if (isConnectError(payment)) throw payment.errorResponse

  address = null // to avoid useEffect race condition

  publicWalletInfo = await getPublicWalletInfo(
    request,
    payment.receiverWalletUrl
  )
  sendAccounts[0] = await getLinkedAccount(request, payment.senderAccount)

  return jsonWithCSRF(request, {
    features,
    results,
    address,
    sendAccounts,
    phoneMask,
    publicWalletInfo,
    fynbosEnv: process.env.FYNBOS_ENV,
    payment
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

export function getPaymentIdentityType(identityEnum: number) {
  switch (identityEnum) {
    case 1:
      return 'Twitter'
    case 4:
      return 'Slack'
    case 5:
      return 'Discord'
    default:
      return 'wallet'
  }
}

export default function Page() {
  const { address, features, sendAccounts, payment, publicWalletInfo } =
    useLoaderData<typeof loader>()
  const [params, setSearchParams] = useSearchParams()
  const [
    setAddress,
    step,
    setStep,
    reset,
    setPayment,
    setNote,
    setAccount,
    setPublicWalletInfo,
    setAmount
  ] = usePayStore((state) => [
    state.setAddress,
    state.step,
    state.setStep,
    state.reset,
    state.setPayment,
    state.setNote,
    state.setAccount,
    state.setPublicWalletInfo,
    state.setAmount
  ])

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  // Check if there is already a user selected from me page
  useEffect(() => {
    if (
      address &&
      address.canSend &&
      params.get('start') == String(PayStep.AMOUNT)
    ) {
      setAddress(address)
      setStep(PayStep.AMOUNT)
      setSearchParams({}, { replace: true })
    }
  }, [address, params, setAddress, setSearchParams, setStep])

  useEffect(() => {
    if (payment) {
      const requiresOTP = 7
      setPayment(payment.id, payment.requiredActions.includes(requiresOTP))
      setAddress({
        identifier: payment.receiverIdentity,
        identifierType: getPaymentIdentityType(payment.receiverIdentityType),
        walletID: '',
        walletUrl: '',
        canSend: true,
        subResults: []
      })
      setNote(payment.note)

      if (sendAccounts && sendAccounts.length > 0) {
        setAccount(sendAccounts[0])
      }

      let floatAmount = parseInt(payment.senderAmount?.amount || '0') / 100
      setAmount(floatAmount.toFixed(2))
      setSearchParams({}, { replace: true })
    }

    if (params.get('start') == String(PayStep.CONFIRM)) {
      setStep(PayStep.CONFIRM)
    }
  }, [
    payment,
    params,
    setSearchParams,
    setStep,
    sendAccounts,
    setAmount,
    setNote,
    setAddress,
    setPayment,
    setAccount
  ])

  useEffect(() => {
    if (publicWalletInfo) {
      setPublicWalletInfo(publicWalletInfo)
    }
  }, [publicWalletInfo, setPublicWalletInfo])

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
      {step === PayStep.SEARCH && <Search />}
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

export async function confirmPaymentAction({ request }: ActionArgs) {
  const form = await request.formData()
  await validateCSRFToken(request, form)
  const serviceAgreement = form.get('serviceAgreement') as string
  const paymentId = form.get('paymentId') as string
  const otp = form.get('otp') as string

  const errors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    errors.serviceAgreement = 'You are required to authorize to continue.'
    return error(request, { errors })
  }

  if (otp) {
    const response = await grpc.updatePayment(request, {
      id: paymentId,
      otp: otp
    })
    if (isConnectError(response)) {
      return response.error({ errors }, {}, { action: 'Contact support' })
    }
  }

  const response = await grpc.confirmPayment(request, {
    id: paymentId
  })

  if (isConnectError(response)) {
    if (
      response.code === Code.FailedPrecondition &&
      response.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' && violation.subject === 'threeDS'
      ) > -1
    ) {
      return redirect(`/pay/3ds?paymentId=${paymentId}&init=`)
    }
    return error(request, { errors }, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}

export async function updatePaymentAction({ request }: ActionArgs) {
  const form = await request.formData()
  const paymentId = form.get('paymentId') as string
  const amount = form.get('amount') as string
  const note = form.get('note') as string
  const accountId = form.get('accountId') as string
  const type = form.get('type') as string
  const walletUrl = form.get('walletUrl') as string
  const amountToSubmit = Math.floor(parseFloat(amount) * 100)

  const errors = {
    form: '',
    amount: '',
    address: '',
    linkedAccount: '',
    note: ''
  }

  if (!amountToSubmit) {
    errors.amount = 'Amount is required.'
    return error(request, { errors, payment: null, type })
  }

  const clientIpAddress = getClientIP(request)
  if (!paymentId) {
    let response = await grpc.createPayment(request, {
      receiverIdentity: walletUrl,
      receiverIdentityType: 3,
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
        return response.error({ errors, payment: null, type })
      }
      return response.error(
        { errors, payment: null, type },
        {},
        { action: 'Contact support' }
      )
    }

    return json({
      payment: response,
      type,
      errors
    })
  }

  let response = await grpc.updatePayment(request, {
    id: paymentId,
    receiverIdentity: walletUrl,
    receiverIdentityType: 3,
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
      return response.error({ errors, payment: null, type })
    }
    return response.error(
      { errors, payment: null, type },
      {},
      { action: 'Contact support' }
    )
  }

  return json({
    payment: response,
    type,
    errors
  })
}
