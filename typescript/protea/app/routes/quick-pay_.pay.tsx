import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { z } from 'zod'
import type { ApplicationProps } from '~/components'
import { ActionMessage, Button, GridColumn, Layouts, TextField, WalletGrid } from '~/components'
import { AmountDisplay, QuoteDialog, PayWithInterledgerMark } from '~/components/QuickPay/'
import { useDialPadContext } from '~/lib/context/dialpad'
import { mergeMeta } from '~/lib/meta'
import { fetchQuote, initializePayment } from '~/lib/open-payments.server'
import { getValidWalletAddress, paymentSchema, formatAmount, formatError, Errors, createError } from '~/lib/utils'
import { commitSession, getSession } from '~/session.server'

type ActionData = {
  errors?: {
    walletAddress?: Errors
    receiverAddress?: Errors
    note?: Errors
    actionError?: Errors
  }
}

export async function loader({ request }: LoaderFunctionArgs) {
  const searchParams = new URL(request.url).searchParams
  const isQuote = searchParams.get('quote') || false

  const session = await getSession(request.headers.get('Cookie'))
  const sessionData = session.get('quickPay')
  const walletAddressInfo = sessionData?.validWalletAddress

  let receiverName = ''
  let receiveAmount = null
  let debitAmount = null

  if (walletAddressInfo === undefined) {
    throw json(
      {
        code: "QUICKPAY_SESSION_ERROR",
        message: "Payment session expired."
      },
      { status: 400 }
    )
  }

  if (isQuote) {
    const quote = session.data.quote
    const receiver = sessionData.receiverAddress

    if (quote === undefined) {
      throw json(
        {
          code: "QUICKPAY_SESSION_ERROR",
          message: "Payment session expired."
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

  return json({
    senderAddress: walletAddressInfo.id,
    receiveAmount: receiveAmount ? receiveAmount.amountWithCurrency : null,
    debitAmount: debitAmount ? debitAmount.amountWithCurrency : null,
    receiverName: receiverName,
    isQuote: isQuote
  } as const)
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
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
  const { amountValue } = useDialPadContext()
  const [modalOpen, setModalOpen] = useState(false)
  const errors = actionData?.errors

  useEffect(() => {
    setModalOpen(Boolean(isQuote))
  }, [isQuote])

  return (
    <WalletGrid>
      <GridColumn
        className='col-span-full mt-20 mx-auto'
      >
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
                errorMessage={formatError(errors?.receiverAddress)}
              />
              <TextField
                label="Payment note"
                name="note"
                placeholder="Note"
                errorMessage={formatError(errors?.note)}
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
                >
                  <span className="text-md">Pay with</span>
                  <PayWithInterledgerMark className="h-8 w-40 mx-2" />
                </Button>
              </div>
              <ActionMessage message= {formatError(errors?.actionError)} />
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

export async function action({ request }: ActionFunctionArgs) {
  const session = await getSession(request.headers.get('Cookie'))
  const formData = Object.fromEntries(await request.formData())

  const intent = formData.intent

  if (intent !== 'pay' && intent !== 'confirm') {
    return json(formData)
  }

  const result = paymentSchema.safeParse(formData)

  if (!result.success) {
    const errors = z.treeifyError(result.error).properties
    return json({
      errors
    })
  }

  let receiverAddress
  try {
    receiverAddress = await getValidWalletAddress(result.data.receiverAddress)
    session.set('quickPay', { receiverAddress: receiverAddress })
  } catch (err) {
    return json({ errors: createError("receiverAddress", "Your wallet address is not valid.") })
  }

  const sessionData = session.get('quickPay')
  const walletAddressInfo = sessionData.validWalletAddress

  if (intent === 'pay') {
    try {
      const quote = await fetchQuote(result.data, receiverAddress)
      session.set('quickPay', { quote: quote })

      return redirect(`/quick-pay/pay?quote=true`, {
        headers: { 'Set-Cookie': await commitSession(session) }
      })
    } catch (err) { console.log({ err }) }
    return json({ errors: createError("actionError", "An error occured, please try again.") })
  }

  const quote = sessionData?.quote
  if (intent === 'confirm') {


    if (quote === undefined || walletAddressInfo === undefined) {
      throw json(
        {
          code: "QUICKPAY_SESSION_ERROR",
          message: "Payment session expired."
        },
        { status: 400 }
      )
    }
  }

  const grant = await initializePayment({
    walletAddress: walletAddressInfo.walletAddress.id,
    quote
  })

  session.set('payment-grant', grant)
  return redirect(grant.interact.redirect, {
    headers: { 'Set-Cookie': await commitSession(session) }
  })
}

