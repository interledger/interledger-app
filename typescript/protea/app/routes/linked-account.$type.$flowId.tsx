import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useNavigate, useParams } from '@remix-run/react'
import { useCallback, useEffect, useState } from 'react'
import { useScript } from '~/lib/useScript'
import { requireUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { Button, Layouts, Select, Shape, TextField } from '~/components'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  await requireUserSession(request)

  let card, bank

  let cardRpc = await grpcClient
    .getMachnetWidgetToken(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(cardRpc)) {
    throw json({}, httpMapping(cardRpc.code))
  }

  card = {
    widgetScriptUrl: 'https://widget.v4sandbox.machpay.com/widget/widget.js',
    widgetUserId: cardRpc.response.userId,
    widgetToken: cardRpc.response.value
  }

  let bankRpc = await grpcClient
    .listBanks(
      {},
      {
        meta: { cookies: String(request.headers.get('cookie')) }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(bankRpc)) {
    throw json({}, httpMapping(bankRpc.code))
  }
  bank = {
    banks: bankRpc.response.banks.map((bank) => ({
      id: bank.id.toString(10),
      name: bank.name,
      branches: bank.branches.map((branch) => ({
        id: branch.id.toString(10),
        name: branch.name
      }))
    })),
    accountTypes: [
      { id: 'CHECKING', name: 'CHECKING' },
      { id: 'SAVINGS', name: 'SAVINGS' }
    ]
  }

  return json({
    card,
    bank
  })
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
  const { card } = useLoaderData<typeof loader>()

  if (!card) throw Error('No card ')

  const scriptStatus = useScript(card.widgetScriptUrl)
  const navigate = useNavigate()
  const params = useParams()

  const listener = useCallback(
    (event: any) => {
      if (event.data.type == 'CARD' && event.data.status == 'CARD_ADDED') {
        navigate(
          route('/linked-account/:type/:flowId/success', {
            type: params.type as string,
            flowId: params.flowId as string
          })
        )
      }
    },
    [navigate, params.flowId, params.type]
  )

  useEffect(() => {
    if (scriptStatus === 'ready') {
      const widget = new (window as any).MachnetWidget({
        elementId: 'widget',
        userId: card.widgetUserId,
        width: '100%',
        height: '200px',
        type: 'card',
        locale: 'en',
        stylesheet: '',
        token: card.widgetToken
      })
      widget.init()

      window.addEventListener('message', listener)
    }
  }, [card.widgetToken, card.widgetUserId, listener, scriptStatus])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Debit card</h1>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-br-full'}
            color={'bg-lime-300'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-br-full'}
            color={'bg-slate-500'}
          />
        </div>
      </div>
      <p className='mt-6 text-medium'>
        Please provide your debit card details.
      </p>
      <div id='widget' className='mt-6 w-full' />
    </div>
  )
}

type SelectOption = {
  id: string
  name: string
}

function BankPage() {
  const params = useParams()
  const { bank } = useLoaderData<typeof loader>()
  const [branches, setBranches] = useState<Array<SelectOption>>([])
  const [selectedBank, setSelectedBank] = useState<SelectOption>({
    id: '-1',
    name: ''
  })
  const [selectedBranch, setSelectedBranch] = useState<SelectOption>({
    id: '-1',
    name: ''
  })
  const [selectedAccountType, setSelectedAccountType] = useState<SelectOption>({
    id: '-1',
    name: ''
  })

  const onChangeBank = useCallback(
    (newBank) => {
      setSelectedBank(newBank)
      setSelectedBranch({ id: '-1', name: '' })
      let currentBank = bank.banks.find((bank) => bank.id === newBank.id)
      setBranches(currentBank?.branches ?? [])
    },
    [bank.banks]
  )

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Bank details</h1>
        <div className='hidden sm:flex'>
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
        </div>
      </div>
      <p className='mt-6 text-medium'>
        Please provide your debit card details.
      </p>
      <Form
        id='linked-account-bank-details'
        action={`/linked-account/${params.type}/${params.flowId}`}
        method='post'
        className='hidden'
      />

      <Select
        id='bank-select'
        options={bank.banks}
        label='Bank'
        value={selectedBank}
        onChange={onChangeBank}
      />
      <input
        id='bank_id'
        name='bankId'
        form='linked-account-bank-details'
        value={selectedBank.id}
        type='hidden'
      />

      <div className='mt-16' />

      <Select
        id='branch-select'
        options={branches}
        label='Branch'
        value={selectedBranch}
        disabled={false}
        onChange={setSelectedBranch}
      />
      <input
        id='branch_id'
        name='branchId'
        form='linked-account-bank-details'
        value={selectedBranch.id}
        type='hidden'
      />

      <Select
        id='account-type'
        options={bank.accountTypes}
        label='Account type'
        disabled={false}
        value={selectedAccountType}
        onChange={setSelectedAccountType}
      />
      <input
        id='account_type'
        name='accountType'
        form='linked-account-bank-details'
        value={selectedAccountType.id}
        type='hidden'
      />

      <TextField
        id='account-number'
        form='linked-account-bank-details'
        label='Account number'
        name='accountNumber'
        type='text'
        defaultValue=''
        className='mt-6'
        required
      />

      <TextField
        id='name'
        form='linked-account-bank-details'
        label='Nickname'
        name='name'
        type='text'
        defaultValue=''
        className='mt-6'
        required
      />

      <Button
        className='mt-12'
        form='linked-account-bank-details'
        type='submit'
      >
        Submit
      </Button>
    </div>
  )
}

type fieldErrorsMap =
  | 'BankID'
  | 'BranchID'
  | 'AccountType'
  | 'AccountNumber'
  | 'Name'
  | 'OTP'

function mapper(
  field: fieldErrorsMap
):
  | 'branchId'
  | 'bankId'
  | 'accountType'
  | 'accountNumber'
  | 'name'
  | 'otp'
  | null {
  switch (field) {
    case 'BankID':
      return 'bankId'
    case 'BranchID':
      return 'branchId'
    case 'AccountType':
      return 'accountType'
    case 'AccountNumber':
      return 'accountNumber'
    case 'Name':
      return 'name'
    case 'OTP':
      return 'otp'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const accountType = form.get('accountType') as string
  const accountNumber = form.get('accountNumber') as string
  const bankId = form.get('bankId') as string
  const branchId = form.get('branchId') as string
  const name = form.get('name') as string
  const fieldErrors = {
    accountType: '',
    accountNumber: '',
    name: '',
    bankId: '',
    branchId: '',
    otp: ''
  }

  // TODO: check it is valid selections

  const payload = {
    accountType,
    accountNumber,
    name,
    bankId: parseInt(bankId),
    branchId: parseInt(branchId),
    otp: '' // TODO
    // TODO: csrf token
  }

  let rpc = await grpcClient
    .createReceiveBankAccount(payload, {
      meta: { cookies: String(request.headers.get('cookie')) }
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    if (rpc.code == 3) {
      for (let violation of (rpc as GrpcError).details[0].fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(rpc.code))
  }

  return redirect(
    route('/linked-account/:type/:flowId/success', {
      type: params.type as string,
      flowId: params.flowId as string
    })
  )
}
