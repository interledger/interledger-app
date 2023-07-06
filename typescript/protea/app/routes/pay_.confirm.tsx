import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardButton,
  CardContent,
  CardIcon,
  Checkbox,
  FynbosIcon,
  Icon,
  Layouts,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { getClientIP } from '~/lib/ip.server'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { getLinkedAccounts } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)
  const { cardAccounts, bankAccounts } = await getLinkedAccounts(request)
  return json({
    flow,
    traits: session.identity.traits,
    // TODO use a lookup for this account rather?
    linkedAccount: [...cardAccounts, ...bankAccounts].find(
      (account) => account.id == flow.data.toLinkedAccountId
    )
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/pay/amount'),
      title: 'Confirm Payment'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay | Confirm'
  }
}

export default function Page() {
  const { flow, linkedAccount } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='pay-confirm'
        action={route('/pay/confirm')}
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
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>
              Free <sup>*</sup>
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>
              {flow?.data.displayReceiveAmount || '$ 0.00'}
            </span>
          </div>
          <div className='mt-4 flex w-full space-x-2'>
            <span className='text-xs text-medium'>*</span>
            <span className='text-xs text-medium'>
              For a limited time, Fynbos will absorb the fees associated with
              making a payment.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Source</span>
            <span className='text-medium'>{linkedAccount?.name}</span>
          </div>
          {flow?.data.note && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{flow?.data.note}</span>
            </div>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='service-agreement'
            form='pay-confirm'
            className='flex'
            aria-invalid={
              Boolean(actionData?.errors.serviceAgreement) || undefined
            }
            aria-describedby={
              actionData?.errors.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={actionData?.errors.serviceAgreement}
          >
            I authorize Fynbos to debit
            {linkedAccount?.type == 'card'
              ? ' the card indicated '
              : ' my account '}
            for the amount noted on today’s date. I will not dispute Fynbos
            debiting my account, so long as the transaction corresponds to the
            terms in this online form and my agreement with Fynbos.
          </Checkbox>
          <input
            form='pay-confirm'
            defaultValue={linkedAccount?.type}
            name='linked-account-type'
            type='hidden'
          />
        </CardContent>
      </Card>
      <Button form='pay-confirm' type='submit'>
        Confirm payment
      </Button>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const serviceAgreement = form.get('service-agreement') as string
  const linkedAccountType = form.get('linked-account-type') as string

  const fieldErrors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to authorize to continue.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  if (linkedAccountType === 'card') {
    return redirect(route('/pay/3ds'))
  }

  const clientIpAddress = getClientIP(request)

  const response = await openPaymentsClient
    .createOutgoingPayment(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress,
        threeDSID: '',
        identityType: flow.data.address.type,
        identity: flow.data.address.handle
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) throw json({}, httpMapping(response.code))

  await exitFlow(request, flowType.Pay)
  return redirect(route('/'))
}
