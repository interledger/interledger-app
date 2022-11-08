import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Icon, Layouts, TextField } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import {
  flowType,
  getCurrentFlow,
  requireFlow,
  updateFlow
} from '~/lib/flows.server'
import { route } from 'routes-gen'
import type { GrpcError } from '~/lib/proto.server'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { generateQR, qrSvg } from '~/lib/qr.server'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  const cookie = String(request.headers.get('cookie'))

  let response = await openPaymentsClient
    .listWalletPaymentPointers(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const qr = await generateQR(response.response.pointers[0].url)
  const svg = qrSvg(qr)

  return json({ qr: svg, paymentPointer: response.response.pointers[0] })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { qr, paymentPointer } = useLoaderData<typeof loader>()
  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>QR code</h1>
        <span>Present your qr code to receive a payment.</span>

        <div className='w-full p-12' dangerouslySetInnerHTML={{ __html: qr }} />
      </div>
      <div className='mt-6 flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>
          Payment pointer
        </h1>
        <span>Share your payment pointer to receive a payment.</span>
        <div className='mt-4 flex justify-between rounded-xl bg-container p-4'>
          <span className='font-medium text-medium'>
            {paymentPointer.formatted}
          </span>
          <Icon>share</Icon>
        </div>
      </div>
    </>
  )
}
