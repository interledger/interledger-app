import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  Avatar,
  Button,
  Card,
  Icon,
  Layouts,
  Router,
  TextField
} from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { generateQR, qrSvg } from '~/lib/qr.server'
import {
  getKycStatus,
  getWalletContacts,
  getWalletPaymentPointer
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const paymentPointer = await getWalletPaymentPointer(request)
  const { kycStatus } = await getKycStatus(request)

  if (kycStatus != KycStatus.Verified)
    return redirect(route('/personal-details'))

  const paymentPointerQR = qrSvg(await generateQR(paymentPointer.url))

  const contacts = (
    await getWalletContacts(request, {
      pageSize: 3,
      orderBy: 'last_paid_at desc'
    })
  ).contacts

  return json({ contacts, flow, paymentPointer, paymentPointerQR })
}

export const handle = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay' }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Pay'
  }
}

export default function Page() {
  const { contacts, flow } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='pay-payment-pointer'
        action='/pay'
        method='post'
        className='hidden'
      />
      <Card>
        <span>Enter the recipient’s payment pointer.</span>
        <TextField
          id='paymentPointer'
          label='Payment pointer'
          name='paymentPointer'
          form='pay-payment-pointer'
          type='text'
          className='mt-6'
          defaultValue={flow.data?.paymentPointer?.formatted}
          aria-invalid={Boolean(actionData?.errors.paymentPointer) || undefined}
          aria-describedby={
            actionData?.errors.paymentPointer
              ? 'paymentPointer-error'
              : undefined
          }
          errorMessage={actionData?.errors.paymentPointer}
        />
      </Card>
      <Button form='pay-payment-pointer' type='submit'>
        Pay
      </Button>
      <Card>
        <div className='flex items-center justify-between'>
          <h1 className='text-lg font-medium'>Last transacted with</h1>
          <Router className='flex max-h-fit' to={route('/contacts')}>
            <Icon className='text-medium'>read_more</Icon>
          </Router>
        </div>
        {contacts.length == 0 && (
          <div className='mt-4 flex flex-col space-y-4'>
            <span className='text-sm text-medium'>
              You haven't paid anyone yet.
            </span>
          </div>
        )}
        {contacts.map((contact, index) => (
          <button
            key={contact.id}
            name='paymentPointer'
            form='pay-payment-pointer'
            value={contact.paymentPointer}
            className='mt-6 flex flex w-full items-center space-x-3 rounded-xl'
          >
            <Avatar index={index}>{contact.name.charAt(0)}</Avatar>
            <span className='text-medium'>{contact.name}</span>
          </button>
        ))}
      </Card>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'url'

function mapper(field: fieldErrorsMap): 'paymentPointer' | null {
  switch (field) {
    case 'url':
      return 'paymentPointer'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const paymentPointer = form.get('paymentPointer') as string

  const fieldErrors = {
    form: '',
    paymentPointer: ''
  }

  const response = await openPaymentsClient
    .getPaymentPointer({ url: paymentPointer })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else if (response.code == 5) {
      fieldErrors.paymentPointer = 'Payment pointer not found.'
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  const canSendResponse = await openPaymentsClient
    .canSendToPaymentPointer(
      { paymentPointer: paymentPointer },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(canSendResponse)) {
    if (canSendResponse.code == 3) {
      for (let violation of (canSendResponse as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(canSendResponse.code))
  } else if (!canSendResponse.response.canSend) {
    fieldErrors.paymentPointer = "Payment pointer can't receive payments."
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  await updateFlow(request, flowType.Pay, {
    paymentPointer: { ...response.response }
  })

  return redirect(route('/pay/amount'))
}
