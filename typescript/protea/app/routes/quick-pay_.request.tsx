import { useEffect, useState } from 'react'
import type { MetaFunction } from 'react-router'
import {
  Form,
  data,
  redirect,
  useActionData,
  useLoaderData
} from 'react-router'
import { z } from 'zod'
import type { ApplicationProps } from '~/components'
import {
  Button,
  GridColumn,
  Layouts,
  TextField,
  WalletGrid
} from '~/components'
import { Icon } from '~/components/Icon'
import { AmountDisplay } from '~/components/QuickPay/Dialpad'
import {
  NOTE_MAX_CHARACTERS,
  charactersRemaining,
  formatError
} from '~/lib/helpers'
import { mergeMeta } from '~/lib/meta'
import { createRequestPayment } from '~/lib/open-payments.server'
import { QuickPaySession } from '~/lib/types'
import { useDialPadStore } from '~/lib/useDialPadStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import {
  formatAmount,
  formatDate,
  requestPaymentSchema,
  routeAllowed
} from '~/lib/utils.server'
import { commitSession, getSession } from '~/session.server'
import type { Route } from './+types/quick-pay_.request'

export async function loader({ request }: Route.LoaderArgs) {
  routeAllowed('OP_INTPAY_ENABLED')
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData = session.get('quickPay')
  const walletAddressInfo = sessionData?.senderAddress
  const assetCode = walletAddressInfo?.senderAddress?.assetCode

  const incomingPayment = sessionData?.request
  const showGeneratedRequest = !!incomingPayment

  if (walletAddressInfo === undefined) {
    throw data(
      {
        code: 'QUICKPAY_SESSION_ERROR',
        title: 'Payment session expired.'
      },
      { status: 400 }
    )
  }

  const incomingPaymentData = showGeneratedRequest
    ? {
        amount: formatAmount({
          value: incomingPayment.incomingAmount.value,
          assetCode: incomingPayment.incomingAmount.assetCode,
          assetScale: incomingPayment.incomingAmount.assetScale
        }),
        date: formatDate({
          date: incomingPayment.createdAt
        }),
        url: `${process.env.OP_INTPAY_HOST}quick-pay/payment?url=${incomingPayment.id}&receiver=${incomingPayment.walletAddress}`,
        note: incomingPayment?.metadata?.description
      }
    : undefined

  return data({
    walletAddress: walletAddressInfo.id,
    assetCode,
    incomingPaymentData,
    showGeneratedRequest
  } as const)
}

export const handle: ApplicationProps = {
  layout: (_match, context) =>
    context?.isUser ? Layouts.Wallet : Layouts.Marketing,
  scaffold: {
    header: { title: 'Interledger Pay' }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Interledger Pay'
  }
])

