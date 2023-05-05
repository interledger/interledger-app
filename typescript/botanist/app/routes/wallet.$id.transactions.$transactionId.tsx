import type { LoaderArgs } from '@remix-run/node'

import { Error, GridCard, GridCardError } from '~/components'
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

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const transaction = await GetWalletTransactionDetails(
    request,
    params.id as string,
    params.transactionId as string
  )

  return json({
    transaction,
    wallet
  })
}

export default function Page() {
  const { transaction } = useLoaderData<typeof loader>()

  return (
    <GridCard
      className='sticky top-4 col-span-full lg:col-span-4'
      options={transaction}
    />
  )
}

export function ErrorBoundary() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <GridCardError
        className='sticky top-4 col-span-full lg:col-span-4'
        error={error}
      />
    )
  }
  return <Error data={{ title: 'error.data.message' }} />
}
