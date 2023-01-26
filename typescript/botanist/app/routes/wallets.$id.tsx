import type { ActionArgs, LoaderArgs } from '@remix-run/node'

import { Button, TextField, WalletGrid } from '~/components'
import { json } from '@remix-run/node'
import { Form, useNavigation, useLoaderData } from '@remix-run/react'
import { GetWalletDetails } from '~/lib/wallet.server'
import { DateTime } from 'luxon'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)

  return json({
    defaultMonth: DateTime.now().toFormat('yyyy-MM'),
    // TODO: Refactor formatting into wallet.server
    wallet: {
      ...wallet,
      gender:
        wallet.gender == 0
          ? 'Unknown'
          : wallet.gender == 1
          ? 'Male'
          : wallet.gender == 2
          ? 'Female'
          : 'Other',
      dateOfBirth: DateTime.fromSeconds(
        parseInt(wallet.dateOfBirth?.seconds ?? '')
      ).toFormat('dd MMM yyyy')
    }
  })
}

export default function Page() {
  const { defaultMonth, wallet } = useLoaderData<typeof loader>()
  const navigation = useNavigation()

  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div>
          <h3 className='text-lg font-medium leading-6 text-gray-900'>
            Wallet user info
          </h3>
          <p className='mt-1 mb-5 max-w-2xl text-sm text-gray-500'>
            Personal details and wallet information of a specified wallet.
          </p>
        </div>
        <div className='border-t border-gray-200'>
          <dl className='sm:divide-y sm:divide-gray-200'>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>Full name</dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.firstName + ' ' + wallet.lastName}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>Email</dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.users[0].email}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>
                Phone number
              </dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.users[0].phoneNumber}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>Country</dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.countryCode}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>Address</dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.address}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>
                Date of birth
              </dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.dateOfBirth}
              </dd>
            </div>
            <div className='py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:px-6'>
              <dt className='text-sm font-medium text-gray-500'>gender</dt>
              <dd className='mt-1 text-sm text-gray-900 sm:col-span-2 sm:mt-0'>
                {wallet.gender}
              </dd>
            </div>
          </dl>
        </div>
      </div>
      <div className='col-span-6 flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div>
          <h3 className='text-lg font-medium leading-6 text-gray-900'>
            Statement
          </h3>
          <p className='mt-1 mb-5 max-w-2xl text-sm text-gray-500'>
            Send the user a copy of their statement for a given month.
          </p>
        </div>
        <Form
          id='statement'
          action={`/wallets/${wallet.walletID}`}
          method='post'
          className='hidden'
        />
        <TextField
          id='month'
          label='Month'
          name='month'
          form='statement'
          className='mt-6'
          defaultValue={defaultMonth}
          max={defaultMonth}
          type='month'
        />
        <Button form='statement' type='submit'>
          {navigation.state == 'submitting' ? 'Sending...' : 'Send email'}
        </Button>
      </div>
    </WalletGrid>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Period'

function mapper(field: fieldErrorsMap): 'month' | null {
  switch (field) {
    case 'Period':
      return 'month'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const month = form.get('month') as string
  console.log('month', month)

  const fieldErrors = {
    form: '',
    month: ''
  }
  const response = await grpcClient
    .emailWalletStatement(
      {
        period: month,
        walletID: params.id as string
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
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  return json({ errors: { ...fieldErrors } }, { status: 200 })
}
