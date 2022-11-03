import type { LoaderArgs, ActionArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useParams } from '@remix-run/react'
import { requireUserSession } from '~/lib/kratos.server'
import { Button, Layouts, Shape } from '~/components'
import { route } from 'routes-gen'
import {
  flowType,
  getCurrentFlow,
  requireFlow,
  updateFlow
} from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  await requireUserSession(request)
  let headers

  const response = await grpcClient
    .hasSendUser(
      {},
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

  const hasSendUser = response.response.hasSendUser
  if (!hasSendUser) {
    headers = await requireFlow(request, flowType.PersonalDetails)
  } else if (params.type == 'card') {
    headers = await requireFlow(request, flowType.LinkCardAccount)
  } else if (params.type == 'bank') {
    headers = await requireFlow(request, flowType.LinkBankAccount)
  } else
    throw json(
      { title: `Linking type ${params.type} not allowed.` },
      { status: 400 }
    )

  return json({ hasSendUser }, { headers })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const params = useParams()
  switch (params.type) {
    case 'card':
      return <CardPage />
    case 'bank':
      return <BankPage />
    default:
      return null
  }
}

function CardPage() {
  const { hasSendUser } = useLoaderData<typeof loader>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Add a send account
      </h1>
      <span>Here’s what we will need to add a send account.</span>
      {!hasSendUser && (
        <>
          <div className='mt-6 flex items-start'>
            <Shape
              width={'w-8'}
              radius={'rounded-br-full'}
              color={'bg-rose-300'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-lime-500'}
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Personal details</h3>
              <p className='text-sm text-medium'>Date of birth and gender.</p>
            </div>
          </div>
          <div className='mt-10 flex items-start'>
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-yellow-300'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-bl-full'}
              color={'bg-slate-300'}
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Address details</h3>
              <p className='text-sm text-medium'>Your physical address.</p>
            </div>
          </div>
        </>
      )}
      <div className='mt-10 flex items-start'>
        <Shape width={'w-8'} radius={'rounded-br-full'} color={'bg-lime-300'} />
        <Shape
          width={'w-8'}
          radius={'rounded-br-full'}
          color={'bg-slate-500'}
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Debit card</h3>
          <p className='text-sm text-medium'>Your debit card details.</p>
        </div>
      </div>

      <Form
        id='link-card'
        action={'/linked-account/card'}
        method='post'
        className='hidden'
      />
      <input
        form='link-card'
        value={hasSendUser ? 'hasSendUser' : undefined}
        name='hasSendUser'
        type='hidden'
      />
      <div className='mt-12'>
        <Button form='link-card' type='submit'>
          Let's go
        </Button>
      </div>
    </div>
  )
}

function BankPage() {
  const { hasSendUser } = useLoaderData<typeof loader>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Add a receive account
      </h1>
      <span>Here's what we will need to add a receive account.</span>
      {!hasSendUser && (
        <>
          <div className='mt-6 flex items-start'>
            <Shape
              width={'w-8'}
              radius={'rounded-br-full'}
              color={'bg-rose-300'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-lime-500'}
            />
            <div className='ml-5'>
              <h3 className='mb-1 font-medium text-strong'>Personal details</h3>
              <p className='text-sm text-medium'>Date of birth and gender.</p>
            </div>
          </div>
        </>
      )}
      <div className='mt-10 flex items-start'>
        <Shape
          width={'w-8'}
          radius={'rounded-tr-full'}
          color={'bg-slate-500'}
        />
        <Shape
          width={'w-8'}
          radius={'rounded-br-full'}
          color={'bg-yellow-300'}
        />
        <div className='ml-5'>
          <h3 className='mb-1 font-medium text-strong'>Bank details</h3>
          <p className='text-sm text-medium'>
            We will retrieve you bank information with your permission via a
            secure connection.
          </p>
        </div>
      </div>

      <Form
        id='link-bank'
        action={'/linked-account/bank'}
        method='post'
        className='hidden'
      />
      <input
        form='link-bank'
        value={hasSendUser ? 'hasSendUser' : undefined}
        name='hasSendUser'
        type='hidden'
      />
      <div className='mt-12'>
        <Button form='link-bank' type='submit'>
          Let's go
        </Button>
      </div>
    </div>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const hasSendUser = form.get('hasSendUser') as string

  let flow
  if (!hasSendUser) {
    flow = await getCurrentFlow(request, flowType.PersonalDetails)
    const headers = await updateFlow(request, flowType.PersonalDetails, {
      returnTo: route('/linked-account/:type/:flowId', {
        type: params.type as string,
        flowId: flow.id
      })
    })

    return redirect(
      route('/personal-details/:flowId/about', {
        flowId: flow.id
      }),
      {
        headers
      }
    )
  } else if (params.type == 'card') {
    flow = await getCurrentFlow(request, flowType.LinkCardAccount)
  } else if (params.type == 'bank') {
    flow = await getCurrentFlow(request, flowType.LinkBankAccount)
  } else
    throw json(
      { title: `Linking type ${params.type} not allowed.` },
      { status: 400 }
    )

  return redirect(
    route('/linked-account/:type/:flowId', {
      type: params.type,
      flowId: flow.id
    })
  )
}
