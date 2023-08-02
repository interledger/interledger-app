import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSearchParams } from '@remix-run/react'
import { DateTime } from 'luxon'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import type {
  PublicWalletInfo,
  SearchResult
} from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { getUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'
import {
  getKycStatus,
  getLinkedAccounts,
  getPublicWalletInfo,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
import { Amount } from '~/routes/pay/Amount'
import { Confirm } from '~/routes/pay/Confirm'
import { Search } from '~/routes/pay/Search'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  let results: SearchResult[] = []
  let address: SearchResult | null = null
  let sendAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let phoneMask: string = ''

  if (url.search == '') {
    const { kycStatus } = await getKycStatus(request)
    if (kycStatus != KycStatus.Approved)
      return redirect(route('/personal-details'))
  }

  const accounts = url.searchParams.get('accounts')
  if (accounts) {
    const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)

    sendAccounts = [...cardAccounts, ...bankAccounts].filter(
      (acc) => acc.canSend
    )

    if (sendAccounts.length === 0) {
      await flashSnackbar(request, {
        message: 'You need a connected account to make a payment.',
        icon: 'close'
      })
      return redirect(route('/accounts'))
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
  const { address } = useLoaderData<typeof loader>()
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
      setSearchParams({})
    }
  }, [address, params, setAddress, setSearchParams, setStep])

  return (
    <>
      {step === PayStep.SEARCH && <Search />}
      {step === PayStep.AMOUNT && <Amount />}
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

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const formName = (await form.get('formName')) as string

  if (formName === 'quote') {
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
      return json({ errors: { ...fieldErrors } }, { status: 400 })
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
        return json({ errors: { ...fieldErrors } }, { status: 400 })
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
        return json({ errors: { ...fieldErrors } }, { status: 400 })
      } else throw json({}, httpMapping(response.code))
    }

    const data = {
      errors: { ...fieldErrors },
      quoteId: response.response.id,
      requiresOTP: response.response.requiresOTP,
      type
    }

    return json(data)
  }

  if (formName === 'confirm') {
    await validateCSRFToken(request, form)
    const serviceAgreement = form.get('serviceAgreement') as string
    const quoteId = form.get('quoteId') as string
    const otp = form.get('otp') as string
    const fieldErrors = {
      form: '',
      serviceAgreement: ''
    }

    if (serviceAgreement == null) {
      fieldErrors.serviceAgreement =
        'You are required to authorize to continue.'
      return json(
        {
          errors: {
            ...fieldErrors
          }
        },
        { status: 400 }
      )
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
        throw json({}, httpMapping(response.code))
      }
    }

    // TODO: Bank payments should just create outgoing payment here
    return redirect(`/pay/3ds?quoteId=${quoteIdParam}&init=`)
  }

  throw json(
    { title: "Submitted a form that doesn't exist" },
    {
      status: 400
    }
  )
}
