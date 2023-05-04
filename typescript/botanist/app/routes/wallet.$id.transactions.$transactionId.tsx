import type { LoaderArgs } from '@remix-run/node'

import { Error, Grid } from '~/components'
import { json } from '@remix-run/node'
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from '@remix-run/react'
import {
  GetWalletDetails,
  GetWalletTransactionDetails
} from '~/lib/wallet.server'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)

  // const transaction = await GetWalletTransactionDetails(
  //   request,
  //   params.id as string,
  //   params.transactionId as string
  // )

  return json({
    transaction: {
      id: '123',
      amount: 100,
      currency: 'USD',
      status: 'pending'
    },
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

function ListItem({ title, body }: { title: string; body: string }) {
  return (
    <div className='flex flex-col space-y-1'>
      <dt className='text-sm font-medium text-weak capitalize'>{title}</dt>
      <dd className='text-strong'>{body || '-'}</dd>
    </div>
  )
}

export default function Page() {
  const { wallet, transaction } = useLoaderData<typeof loader>()

  return (
    <>
      <dl className='sticky top-4 col-span-full flex max-h-max h-max lg:col-span-6 flex-col space-y-4 rounded-2xl bg-page p-4'>
        {Object.entries(transaction).map(([key, value]) => (
          <ListItem key={key} title={key} body={value.toString()} />
        ))}
      </dl>
    </>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <div className='sticky top-4 col-span-full flex max-h-max h-max lg:col-span-6 flex-col space-y-4 rounded-2xl bg-page p-4'>
        <h3 className='text-5xl font-medium text-error'>{error.status}</h3>
        <h3 className='text-lg font-medium text-medium'>{error.statusText}</h3>
        <h3 className='text-lg leading-6 text-strong'>
          {JSON.stringify(error.data)}
        </h3>
      </div>
    )
  }
  return <Error data={{ title: 'error.data.message' }} />
}
