import {
  Form,
  data,
  href,
  redirect,
  useLoaderData,
  useSearchParams
} from 'react-router'
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
import { envValue } from '~/env.server'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import type { Amount } from '~/lib/rafikiauth.server'
import { consent, getInteraction } from '~/lib/rafikiauth.server'
import type { Route } from './+types/consent'

// Outgoing-payment actions the consent screen implements:
// Grants requesting any other action aren't supported yet and 404.
const SUPPORTED_ACTIONS = ['create', 'read-all']

export async function loader({ request }: Route.LoaderArgs) {
  await getUserSession(request)

  const url = new URL(request.url)
  const interactId = url.searchParams.get('interactId') || ''
  const nonce = url.searchParams.get('nonce') || ''
  const clientName = url.searchParams.get('clientName') || ''
  const clientUri = url.searchParams.get('clientUri') || ''

  const interaction = await requireOwnedInteraction(request, interactId, nonce)

  return data({
    access: interaction.access,
    subject: interaction.subject,
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
  const { access, subject, clientName, clientUri } =
    useLoaderData<typeof loader>()
  const [params] = useSearchParams()
  const amounts = access
    .map((entry) => entry.limits?.debitAmount ?? entry.limits?.receiveAmount)
    .filter((amount): amount is Amount => Boolean(amount))
    .map(formatAmount)
  // limits take priority: a grant with a debit/receive amount
  // is shown as a payment, even when a subject is present
  const isIdentityRequest =
    amounts.length === 0 && Boolean(subject?.sub_ids?.length)

  const message = isIdentityRequest
    ? 'wants to confirm your identity.'
    : amounts.length > 0
      ? 'is requesting access to the following:'
      : 'wants to view your payments.'

  return (
    <>
      <Form
        id='consent'
        action={`/consent?${params}`}
        method='post'
        className='hidden'
      />

      <Card>
        <CardContent>
          <span className='text-lg'>{clientName}</span> {message}
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

      {amounts.map((amount, index) => (
        <Card key={index}>
          <CardContent>
            <div className='flex w-full justify-between'>
              <span className='text-medium'>Total amount to debit</span>
              <span className='text-error'>{amount}</span>
            </div>
          </CardContent>
        </Card>
      ))}

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

async function requireOwnedInteraction(
  request: Request,
  interactId: string,
  nonce: string
) {
  const interaction = await getInteraction(interactId, nonce)
  const hasSubject = Boolean(interaction.subject?.sub_ids?.length)
  const isAccessSupported = interaction.access.some(
    (access) =>
      access.type === 'outgoing-payment' &&
      access.actions.every((action) => SUPPORTED_ACTIONS.includes(action))
  )

  if (!hasSubject && !isAccessSupported) {
    throw data({}, 404)
  }

  const userWalletAddress = (await getWalletInfo(request)).url

  const accessIdentifiers = interaction.access.map(
    (access) => access.identifier
  )
  const subjectIdentifiers =
    interaction.subject?.sub_ids.map((subId) => subId.id) ?? []

  const referencedWalletAddresses = [
    ...accessIdentifiers,
    ...subjectIdentifiers
  ].filter((address): address is string => Boolean(address))

  const userOwnsEveryReferencedWallet = referencedWalletAddresses.every(
    (address) => address.includes(userWalletAddress)
  )

  if (!userOwnsEveryReferencedWallet) {
    try {
      await consent(interactId, nonce, 'reject')
    } catch {
      // the grant is denied regardless
    }
    throw redirect(href('/no-access'))
  }

  return interaction
}

function formatAmount(amount: Amount): string {
  let currency = '$'
  if (amount.assetCode != 'USD') {
    currency = amount.assetCode
  }

  const amt = parseInt(amount.value) * Math.pow(10, -amount.assetScale)
  return `${currency} ${amt.toFixed(2)}`
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const action = String(form.get('action') || '')
  const url = new URL(request.url)
  const interactId = url.searchParams.get('interactId') || ''
  const nonce = url.searchParams.get('nonce') || ''

  await requireOwnedInteraction(request, interactId, nonce)

  const userDecision: 'accept' | 'reject' =
    action == 'approve' ? 'accept' : 'reject'
  await consent(interactId, nonce, userDecision)

  // TODO: Move to environment variables
  const publicOpenPaymentsAuthHost = envValue('PUBLIC_OP_AUTH_HOST')

  return redirect(
    `https://${publicOpenPaymentsAuthHost}/interact/${interactId}/${nonce}/finish`
  )
}
