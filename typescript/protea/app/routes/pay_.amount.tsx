import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import type { ChangeEventHandler } from 'react'
import { useCallback, useState } from 'react'
import { route } from 'routes-gen'
import { v4 } from 'uuid'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Button,
  Card,
  CardButton,
  CardContent,
  CardIcon,
  FynbosIcon,
  Icon,
  Layouts,
  Select,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { hasUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { getLinkedAccounts, getWalletInfo } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  if (!hasUserSession(request)) {
    // Handles un-authed clicks from /me/$
    return redirect('/login?return_to=/pay/amount')
  }
  const flow = await requireFlow(request, flowType.Pay)
  const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)

  const linkedAccounts = [...cardAccounts, ...bankAccounts].filter(
    (acc) => acc.canSend
  )

  if (linkedAccounts.length === 0) {
    await flashSnackbar(request, {
      message: 'You need a connected account to make a payment.',
      icon: 'close'
    })
    return redirect(route('/accounts'))
  }

  return json({
    flow,
    linkedAccounts
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/pay'),
      title: 'Make a payment'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay | Amount'
  }
}

export default function Page() {
  const { flow, linkedAccounts } = useLoaderData<typeof loader>()
  const fetcher = useFetcher()

  const [linkedAccount, setLinkedAccount] = useState<{
    id: string
    name: string
  }>(linkedAccounts[0])

  const _onChangeLinkedAccount = useCallback((event: SelectOptions) => {
    setLinkedAccount(event)
  }, [])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let amount = event.target.value
      fetcher.submit(
        { amount: amount, toLinkedAccountId: linkedAccount.id },
        { method: 'post' }
      )
    },
    [fetcher, linkedAccount.id]
  )

  return (
    <>
      <fetcher.Form
        id='amount-form'
        action='/pay/amount'
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {flow?.data.displayReceiveAmount || '$ 0.00'}
            </h2>
            <CardIcon>
              {flow.data.address.type === 'wallet' && <FynbosIcon />}
              {flow.data.address.type === 'twitter' && <TwitterIcon />}
            </CardIcon>
          </div>
          <Label className='-mb-5 mt-4'>Payment to</Label>
        </CardContent>
        <CardButton>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{flow.data.address.handle}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <TextField
            id='amount'
            form='amount-form'
            label='Amount'
            name='amount'
            defaultValue={flow?.data.amount}
            onChange={_onChangeInput}
            prefix='$'
            type='number'
            min='0'
            step='0.01'
            aria-invalid={Boolean(fetcher.data?.errors.amount) || undefined}
            aria-describedby={
              fetcher.data?.errors.amount ? 'amount-error' : undefined
            }
            errorMessage={fetcher.data?.errors.amount || undefined}
            required
          />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <span>Select an account to pay from:</span>
          <Select
            id='linkedAccount'
            label='Connected accounts'
            className='mt-4'
            value={linkedAccount}
            options={linkedAccounts}
            onChange={_onChangeLinkedAccount}
            aria-invalid={
              Boolean(fetcher.data?.errors.linkedAccount) || undefined
            }
            aria-describedby={
              fetcher.data?.errors.linkedAccount
                ? 'linkedAccount-error'
                : undefined
            }
            errorMessage={fetcher.data?.errors.linkedAccount}
          />
          <input
            form='amount-form'
            value={linkedAccount.id}
            name='toLinkedAccountId'
            type='hidden'
          />
          <TextField
            id='note'
            label='Reference'
            name='note'
            form='amount-form'
            type='text'
            defaultValue={flow.data.note || ''}
            className='mt-4'
            aria-invalid={Boolean(fetcher.data?.errors.note) || undefined}
            aria-describedby={
              fetcher.data?.errors.note ? 'reference-error' : undefined
            }
            errorMessage={fetcher.data?.errors.note}
          />
        </CardContent>
      </Card>
      <Button form='amount-form' type='submit' name='route-to' value='next'>
        Continue
      </Button>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'amount'

function mapper(field: fieldErrorsMap): 'amount' | null {
  switch (field) {
    case 'amount':
      return 'amount'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const amount = form.get('amount') as string
  const note = form.get('note') as string
  const toLinkedAccountId = form.get('toLinkedAccountId') as string
  const amountToSubmit = String(Math.floor(parseFloat(amount) * 100))
  const routeTo = form.get('route-to')

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

  // TODO: Submit note with quote
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
        if (field != null) fieldErrors[field] = violation.description
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

  await updateFlow(request, flowType.Pay, data)

  // TODO: should always return data, because using fetcher means redirecting from here doesn't add the route to the stack which breaks the back button.
  if (routeTo == 'next') {
    return redirect(route('/pay/confirm'))
  } else {
    return json(data)
  }
}

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
