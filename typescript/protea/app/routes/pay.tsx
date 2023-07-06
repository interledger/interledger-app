import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  Avatar,
  Button,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardTitle,
  Icon,
  Layouts,
  Router,
  TextField
} from '~/components'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { generateQR, qrSvg } from '~/lib/qr.server'
import {
  getKycStatus,
  getWalletContacts,
  getWalletInfo
} from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.Pay)
  const walletInfo = await getWalletInfo(request)
  const { kycStatus } = await getKycStatus(request)

  if (kycStatus != KycStatus.Verified)
    return redirect(route('/personal-details'))

  const paymentPointerQR = qrSvg(await generateQR(walletInfo.url))

  const contacts = (
    await getWalletContacts(request, {
      pageSize: 3,
      orderBy: 'last_paid_at desc'
    })
  ).contacts

  return json({ contacts, flow, paymentPointerQR })
}

export const handle = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: 'Pay', back: route('/') }
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
        id='pay-address'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <span>Enter the recipient’s wallet address.</span>
          <TextField
            id='address'
            label='Wallet address'
            name='address'
            form='pay-address'
            type='text'
            className='mt-6'
            defaultValue={flow.data?.address?.walletUrl}
            aria-invalid={Boolean(actionData?.errors.address) || undefined}
            aria-describedby={
              actionData?.errors.address ? 'paymentPointer-error' : undefined
            }
            errorMessage={actionData?.errors.address}
          />
        </CardContent>
      </Card>
      <Button form='pay-address' type='submit'>
        Continue
      </Button>
      <Card>
        <CardHeader>
          <CardTitle>Last transacted with</CardTitle>
          <Router className='flex max-h-fit' to={route('/contacts')}>
            <Icon className='text-medium'>read_more</Icon>
          </Router>
        </CardHeader>
        {contacts.length == 0 && (
          <CardContent>
            <p className='text-sm text-medium'>You haven't paid anyone yet.</p>
          </CardContent>
        )}
        {contacts.map((contact, index) => (
          <CardButton
            key={contact.id}
            name='address'
            form='pay-address'
            value={contact.paymentPointer}
            className='items-center space-x-3'
          >
            <Avatar index={index}>{contact.name.charAt(0)}</Avatar>
            <span className='text-medium'>{contact.name}</span>
          </CardButton>
        ))}
      </Card>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'url'

function mapper(field: fieldErrorsMap): 'address' | null {
  switch (field) {
    case 'url':
      return 'address'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.Pay)
  const form = await request.formData()
  const address = form.get('address') as string

  const fieldErrors = {
    form: '',
    address: ''
  }

  const response = await grpcClient
    .getPaymentAddress(
      { address: address },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
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
      fieldErrors.address = 'Wallet address not found.'
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  const canSendResponse = await openPaymentsClient
    .canSendToPaymentPointer(
      { paymentPointer: response.response.walletUrl },
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
    fieldErrors.address = "Wallet address can't receive payments."
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  await updateFlow(request, flowType.Pay, {
    address: { ...response.response }
  })

  return redirect(route('/pay/amount'))
}