export default function Page() {
  const {
    walletAddress,
    assetCode,
    incomingPaymentData,
    showGeneratedRequest
  } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const { amountValue } = useDialPadStore()
  const errors = actionData?.errors
  const [copied, setCopied] = useState(false)
  const pushSnackbar = useScaffoldStore((state) => state.pushSnackbar)
  const [errorsList, setErrorsList] = useState(errors)
  const [note, setNote] = useState('')

  useEffect(() => {
    setErrorsList(errors)
  }, [errors, setErrorsList])

  function copyToClipboard(e: React.MouseEvent<HTMLElement>, value: string) {
    e.preventDefault()
    setCopied(true)
    navigator.clipboard.writeText(value)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  function shareUrl(e: React.MouseEvent<HTMLElement>, url: string) {
    e.preventDefault()

    const showFallback = (message: string) => {
      pushSnackbar({
        id: `share-${Date.now()}`,
        message,
        icon: 'info'
      })
    }

    if (navigator.share) {
      navigator
        .share({
          title: 'Payment link',
          text: 'Interledger Pay payment link:',
          url
        })
        .catch(() => {
          navigator.clipboard.writeText(url)
          showFallback('Sharing failed. Link copied to clipboard.')
        })
    } else {
      navigator.clipboard.writeText(url)
      showFallback('Sharing not supported. Link copied to clipboard.')
    }
  }

  return (
    <WalletGrid>
      <GridColumn className='col-span-full mx-auto mt-20'>
        <AmountDisplay
          displayAmount={
            showGeneratedRequest && incomingPaymentData
              ? String(incomingPaymentData.amount.amount)
              : amountValue
          }
          assetCode={assetCode}
        />
        <div className='mx-auto w-full max-w-sm'>
          {showGeneratedRequest && incomingPaymentData ? (
            <Form method='POST'>
              <TextField
                label='Amount requested'
                defaultValue={incomingPaymentData.amount.amountWithCurrency}
                disabled
              ></TextField>
              <TextField
                label='Date requested'
                defaultValue={incomingPaymentData.date}
                disabled
              ></TextField>
              <TextField
                label='Request note'
                defaultValue={
                  incomingPaymentData.note === undefined
                    ? '-'
                    : incomingPaymentData.note
                }
                disabled
              ></TextField>
              <div className="before:border-input after:border-inout mx-4 mb-6 flex items-center justify-center gap-4 text-sm font-light before:flex-1 before:border before:border-solid before:content-[''] after:flex-1 after:border after:border-solid after:content-['']">
                Share link
              </div>
              <TextField
                label='Payment link'
                readOnly
                defaultValue={incomingPaymentData.url}
                className='-z-1 flex-1'
                appendIcon={
                  <>
                    <Button
                      aria-label='copy payment link'
                      className='h-7 w-7 bg-transparent'
                      onClick={(e) =>
                        copyToClipboard(e, incomingPaymentData.url)
                      }
                      type='button'
                    >
                      {copied ? <Icon>check</Icon> : <Icon>content_copy</Icon>}
                    </Button>
                    <Button
                      aria-label='share payment link'
                      className='h-7 w-7 bg-transparent'
                      type='button'
                      onClick={(e) => {
                        shareUrl(e, incomingPaymentData.url)
                      }}
                    >
                      <Icon>share</Icon>
                    </Button>
                  </>
                }
              ></TextField>
              <div className='flex justify-center'>
                <Button type='submit' className='mt-8'>
                  Close
                </Button>
              </div>
            </Form>
          ) : (
            <Form method='POST'>
              <input type='hidden' name='amount' value={Number(amountValue)} />
              <div className='flex flex-col gap-4'>
                <TextField
                  type='text'
                  label='Wallet Address'
                  placeholder='Wallet address'
                  name='receiverAddress'
                  value={walletAddress}
                  readOnly
                  defaultValue={walletAddress || ''}
                  errorMessage={formatError(errorsList?.receiverAddress)}
                />
                <TextField
                  label='Payment note'
                  name='note'
                  placeholder='Note'
                  value={note}
                  maxLength={NOTE_MAX_CHARACTERS}
                  onChange={(e) => {
                    setNote(e.target.value)
                    setErrorsList({ ...errorsList, note: undefined })
                  }}
                  errorMessage={formatError(errorsList?.note)}
                  successMessage={charactersRemaining(note)}
                />
                <div className='flex justify-center'>
                  <Button
                    aria-label='Create Request'
                    type='submit'
                    name='intent'
                    value='request'
                  >
                    Create Request
                  </Button>
                </div>
              </div>
            </Form>
          )}
        </div>
      </GridColumn>
    </WalletGrid>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData: QuickPaySession = session.get('quickPay') || {}
  const formData = Object.fromEntries(await request.formData())
  const intent = formData.intent
  const path = '/quick-pay/request'

  if (intent !== 'request') {
    sessionData.request = undefined
    session.set('quickPay', sessionData)
    return redirect(path, {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  }

  const result = requestPaymentSchema.safeParse(formData)

  if (!result.success || !result.data) {
    const errors = z.treeifyError(result.error).properties
    return data({
      errors
    })
  }

  const incomingPayment = await createRequestPayment(result.data)
  sessionData.request = incomingPayment
  session.set('quickPay', sessionData)

  return redirect(path, {
    headers: { 'Set-Cookie': await commitSession(session) }
  })
}
