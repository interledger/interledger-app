import { WalletAddress } from '@interledger/open-payments'
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
import { Button, GridColumn, Layouts, TextField, WalletGrid } from '~/components'
import { AmountDisplay, QuoteDialog, PayWithInterledgerMark } from '~/components/QuickPay/'
import { useDialPadContext } from '~/lib/context/dialpad'
import { mergeMeta } from '~/lib/meta'
import { fetchQuote, initializePayment } from '~/lib/open-payments.server'
import { formatAmount, formatError } from '~/lib/utils'
import { getValidWalletAddress, paymentSchema } from '~/lib/utils'
import { commitSession, destroySession, getSession } from '~/session.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const searchParams = new URL(request.url).searchParams
  const isQuote = searchParams.get('quote') || false

  const session = await getSession(request.headers.get('Cookie'))
  const walletAddressInfo = session.get('wallet-address')

  let receiverName = ''
  let receiveAmount = null
  let debitAmount = null

  if (walletAddressInfo === undefined) {
    throw new Error('Payment session expired.')
  }

  if (isQuote) {
    const quote = session.get('quote')
    const receiver = session.get('receiver-wallet-address')

    if (quote === undefined) {
      throw new Error('Payment session expired.')
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
    senderAddress: walletAddressInfo.walletAddress.id,
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
  const actionData = useActionData<typeof action>()
  const { amountValue } = useDialPadContext()
  const [modalOpen, setModalOpen] = useState(false)

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
                errorMessage={actionData?.errors?.receiverAddress}
              />
              <TextField
                label="Payment note"
                name="note"
                placeholder="Note"
                errorMessage={actionData?.errors?.note}
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

  let receiverAddress = {} as WalletAddress

  const formData = await request.formData()
  const intent = formData.get('intent')

  if (intent === 'pay') {
    const formData = Object.fromEntries(await request.formData())

    if (!formData.value || formData.intent !== 'submit') {
      return json(formData)
    }

    const quote = await fetchQuote(formData.value, receiverAddress)
    session.set('quote', quote)

    return redirect(`/quick-pay/pay?quote=true`, {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  } else if (intent === 'confirm') {
    const quote = session.get('quote')
    const walletAddressInfo = session.get('wallet-address')

    if (quote === undefined || walletAddressInfo === undefined) {
      throw new Error('Payment session expired.')
    }

    const grant = await initializePayment({
      walletAddress: walletAddressInfo.walletAddress.id,
      quote: quote
    })

    session.set('payment-grant', grant)
    return redirect(grant.interact.redirect, {
      headers: { 'Set-Cookie': await commitSession(session) }
    })
  } else {
    return redirect('/quick-pay/', {
      headers: { 'Set-Cookie': await destroySession(session) }
    })
  }
}
