import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Icon, Layouts, TextField } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import type { GrpcError } from '~/lib/proto.server'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { getWalletPaymentPointer } from '~/lib/wallet.server'
import { generateQR, qrSvg } from '~/lib/qr.server'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  const flow = await requireFlow(request, flowType.Pay)
  const paymentPointer = await getWalletPaymentPointer(request)

  const paymentPointerQR = qrSvg(await generateQR(paymentPointer.url))

  return json({ flow, paymentPointer, paymentPointerQR })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { flow, paymentPointer, paymentPointerQR } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>Pay</h1>
        <span>Enter the recipient’s payment pointer.</span>
        <Form
          id='pay-payment-pointer'
          action='/pay'
          method='post'
          className='hidden'
        />
        <TextField
          id='paymentPointer'
          label='Payment pointer'
          name='paymentPointer'
          form='pay-payment-pointer'
          type='text'
          className='mt-6'
          defaultValue={flow.data?.paymentPointer?.url}
          aria-invalid={Boolean(actionData?.errors.paymentPointer) || undefined}
          aria-describedby={
            actionData?.errors.paymentPointer
              ? 'paymentPointer-error'
              : undefined
          }
          required
          errorMessage={actionData?.errors.paymentPointer}
        />

        <div className='mt-6'>
          <Button form='pay-payment-pointer' type='submit'>
            Pay
          </Button>
        </div>
      </div>
      <div className='mt-6 flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>Receive</h1>
        <span>Present or share your payment pointer.</span>

        <div
          className='mt-8 sm:px-16'
          dangerouslySetInnerHTML={{ __html: paymentPointerQR }}
        />
        <button
          type='button'
          onClick={async () => {
            navigator.clipboard.writeText(paymentPointer.formatted).then(
              () => {
                // TODO show success snackbar.
                console.log('Successfully set payment pointer to clipboard.')
              },
              () => {
                console.log("Couldn't set payment pointer to clipboard.")
              }
            )
          }}
          className='mt-8 flex justify-between rounded-xl bg-container p-4'
        >
          <span className='font-medium text-medium'>
            {paymentPointer.formatted}
          </span>
          <Icon className='text-medium'>content_copy</Icon>
        </button>
      </div>
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

  await updateFlow(request, flowType.Pay, {
    paymentPointer: { ...response.response }
  })

  return redirect(route('/pay/amount'))
}
