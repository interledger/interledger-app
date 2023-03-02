import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Card, Checkbox, Layouts } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import { getLinkedAccounts } from '~/lib/wallet.server'
import { getClientIP } from '~/lib/ip.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.TopUp)
  const { linkedAccounts } = await getLinkedAccounts(request)
  return json({
    flow,
    linkedAccount: linkedAccounts.find(
      (account) => account.id == flow.data.toLinkedAccountId
    )
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Top up | Confirm'
  }
}

export default function Page() {
  const { flow, linkedAccount } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Card>
        <h1 className='mb-6 font-display text-2xl font-medium'>
          Confirm top up
        </h1>
        <span>Please check the details and confirm your top up.</span>

        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>Top up from:</span>
          <span className='text-sm font-medium text-strong'>
            {linkedAccount?.name}
          </span>
        </div>
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>Top up amount</span>
          <span className='text-sm font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            Free<sup>*</sup>
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>You receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>

        <Form
          id='deposit-confirm'
          action='/deposit/confirm'
          method='post'
          className='hidden'
        />
        <Checkbox
          id='service-agreement'
          name='service-agreement'
          form='deposit-confirm'
          className='mt-8 flex'
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
          I authorize Fynbos to debit the card indicated for the amount noted on
          today’s date. I will not dispute Fynbos debiting my account, so long
          as the transaction corresponds to the terms in this online form and my
          agreement with Fynbos.
        </Checkbox>
        <div className='mt-6'>
          <Button form='deposit-confirm' type='submit'>
            Confirm payment
          </Button>
        </div>
      </Card>
      <div className='mt-6 flex w-full space-x-2'>
        <span className='text-xs text-medium'>*</span>
        <span className='text-xs text-medium'>
          For a limited time, Fynbos will absorb the fees associated with making
          a payment.
        </span>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const flow = await requireFlow(request, flowType.TopUp)
  const form = await request.formData()
  const serviceAgreement = form.get('service-agreement') as string

  const clientIpAddress = getClientIP(request)

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
  const response = await grpcClient
    .startMachnetWalletTopup(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        fromLinkedAccountId: flow.data.toLinkedAccountId,
        amount: flow.data.receiveAmount,
        ipAddress: clientIpAddress,
        currency: 'USD'
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
    throw json({}, httpMapping(response.code))
  }
  await exitFlow(request, flowType.TopUp)

  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Top up created successfully.',
        icon: 'close'
      })
    }
  })
}
