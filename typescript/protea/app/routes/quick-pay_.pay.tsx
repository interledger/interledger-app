import type { Route } from './+types/quick-pay_.pay'
import { data, redirect } from 'react-router'
import { Form, useActionData, useLoaderData, useNavigation } from 'react-router'
import type { MetaFunction } from 'react-router'
import { useEffect, useState } from 'react'
import { z } from 'zod'
import { ActionMessage, type ApplicationProps, Button, GridColumn, Layouts, TextField, WalletGrid } from '~/components'
import { AmountDisplay, QuoteDialog, PayWithInterledgerMark } from '~/components/QuickPay/'
import { useDialPadStore } from '~/lib/useDialPadStore'
import { mergeMeta } from '~/lib/meta'
import { fetchQuote, getValidWalletAddress, initializePayment } from '~/lib/open-payments.server'
import { paymentSchema, formatAmount, createError, routeAllowed } from '~/lib/utils.server'
import { commitSession, getSession } from '~/session.server'
import { type ActionData, QuickPaySession } from '~/lib/types'
import { formatError, NOTE_MAX_CHARACTERS, charactersRemaining } from '~/lib/helpers'
import logger from '~/lib/logger.server'
import { BackButton } from '~/components/QuickPay'

export async function loader({ request }: Route.LoaderArgs) {
  routeAllowed('OP_INTPAY_ENABLED')
  const searchParams = new URL(request.url).searchParams
  const isQuote = searchParams.get('quote') || false
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData = session.get('quickPay')
  const walletAddressInfo = sessionData?.senderAddress

  let receiverName = ''
  let receiveAmount = null
  let debitAmount = null

  if (walletAddressInfo === undefined) {
    throw data(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Payment session expired."
      },
      { status: 400 }
    )
  }

  if (isQuote) {
    const quote = sessionData?.quote
    const receiver = sessionData.receiverAddress

    if (quote === undefined) {
      throw data(
        {
          code: "QUICKPAY_SESSION_ERROR",
          title: "Payment session expired."
        },
        { status: 400 }
      )
    }

    receiverName =
      receiver.publicName === undefined ? 'Recepient' : receiver.publicName
    receiveAmount = formatAmount({
      value: quote.receiveAmount.value,
      assetCode: quote.receiveAmount.assetCode,
      assetScale: quote.receiveAmount.assetScale
    })

    debitAmount = formatAmount({
      value: quote.debitAmount.value,
      assetCode: quote.debitAmount.assetCode,
      assetScale: quote.debitAmount.assetScale
    })
  }

  return data({
    senderAddress: walletAddressInfo.id,
    receiveAmount: receiveAmount ? receiveAmount.amountWithCurrency : null,
    debitAmount: debitAmount ? debitAmount.amountWithCurrency : null,
    receiverName: receiverName,
    isQuote: isQuote
  } as const)
}

