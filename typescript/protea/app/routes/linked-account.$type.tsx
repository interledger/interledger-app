import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useParams } from '@remix-run/react'
import { getUserSession } from '~/lib/kratos.server'
import type { RadioGroupOption } from '~/components'
import { Button, Card, Layouts, RadioGroup, Shape } from '~/components'
import { flowType, requireFlow } from '~/lib/flows.server'
import { useState } from 'react'
import { route } from 'routes-gen'

export async function loader({ request }: LoaderArgs) {
  await getUserSession(request)
  return null
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Add linked account'
  }
}

export default function Page() {
  const params = useParams()
  switch (params.type) {
    case 'card':
      return <CardPage />
    case 'bank':
      return <BankPage />
    default:
      return <GenericPage />
  }
}

const options = [
  {
    id: 'card',
    name: 'Debit card',
    icon: 'credit_card'
  },
  {
    id: 'bank',
    name: 'Bank account',
    icon: 'account_balance'
  }
]

function GenericPage() {
  const [value, setValue] = useState<RadioGroupOption>(options[0])

  return (
    <Card>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Add linked account
      </h1>
      <span>Select the type of account to add.</span>

      <RadioGroup
        id={'type'}
        className='mt-6'
        label='Account type'
        value={value}
        onChange={setValue}
        options={options}
      />
      <input form='link-card' value={value.id} name='radioType' type='hidden' />
      <Form
        id='link-card'
        action={'/linked-account/new'}
        method='post'
        className='hidden'
      />
      <div className='mt-12'>
        <Button form='link-card' type='submit'>
          Continue
        </Button>
      </div>
    </Card>
  )
}

function CardPage() {
  return (
    <Card>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Top up cash balance
      </h1>
      <span>
        Add a debit card in order to top up your cash balance and transact.
      </span>
      <div className='mt-10 flex items-start'>
        <Shape width={'w-8'} radius={'rounded-tl-full'} color={'bg-lime-500'} />
        <Shape
          width={'w-8'}
          radius={'rounded-tl-full'}
          color={'bg-slate-600'}
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
      <div className='mt-12'>
        <Button form='link-card' type='submit'>
          Let's go
        </Button>
      </div>
    </Card>
  )
}

function BankPage() {
  return (
    <Card>
      <h1 className='mb-6 font-display text-2xl font-medium'>Withdraw</h1>
      <span>You first need to add a bank account to make withdrawals.</span>
      <div className='mt-10 flex items-start'>
        <Shape
          flex='flex-none'
          width={'w-8'}
          radius={'rounded-tr-full'}
          color={'bg-slate-500'}
        />
        <Shape
          flex='flex-none'
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
      <div className='mt-12'>
        <Button form='link-bank' type='submit'>
          Let's go
        </Button>
      </div>
    </Card>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const radioType = form.get('radioType') as string
  let flow
  if (params.type == 'card') {
    flow = await requireFlow(request, flowType.LinkCardAccount, {
      startRoute: route('/linked-account/:type/widget', { type: 'card' }),
      data: {},
      returnTo: route('/deposit')
    })
  } else if (params.type == 'new' && radioType == 'card') {
    flow = await requireFlow(request, flowType.LinkCardAccount)
  } else if (params.type == 'bank') {
    flow = await requireFlow(request, flowType.LinkBankAccount, {
      startRoute: route('/linked-account/:type/widget', { type: 'bank' }),
      data: {},
      returnTo: route('/withdraw')
    })
  } else if (params.type == 'new' && radioType == 'bank') {
    flow = await requireFlow(request, flowType.LinkBankAccount)
  } else
    throw json(
      { title: `Linking type ${params.type} not allowed.` },
      { status: 400 }
    )

  return redirect(flow.startRoute)
}
