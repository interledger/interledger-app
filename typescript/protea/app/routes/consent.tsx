import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  ApplicationProps,
  Button,
  Card,
  CardButton,
  CardContent,
  Layouts,
  TextButton
} from '~/components'
import { getWalletInfo } from '~/data/wallet.server'
import { mergeMeta } from '~/lib/meta'
import { Amount, Consent, GetInteraction } from '~/lib/rafikiauth'

export async function loader({ request }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV == 'prod') {
    return redirect(route('/'))
  }

  const url = new URL(request.url)
  let interactId = url.searchParams.get('interactId') || ''
  let nonce = url.searchParams.get('nonce') || ''
  let clientName = url.searchParams.get('clientName') || ''
  let clientUri = url.searchParams.get('clientUri') || ''
  let grants = await GetInteraction(interactId, nonce)

  // there should be a grant. Throw 404 for now.
  if (grants.length < 1) {
    throw json({}, 404)
  }

  let walletInfo = await getWalletInfo(request)
  let ownsResource = true
  grants.forEach((a) => {
    if (!a.identifier) {
      ownsResource = false
    } else if (!a.identifier.includes(walletInfo.url)) {
      ownsResource = false
    }
  })
  if (!ownsResource) {
    throw json({}, 403)
  }

  return json({
    ...grants[0],
    clientName,
    clientUri,
    interactId,
    nonce
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Consent'
  }
])

export default function Page() {
  const { type, interactId, nonce } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='consent'
        action={route('/consent')}
        method='post'
        className='hidden'
      />
      <input
        form='consent'
        value={interactId}
        name='interactId'
        type='hidden'
      />
      <input form='consent' value={nonce} name='nonce' type='hidden' />

      {type == 'outgoing-payment' && <OutgoingPaymentGrant />}
      {type == 'incoming-payment' && <IncomingPaymentGrant />}
      {type == 'quote' && <QuoteGrant />}

      <CardContent className='mt-2 flex w-full justify-end space-x-6'>
        <TextButton form='consent' type='submit' name='action' value='deny'>
          Cancel
        </TextButton>
        <Button form='consent' type='submit' name='action' value='approve'>
          Approve
        </Button>
      </CardContent>
    </>
  )
}

function QuoteGrant() {
  const { clientName, clientUri } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent>
          {clientName} is requesting access to get quotes on your behalf.
        </CardContent>
        <CardButton
          noHover
          onClick={() => {
            /* do nothing  */
          }}
        >
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <span>{clientUri}</span>
            </div>
          </div>
        </CardButton>
      </Card>
    </>
  )
}

function IncomingPaymentGrant() {
  const { clientName, clientUri } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent>
          {clientName} is requesting access to create incoming payments on your
          account.
        </CardContent>
        <CardButton
          noHover
          onClick={() => {
            /* do nothing  */
          }}
        >
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <span>{clientUri}</span>
            </div>
          </div>
        </CardButton>
      </Card>
    </>
  )
}

function OutgoingPaymentGrant() {
  const { clientName, limits, clientUri } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent>
          {clientName} is requesting access to make a payment on your behalf.
        </CardContent>
        <CardButton
          noHover
          onClick={() => {
            /* do nothing  */
          }}
        >
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <span>{clientUri}</span>
            </div>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-medium'>Total amount to debit</span>
            <span className='text-error'>
              {limits && limits.debitAmount && formatAmount(limits.debitAmount)}
            </span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}

function formatAmount(amount: Amount): string {
  let currency = '$'
  if (amount.assetCode != 'USD') {
    currency = amount.assetCode
  }

  let amt = parseInt(amount.value) * Math.pow(10, -amount.assetScale)
  return `${currency} ${amt.toFixed(2)}`
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const action = String(form.get('action') || '')
  const interactId = String(form.get('interactId') || '')
  const nonce = String(form.get('nonce') || '')
  const rafikiAuthEndpoint =
    process.env.RAFIKI_AUTH_ENDPOINT || 'http://rafiki-rafiki-auth.rafiki:3006'

  let grants = await GetInteraction(interactId, nonce)

  // there should be a grant. Throw 404 for now.
  if (grants.length < 1) {
    throw json({}, 404)
  }

  let walletInfo = await getWalletInfo(request)
  let ownsResource = true
  grants.forEach((a) => {
    if (!a.identifier) {
      ownsResource = false
    } else if (!a.identifier.includes(walletInfo.url)) {
      ownsResource = false
    }
  })
  if (!ownsResource) {
    throw json({}, 403)
  }

  let userDecision: 'accept' | 'reject' =
    action == 'approve' ? 'accept' : 'reject'
  await Consent(interactId, nonce, userDecision)

  let publicOpenPaymentsAuthHost = 'fynbos.me'
  if (process.env.FYNBOS_ENV == 'dev') {
    publicOpenPaymentsAuthHost = 'eu1.fynbos.me'
  } else if (process.env.FYNBOS_ENV == 'local') {
    publicOpenPaymentsAuthHost = 'local.fynbos.me'
  }

  return redirect(
    `https://${publicOpenPaymentsAuthHost}/interact/${interactId}/${nonce}/finish`
  )
}
