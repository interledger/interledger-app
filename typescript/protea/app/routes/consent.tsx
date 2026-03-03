import type { Route } from './+types/consent'
import { data, redirect } from 'react-router';
import { Form, useLoaderData, useSearchParams } from 'react-router';
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardButton,
  CardContent,
  Layouts,
  TextButton
} from '~/components'
import { getWalletInfo } from '~/data/wallet.server'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import type { Amount } from '~/lib/rafikiauth'
import { consent, getInteraction } from '~/lib/rafikiauth'

export async function loader({ request }: Route.LoaderArgs) {
  await getUserSession(request)

  const url = new URL(request.url)
  let interactId = url.searchParams.get('interactId') || ''
  let nonce = url.searchParams.get('nonce') || ''
  let clientName = url.searchParams.get('clientName') || ''
  let clientUri = url.searchParams.get('clientUri') || ''
  let grants = await getInteraction(interactId, nonce)

  // there should be a grant. Throw 404 for now.
  if (grants.length < 1) {
    throw data({}, 404)
  }

  return data({
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

export const meta = mergeMeta(() => [
  {
    title: 'Consent'
  }
])

export default function Page() {
  const { type } = useLoaderData()
  const [params] = useSearchParams()

  return (
    <>
      <Form
        id='consent'
        action={`/consent?${params}`}
        method='post'
        className='hidden'
      />

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
  const { clientName, clientUri } = useLoaderData()
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
  const { clientName, clientUri } = useLoaderData()
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
  const { clientName, limits, clientUri } = useLoaderData()
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

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const action = String(form.get('action') || '')
  const url = new URL(request.url)
  let interactId = url.searchParams.get('interactId') || ''
  let nonce = url.searchParams.get('nonce') || ''

  let grants = await getInteraction(interactId, nonce)

  // there should be a grant. Throw 404 for now.
  if (grants.length < 1) {
    throw data({}, 404)
  }

  let walletInfo = await getWalletInfo(request)
  let ownsResource = false
  grants.forEach((a) => {
    if (a.identifier?.includes(walletInfo.url)) {
      ownsResource = true
    }
  })
  if (!ownsResource) {
    throw data({}, 403)
  }

  let userDecision: 'accept' | 'reject' =
    action == 'approve' ? 'accept' : 'reject'
  await consent(interactId, nonce, userDecision)

  // TODO: Move to environment variables
  let publicOpenPaymentsAuthHost = 'auth.ilp.link'
  if (process.env.FYNBOS_ENV == 'dev') {
    publicOpenPaymentsAuthHost = 'auth.sandbox.ilp.link'
  } else if (process.env.FYNBOS_ENV == 'local') {
    publicOpenPaymentsAuthHost = 'auth.local.ilp.link'
  }

  return redirect(
    `https://${publicOpenPaymentsAuthHost}/interact/${interactId}/${nonce}/finish`
  )
}
