import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Checkbox, Layouts } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { requireUserSession } from '~/lib/kratos.server'
import { getClientIP } from '~/lib/ip.server'
import { getLinkedAccounts } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)
  const { linkedAccounts } = await getLinkedAccounts(request)
  return json({
    flow,
    traits: session.identity.traits,
    linkedAccount: linkedAccounts.find(
      (account) => account.id == flow.data.toLinkedAccountId
    )
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { flow, linkedAccount } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>
          Confirm payment
        </h1>
        <span>Please check the details and confirm the payment.</span>
        <div className='mt-6 flex w-full flex-col justify-between space-y-1'>
          <span className='text-sm'>To</span>
          <span className='text-sm font-medium text-strong'>
            {flow?.data.paymentPointer.formatted}
          </span>
        </div>
        <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
          <span className='text-sm'>Beneficiary Name</span>
          <span className='text-sm font-medium text-strong'>
            {flow?.data.paymentPointer.legalName}
          </span>
        </div>

        <div className='mt-6 flex w-full flex-col justify-between space-y-1'>
          <span className='text-sm'>From</span>
          <span className='text-sm font-medium text-strong'>
            {linkedAccount?.name}
          </span>
        </div>
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>You pay</span>
          <span className='text-sm font-medium text-strong'>
            {flow?.data.displaySendAmount || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            free <sup>*</sup>
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>They receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>
        {flow?.data.note && (
          <div className='mt-8 flex w-full flex-col space-y-2'>
            <span className='text-sm'>Note</span>
            <span className='text-sm text-strong'>{flow?.data.note}</span>
          </div>
        )}

        <Form
          id='pay-confirm'
          action='/pay/confirm'
          method='post'
          className='hidden'
        />
        <Checkbox
          id='service-agreement'
          name='service-agreement'
          form='pay-confirm'
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
          <Button form='pay-confirm' type='submit'>
            Confirm payment
          </Button>
        </div>
      </div>
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
  const flow = await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const serviceAgreement = form.get('service-agreement') as string

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

  const clientIpAddress = getClientIP(request)

  const response = await openPaymentsClient
    .createOutgoingPayment(
      {
        quoteID: flow.data.quoteID,
        description: flow.data.note,
        externalRef: '',
        ipAddress: clientIpAddress
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

  const transactionId = response.response.id.split('/').at(-1) as string
  await exitFlow(request, flowType.Pay)
  return redirect(
    route('/transaction/:type/:transactionId', {
      type: 'outgoing',
      transactionId
    })
  )
}
