import { AccessAction } from '@interledger/open-payments'
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
import { buildConsentCards } from '~/lib/consent-cards'
import { getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import { consent, getInteraction } from '~/lib/rafikiauth.server'
import type { Route } from './+types/consent'

const FORBIDDEN_ACTIONS: Partial<AccessAction>[] = ['list-all']

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
    state: interaction.state,
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
  const { access, subject, state, clientName, clientUri } =
    useLoaderData<typeof loader>()
  const [params] = useSearchParams()
  const isPendingGrant = state === 'PENDING'

  if (!isPendingGrant) {
    return (
      <Card>
        <CardContent>
          <span className='text-lg'>{clientName}</span> previously sent this
          request.
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
    )
  }

  const opAccess = access.find((a) => a.type === 'outgoing-payment')
  const cards = buildConsentCards(opAccess)
  // limits take priority: a grant with a debit/receive amount
  // is shown as a payment, even when a subject is present
  const isIdentityRequest =
    cards.length === 0 && Boolean(subject?.sub_ids?.length)

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
          <span className='text-lg'>{clientName}</span>{' '}
          {isIdentityRequest
            ? 'wants to confirm your identity.'
            : 'is requesting access to the following:'}
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

      {!isIdentityRequest &&
        cards.map((card, index) => (
          <Card key={index}>
            <CardContent>
              <div className='flex w-full justify-between'>
                <span className='text-medium'>{card.label}</span>
                {card.value && <span className='text-error'>{card.value}</span>}
              </div>
              {card.description && (
                <p className='mt-0.5 text-sm text-weak'>{card.description}</p>
              )}
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
  const opAccess = interaction.access.find((a) => a.type === 'outgoing-payment')
  const hasActions = Boolean(opAccess?.actions.length)
  if (!hasSubject && !hasActions) {
    throw data({}, 403)
  }

  const isAccessForbidden = opAccess?.actions.some((action) =>
    FORBIDDEN_ACTIONS.includes(action)
  )
  if (isAccessForbidden) {
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
    const { search } = new URL(request.url)
    throw redirect(`${href('/no-access')}${search}`)
  }

  return interaction
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
