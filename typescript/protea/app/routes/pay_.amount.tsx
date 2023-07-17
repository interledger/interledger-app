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
  CardHeader,
  CardLink,
  Chip,
  ChipColor,
  Dialog,
  FynbosIcon,
  Icon,
  Layouts,
  LinkedInIcon,
  Select,
  TextButton,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
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
import {
  getLinkedAccounts,
  getPublicWalletInfo,
  getWalletInfo
} from '~/lib/wallet.server'

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

  const publicWalletInfo = await getPublicWalletInfo(
    request,
    flow.data.address.walletUrl
  )

  return json({
    flow,
    linkedAccounts,
    publicWalletInfo
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
  const { flow, linkedAccounts, publicWalletInfo } =
    useLoaderData<typeof loader>()
  const fetcher = useFetcher()
  const [showDialog, setShowDialog] = useState<boolean>(false)

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
            {flow.data.address.identifierType === 'wallet' && (
              <FynbosIcon height='h-12' />
            )}
            {flow.data.address.identifierType === 'twitter' && (
              <TwitterIcon height='h-12' />
            )}
          </div>
        </CardContent>
        <Label className='mt-2'>Payment to</Label>
        <CardButton onClick={() => setShowDialog(true)}>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{flow.data.address.identifier}</span>
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
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>User information</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            You are viewing public information about the person you intend to
            pay.
          </span>
        </CardContent>
        <Label className='mt-4'>Public name</Label>
        <div className='mt-1 flex rounded-xl bg-nav p-3 text-medium'>
          <span className=''>{publicWalletInfo.publicName}</span>
        </div>
        <Label className='mt-2'>Wallet address</Label>
        <CardLink className='flex w-full' to={publicWalletInfo.address}>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <FynbosIcon />
              <span>{publicWalletInfo.shortAddress}</span>
            </div>
            <Icon>navigate_next</Icon>
          </div>
        </CardLink>
        {publicWalletInfo.identities.map((identity) => (
          <div key={identity.id} className='contents'>
            <Label className='mt-2 capitalize'>{identity.platform}</Label>
            <CardLink className='flex w-full' to={publicWalletInfo.address}>
              <div className='flex w-full items-center justify-between text-medium'>
                <div className='flex space-x-2'>
                  {identity.platform == 'twitter' && <TwitterIcon />}
                  {identity.platform == 'linkedin' && <LinkedInIcon />}
                  <span>{identity.identifier}</span>
                </div>
                {identity.state == 'verified' && (
                  <Chip color={ChipColor.green}>Verified</Chip>
                )}
              </div>
            </CardLink>
          </div>
        ))}

        <CardContent className='flex w-full justify-end space-x-6'>
          <TextButton type='button' onClick={() => setShowDialog(false)}>
            Close
          </TextButton>
        </CardContent>
      </Dialog>
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

type quoteLimitError = 
  | 'Failed precondition: LimitTransaction'
  | 'Failed precondition: LimitDaily'
  | 'Failed precondition: LimitMonthly'
  | 'Failed precondition: Limit6Monthly'

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
          fieldErrors["amount"] = "Exceeds per transaction limit."
          break
        case 'Failed precondition: LimitDaily':
          fieldErrors["amount"] = "Exceeds daily limit."
          break
        case 'Failed precondition: LimitMonthly':
          fieldErrors["amount"] = "Exceeds monthly limit."
          break
        case 'Failed precondition: Limit6Monthly':
          fieldErrors["amount"] = "Exceeds rolling 6 month limit."
          break
        default:
          fieldErrors["amount"] = "Exceeds account limit."
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
