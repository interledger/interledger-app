import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSearchParams } from '@remix-run/react'
import { DateTime } from 'luxon'
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
import type {
  Features,
  PublicWalletInfo,
  SearchResult
} from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'
import {
  getFeatures,
  getKycStatus,
  getLinkedAccounts,
  getPublicWalletInfo,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
import { Amount, AmountWithOpenPayments } from '~/routes/pay/Amount'
import { Confirm } from '~/routes/pay/Confirm'
import { Search } from '~/routes/pay/Search'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  let results: SearchResult[] = []
  let address: SearchResult | null = null
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''
  let features: Features | null = null

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
    const response = await grpcClient
      .searchWallets(
        { term },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)

    if (!isGrpcError(response)) {
      results = response.response.results
    }
  }

  const walletAddressParam = url.searchParams.get('address')
  if (walletAddressParam) {
    // TODO: refactor this to return a SearchResult
    const response = await grpcClient
      .getPaymentAddress(
        { address: walletAddressParam },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (!isGrpcError(response)) {
      address = {
        walletID: '',
        walletUrl: response.response.walletUrl,
        canSend: response.response.canSendToAddress,
        identifier: response.response.handle,
        identifierType: response.response.type,
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
    fynbosEnv: process.env.FYNBOS_ENV
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
  const { address, features, sendAccounts, fynbosEnv } =
    useLoaderData<typeof loader>()
  const [params, setSearchParams] = useSearchParams()
  const [setAddress, step, setStep, reset] = usePayStore((state) => [
    state.setAddress,
    state.step,
    state.setStep,
    state.reset
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
      {step === PayStep.AMOUNT && fynbosEnv === 'prod' && (
        <AmountWithOpenPayments />
      )}
      {step === PayStep.AMOUNT && fynbosEnv !== 'prod' && <Amount />}
      {step === PayStep.CONFIRM && <Confirm />}
    </>
  )
}

type quoteLimitError =
  | 'Failed precondition: LimitTransaction'
  | 'Failed precondition: LimitDaily'
  | 'Failed precondition: LimitMonthly'
  | 'Failed precondition: Limit6Monthly'

// The field names given by the backend for field violations
type fieldErrorsMap = 'url' | 'amount'

function mapper(field: fieldErrorsMap): 'address' | 'amount' | null {
  switch (field) {
    case 'url':
      return 'address'
    case 'amount':
      return 'amount'
    default:
      return null
  }
}

export async function action(args: ActionArgs) {
  const formName = (await args.request.clone().formData()).get(
    'formName'
  ) as string
  if (formName === 'quote') {
    return openpaymentsQuoteAction(args)
  } else if (formName === 'confirm') {
    return openPaymentsConfirmAction(args)
  } else if (formName === 'updatePayment') {
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
  const fieldErrors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to authorize to continue.'
    return error(request, { errors: { ...fieldErrors } })
  }

  if (otp) {
    const rpc = await grpcClient
      .updatePayment(
        {
          id: paymentId,
          otp: otp
        },
        {
          meta: { cookies: String(request.headers.get('cookie')) }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(rpc)) {
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
    }
  }

  // TODO: Bank payments should just create outgoing payment here
  return redirect(`/pay/3ds?paymentId=${paymentId}&init=`)
}

export async function updatePaymentAction({ request }: ActionArgs) {
  const form = await request.formData()
  const paymentId = form.get('paymentId') as string
  const amount = form.get('amount') as string
  const note = form.get('note') as string
  const accountId = form.get('accountId') as string
  const type = form.get('type') as string
  const walletUrl = form.get('walletUrl') as string
  const amountToSubmit = String(Math.floor(parseFloat(amount) * 100))
  const fieldErrors = {
    form: '',
    amount: '',
    address: '',
    linkedAccount: '',
    note: ''
  }

  if (amountToSubmit == 'NaN') {
    fieldErrors.amount = 'Amount is required.'
    return error(request, { errors: { ...fieldErrors }, payment: null, type })
  }

  const clientIpAddress = getClientIP(request)
  if (!paymentId) {
    let rpc = await grpcClient
      .createPayment(
        {
          receiverIdentity: walletUrl,
          receiverIdentityType: 3,
          note,
          senderAccount: accountId,
          senderAmount: {
            amount: amountToSubmit,
            assetScale: 2,
            asset: 'USD'
          },
          ipAddress: clientIpAddress
        },
        {
          meta: { cookies: String(request.headers.get('cookie')) }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(rpc)) {
      if (rpc.code == Code.INVALID_ARGUMENT) {
        for (let violation of (rpc as GrpcError).details[0].fieldViolations) {
          const field = mapper(violation.field as fieldErrorsMap)
          if (field != null) fieldErrors[field] = violation.description
        }
        return error(request, { errors: fieldErrors, payment: null, type })
      } else if (rpc.code == Code.FAILED_PRECONDITION) {
        switch (rpc.message as quoteLimitError) {
          case 'Failed precondition: LimitTransaction':
            fieldErrors['amount'] = 'Exceeds per transaction limit.'
            break
          case 'Failed precondition: LimitDaily':
            fieldErrors['amount'] = 'Exceeds daily limit.'
            break
          case 'Failed precondition: LimitMonthly':
            fieldErrors['amount'] = 'Exceeds monthly limit.'
            break
          case 'Failed precondition: Limit6Monthly':
            fieldErrors['amount'] = 'Exceeds rolling 6 month limit.'
            break
          default:
            fieldErrors['amount'] = 'Exceeds account limit.'
        }
        return error(request, { errors: fieldErrors, payment: null, type })
      } else
        return error(
          request,
          { errors: fieldErrors, payment: null, type },
          { action: 'Contact support' }
        )
    }

    return json({
      payment: rpc.response,
      type,
      errors: fieldErrors
    })
  }

  let rpc = await grpcClient
    .updatePayment(
      {
        id: paymentId,
        receiverIdentity: walletUrl,
        receiverIdentityType: 3,
        note,
        senderAccount: accountId,
        senderAmount: {
          amount: amountToSubmit,
          assetScale: 2,
          asset: 'USD'
        },
        ipAddress: clientIpAddress
      },
      {
        meta: { cookies: String(request.headers.get('cookie')) }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    if (rpc.code == Code.INVALID_ARGUMENT) {
      for (let violation of (rpc as GrpcError).details[0].fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return error(request, { errors: fieldErrors, payment: null, type })
    } else if (rpc.code == Code.FAILED_PRECONDITION) {
      switch (rpc.message as quoteLimitError) {
        case 'Failed precondition: LimitTransaction':
          fieldErrors['amount'] = 'Exceeds per transaction limit.'
          break
        case 'Failed precondition: LimitDaily':
          fieldErrors['amount'] = 'Exceeds daily limit.'
          break
        case 'Failed precondition: LimitMonthly':
          fieldErrors['amount'] = 'Exceeds monthly limit.'
          break
        case 'Failed precondition: Limit6Monthly':
          fieldErrors['amount'] = 'Exceeds rolling 6 month limit.'
          break
        default:
          fieldErrors['amount'] = 'Exceeds account limit.'
      }
      return error(request, { errors: fieldErrors, payment: null, type })
    } else
      return error(
        request,
        { errors: fieldErrors, payment: null, type },
        { action: 'Contact support' }
      )
  }

  return json({
    payment: rpc.response,
    type,
    errors: fieldErrors
  })
}

async function openpaymentsQuoteAction({ request }: ActionArgs) {
  const form = await request.formData()
  const amount = form.get('amount') as string
  const note = form.get('note') as string
  const walletUrl = form.get('walletUrl') as string
  const accountId = form.get('accountId') as string
  const type = form.get('type') as string
  const identity = form.get('identity') as string
  const identityType = form.get('identityType') as string
  const amountToSubmit = String(Math.floor(parseFloat(amount) * 100))

  const expiresAt = {
    seconds: `${Math.floor(DateTime.now().plus({ hour: 1 }).toSeconds())}`,
    nanos: 0
  }

  const fieldErrors = {
    form: '',
    amount: '',
    address: '',
    linkedAccount: '',
    note: ''
  }

  if (amountToSubmit == 'NaN') {
    fieldErrors.amount = 'Amount is required.'
    return error(request, { errors: { ...fieldErrors } })
  }

  let walletInfo = await getWalletInfo(request)

  const response = await openPaymentsClient
    .createQuote(
      {
        sendPaymentPointer: walletInfo.url,
        receivePaymentPointer: walletUrl,
        description: note,
        amount: {
          amount: amountToSubmit,
          asset: 'USD',
          assetScale: 2
        },
        expiresAt,
        sendLinkedAccount: accountId,
        identityType,
        identity
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == Code.INVALID_ARGUMENT) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return error(request, { errors: { ...fieldErrors } })
    } else if (response.code == Code.FAILED_PRECONDITION) {
      switch (response.message as quoteLimitError) {
        case 'Failed precondition: LimitTransaction':
          fieldErrors['amount'] = 'Exceeds per transaction limit.'
          break
        case 'Failed precondition: LimitDaily':
          fieldErrors['amount'] = 'Exceeds daily limit.'
          break
        case 'Failed precondition: LimitMonthly':
          fieldErrors['amount'] = 'Exceeds monthly limit.'
          break
        case 'Failed precondition: Limit6Monthly':
          fieldErrors['amount'] = 'Exceeds rolling 6 month limit.'
          break
        default:
          fieldErrors['amount'] = 'Exceeds account limit.'
      }
      return error(request, { errors: { ...fieldErrors } })
    } else
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
  }

  const data = {
    errors: { ...fieldErrors },
    quoteId: response.response.id,
    requiresOTP: response.response.requiresOTP,
    type
  }

  return json(data)
}

async function openPaymentsConfirmAction({ request }: ActionArgs) {
  const form = await request.formData()
  await validateCSRFToken(request, form)
  const serviceAgreement = form.get('serviceAgreement') as string
  const quoteId = form.get('quoteId') as string
  const otp = form.get('otp') as string
  const fieldErrors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to authorize to continue.'
    return error(request, { errors: { ...fieldErrors } })
  }

  const quoteIdParam = quoteId.split('/').at(-1)

  if (otp) {
    const response = await openPaymentsClient
      .setQuoteOTP(
        {
          quoteID: quoteIdParam as string,
          otp
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(response)) {
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
    }
  }

  // TODO: Bank payments should just create outgoing payment here
  return redirect(`/pay/3ds?quoteId=${quoteIdParam}&init=`)
}
