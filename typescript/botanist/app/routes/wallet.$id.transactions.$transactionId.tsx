import type { LoaderFunctionArgs } from 'react-router'

import { data } from 'react-router'
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from 'react-router'
import { Error, GridCard, GridCardError } from '~/components'
import {
  GetWalletDetails,
  GetWalletTransactionDetails,
  ListExternalApiCalls
} from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  const transaction = await GetWalletTransactionDetails(
    request,
    params.id as string,
    params.transactionId as string
  )
  const externalApiLogs = await ListExternalApiCalls(
    request,
    transaction?.transaction?.paymentId || ''
  )

  return data({
    transaction,
    wallet,
    externalApiLogs
  })
}

export default function Page() {
  const { transaction, externalApiLogs } = useLoaderData<typeof loader>()

  return (
    <>
      <GridCard
        className='sticky top-4 col-span-full lg:col-span-4'
        options={transaction}
      />
      <GridCard
        className='sticky top-4 col-span-full lg:col-span-4'
        options={externalApiLogs}
      />
    </>
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