export const handle: ApplicationProps = {
  layout: (_match, context) => context?.isUser ? Layouts.Wallet : Layouts.Marketing,
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
  const { senderAddress, receiveAmount, debitAmount, receiverName, isQuote } = useLoaderData<typeof loader>()
  const actionData = useActionData<ActionData>()
  const navigation = useNavigation()
  const isSubmitting = navigation.state === "submitting"
  const { amountValue } = useDialPadStore()
  const [modalOpen, setModalOpen] = useState(false)
  const errors = actionData?.errors
  const [errorsList, setErrorsList] = useState(errors)
  const [note, setNote] = useState('')

  useEffect(() => { setErrorsList(errors) }, [errors, setErrorsList])

  useEffect(() => {
    setModalOpen(Boolean(isQuote))
  }, [isQuote])

  console.log({ errorsList })

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full mt-20 mx-auto'
      >
        <BackButton title="Back" to="/quick-pay/amount" />
        <AmountDisplay />
        <div className="mx-auto w-full max-w-sm">
          <Form method="POST">
            <div className="flex flex-col gap-4">
              <TextField
                type="text"
                label="Pay from"
                name="senderAddress"
                value={senderAddress}
                readOnly
              />
              <TextField
                label="Pay into"
                name="receiverAddress"
                placeholder="Enter receiver wallet address"
                autoFocus
                errorMessage={formatError(errorsList?.receiverAddress)}
              />
              <TextField
                label="Payment note"
                name="note"
                placeholder="Note"
                value={note}
                maxLength={NOTE_MAX_CHARACTERS}
                onChange={(e) => {
                  setNote(e.target.value)
                  setErrorsList({ ...errorsList, note: undefined })
                }}
                errorMessage={formatError(errorsList?.note)}
                successMessage={charactersRemaining(note)}
              />
              <input
                type="hidden"
                name="amount"
                value={Number(amountValue)}
              />
              <div className="flex justify-center">
                <Button
                  aria-label="pay"
                  type="submit"
                  name="intent"
                  value="pay"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <span className="animate-pulse">Processing...</span>
                  ) : (
                    <>
                      <span className="text-md">Pay with</span>
                      <PayWithInterledgerMark className="h-8 w-40 mx-2" />
                    </>
                  )}
                </Button>
              </div>
              <ActionMessage message={formatError(errors?.actionError)} />
            </div>
          </Form>
        </div>
        <QuoteDialog
          showDialog={modalOpen}
          setShowDialog={setModalOpen}
          receiverName={receiverName}
          receiveAmount={receiveAmount || ''}
          debitAmount={debitAmount || ''}
        />
      </GridColumn>
    </WalletGrid>)
}

export async function action({ request }: Route.ActionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const sessionData: QuickPaySession = session.get('quickPay') || {}
  const walletAddressInfo = sessionData.senderAddress

  const formData = Object.fromEntries(await request.formData())
  const intent = formData.intent

  if (intent !== 'pay' && intent !== 'confirm') {
    sessionData.quote = undefined
    session.set('quickPay', sessionData)
    return redirect(`/quick-pay/pay`, {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  }

  if (intent === 'pay') {
    const result = paymentSchema.safeParse(formData)

    if (!result.success) {
      const errors = z.treeifyError(result.error).properties
      return data({
        errors
      })
    }

    let receiverAddress
    try {
      receiverAddress = await getValidWalletAddress(result.data.receiverAddress)
      sessionData.receiverAddress = receiverAddress
      session.set('quickPay', sessionData)
    } catch (err) {
      return data({ errors: createError("receiverAddress", "Your wallet address is not valid.") })
    }

    try {
      const quote = await fetchQuote(result.data, receiverAddress)
      sessionData.quote = quote
      session.set('quickPay', sessionData)

    } catch (err) {
      logger.error({ err }, 'Error getting quote.')
      return data({ errors: createError("actionError", "An error occurred, please try again.") })
    }
    return redirect(`/quick-pay/pay?quote=true`, {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  }

  if (sessionData?.quote === undefined) {
    throw data(
      {
        code: "QUICKPAY_SESSION_ERROR",
        title: "Payment session expired."
      },
      { status: 400 }
    )
  }

  const quote = sessionData.quote
  if (intent === 'confirm') {
    if (quote === undefined || walletAddressInfo === undefined) {
      throw data(
        {
          code: "QUICKPAY_SESSION_ERROR",
          title: "Payment session expired."
        },
        { status: 400 }
      )
    }
  }

  const { paymentId, outgoingPaymentGrant } = await initializePayment({
    walletAddress: String(walletAddressInfo?.id),
    quote
  })
  sessionData.grants = { ...(sessionData?.grants || {}), [paymentId]: outgoingPaymentGrant }
  session.set('quickPay', sessionData)
  return redirect(outgoingPaymentGrant.interact.redirect, {
    headers: { 'Set-Cookie': await commitSession(session) }
  })
}
