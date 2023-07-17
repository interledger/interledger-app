import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useFetcher } from '@remix-run/react'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import { v4 } from 'uuid'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import type {
  Init3DSResponse,
  PublicWalletInfo,
  SearchResult
} from '~/generated/protobuf-ts/backend/v1/backend'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import type { FormattedLinkedAccount } from '~/lib/wallet.server'
import {
  getKycStatus,
  getLinkedAccounts,
  getPublicWalletInfo,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'
import { Amount } from '~/routes/pay/Amount'
import { Search } from '~/routes/pay/Search'
import { PayStep, useStore } from '~/store'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  let results: SearchResult[] = []
  let linkedAccounts: FormattedLinkedAccount[] = []
  let publicWalletInfo: PublicWalletInfo | null = null
  let init3DS: Init3DSResponse | null = null

  if (url.search == '') {
    const { kycStatus } = await getKycStatus(request)
    if (kycStatus != KycStatus.Verified)
      return redirect(route('/personal-details'))

    const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)

    linkedAccounts = [...cardAccounts, ...bankAccounts].filter(
      (acc) => acc.canSend
    )

    if (linkedAccounts.length === 0) {
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
      results = response.response.results.filter((v) => v.canSend)
    }
  }

  const walletUrl = url.searchParams.get('walletUrl')
  if (walletUrl) {
    publicWalletInfo = await getPublicWalletInfo(request, walletUrl)
  }

  const idempotencyKey = url.searchParams.get('idempotencyKey')
  const quoteID = url.searchParams.get('quoteID')
  if (idempotencyKey && quoteID) {
    let threeDSInit = await grpcClient
      .init3DS(
        {
          idempotencyKey,
          quoteID
        },
        {
          meta: {
            cookies: String(request.headers.get('cookie'))
          }
        }
      )
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(threeDSInit)) throw json({}, httpMapping(threeDSInit.code))

    init3DS = threeDSInit.response
  }

  return json({
    results,
    linkedAccounts,
    publicWalletInfo,
    init3DS,
    fynbosEnv: process.env.FYNBOS_ENV
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay', back: route('/') }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay'
  }
}

export default function Page() {
  const fetcher = useFetcher()

  const step = useStore((state) => state.step)

  return (
    <>
      <fetcher.Form
        id='pay-form'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <Form
        id='pay-address'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      {step === PayStep.SEARCH && <Search />}
      {step === PayStep.AMOUNT && <Amount />}
    </>
  )
}

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
  const term = form.get('term') as string
  const walletUrl = form.get('walletUrl') as string
  const identifier = form.get('identifier') as string
  const identifierType = form.get('identifierType') as string

  const fieldErrors = {
    form: '',
    address: '',
    amount: ''
  }

  if (formName === 'quote') {
    const amount = form.get('amount') as string
    const note = form.get('note') as string
    const toLinkedAccountId = form.get('toLinkedAccountId') as string
    const amountToSubmit = String(Math.floor(parseFloat(amount) * 100))

    const expiresAt = {
      seconds: `${Math.floor(DateTime.now().plus({ hour: 1 }).toSeconds())}`,
      nanos: 0
    }

    const fieldErrors = {
      form: '',
      amount: '',
      linkedAccount: '',
      note: ''
    }

    if (amountToSubmit == 'NaN') {
      fieldErrors.amount = 'Amount is required.'
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    }

    let walletInfo = await getWalletInfo(request)
    let receivePaymentPointer = flow.data.address.walletUrl

    const response = await openPaymentsClient
      .createQuote(
        {
          sendPaymentPointer: walletInfo.url,
          receivePaymentPointer,
          description: note,
          amount: {
            amount: amountToSubmit,
            asset: 'USD',
            assetScale: 2
          },
          expiresAt,
          sendLinkedAccount: toLinkedAccountId
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
      if (response.code == 3) {
        for (let violation of (response as GrpcError).details[0]
          .fieldViolations) {
          const field = mapper(violation.field as fieldErrorsMap)
          // if (field != null) fieldErrors[field] = violation.description
        }
        return json({ errors: { ...fieldErrors } }, { status: 400 })
      } else throw json({}, httpMapping(response.code))
    }

    let sendAmount = response.response.sendAmount?.amount,
      receiveAmount = response.response.receiveAmount?.amount,
      fee = 0

    // TODO: should fetch this information directly from the quote.
    const data = {
      errors: { ...fieldErrors },
      quoteID: response.response.id,
      note,
      amount: amount,
      fee: fee,
      toLinkedAccountId,
      displayFee: formatMoney(fee),
      sendAmount,
      displaySendAmount: formatMoney(parseFloat(sendAmount as string) / 100),
      receiveAmount,
      displayReceiveAmount: formatMoney(
        parseFloat(receiveAmount as string) / 100
      ),
      receivePaymentPointer,
      sendPaymentPointer: walletInfo.url,
      idempotencyKey: v4()
    }

    return json(data)
  }

  switch (formName) {
    case 'search':
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

      if (isGrpcError(response)) {
        if (response.code == 3) {
          for (let violation of (response as GrpcError).details[0]
            .fieldViolations) {
            const field = mapper(violation.field as fieldErrorsMap)
            if (field != null) fieldErrors[field] = violation.description
          }
          return json(
            { results: [], errors: { ...fieldErrors } },
            { status: 400 }
          )
        } else if (response.code == 5) {
          fieldErrors.address = 'Wallet address not found.'
          return json(
            { results: [], errors: { ...fieldErrors } },
            { status: 400 }
          )
        } else throw json({}, httpMapping(response.code))
      }
      return json({
        results: response.response.results.filter((v) => v.canSend)
      })

    case 'submit':
      console.log(
        'submit',
        formName,
        term,
        walletUrl,
        identifier,
        identifierType
      )
      // await updateFlow(request, flowType.Pay, {
      //   term,
      //   address: { walletUrl, identifier, identifierType }
      // })
      return redirect(route('/pay/amount'))
    default:
      throw json(
        { title: "Submitted a form that doesn't exist" },
        {
          status: 400
        }
      )
  }
}

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
