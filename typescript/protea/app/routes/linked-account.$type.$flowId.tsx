import { ActionArgs, json, LoaderArgs, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { Button, Select, TextField } from '~/components'
import { Bank, Branch } from '~/generated/protobuf-ts/backend/v1/backend'
import { requireUserSession } from '~/lib/kratos.server'
import { grpcClient, GrpcError, httpMapping, isGrpcError, StatusError } from '~/lib/proto.server'

type SelectOption = {
  id: string
  name: string
}

const accountTypes = [
  { id: "CHECKING", name: "CHECKING" },
  { id: "SAVINGS", name: "SAVINGS" },
]

function formatBranchForSelect(branch: Branch) {
  return {
    id: branch.id.toString(10),
    name: branch.name
  }
}

function formatBankForSelect(bank: Bank) {
  return {
    id: bank.id.toString(10),
    name: bank.name,
    branches: bank.branches.map(formatBranchForSelect)
  }
}

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  let rpc = await grpcClient.listBanks({}, {
    meta: { cookies: String(request.headers.get('cookie')) }
  }).then(v => v).catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return json({
    banks: rpc.response.banks.map(formatBankForSelect),
    accountTypes,
    flow: {
      type: "bank",
      id: "test",
    },
  })
}

export default function Page() {
  const { banks, flow, accountTypes } = useLoaderData<typeof loader>()
  const [branches, setBranches] = useState<Array<SelectOption>>([])
  const [selectedBank, setSelectedBank] = useState<SelectOption>({ id: "-1", name: "" })
  const [selectedBranch, setSelectedBranch] = useState<SelectOption>({ id: "-1", name: "" })
  const [selectedAccountType, setSelectedAccountType] = useState<SelectOption>({ id: "-1", name: "" })

  useEffect(() => {
    setSelectedBranch({ id: "-1", name: "" })
    let bank = banks.find((bank) => bank.id === selectedBank.id)
    setBranches(bank?.branches ?? [])
  }, [selectedBank])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>

      <Form
        id='linked-account-bank-details'
        action={'/linked-account/' + flow.type + "/" + flow.id}
        method='post'
        className='hidden'
      />

      <Select
        id="bank-select"
        options={banks}
        label="Bank"
        value={selectedBank}
        onChange={setSelectedBank}
      />
      <input
        id='bank_id'
        name="bankId"
        form='linked-account-bank-details'
        value={selectedBank.id}
        type='hidden'
      />


      <div className="mt-16" />

      <Select
        id="branch-select"
        options={branches}
        label="Branch"
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
        id="account-type"
        options={accountTypes}
        label="Account type"
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
        defaultValue=""
        className='mt-6'
        required
      />

      <TextField
        id='name'
        form='linked-account-bank-details'
        label='Nickname'
        name='name'
        type='text'
        defaultValue=""
        className='mt-6'
        required
      />

      <Button className='mt-12' form='linked-account-bank-details' type='submit'>
        Submit
      </Button>
    </div>
  )
}

type fieldErrorsMap = 'BankID' | 'BranchID' | 'AccountType' | 'AccountNumber' | 'Name' | 'OTP'

function mapper(field: fieldErrorsMap): 'branchId' | 'bankId' | 'accountType' | 'accountNumber' | 'name' | 'otp' | null {
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

export async function action({ request }: ActionArgs) {
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
    otp: '',
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

  let rpc = await grpcClient.createReceiveBankAccount(
    payload,
    {
      meta: { cookies: String(request.headers.get('cookie')) },
    },
  ).then(v => v).catch(StatusError)

  if (isGrpcError(rpc)) {
    if (rpc.code == 3) {
      for (let violation of (rpc as GrpcError).details[0].fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(rpc.code))
  }

  return redirect("/settings/linked-accounts")
}
