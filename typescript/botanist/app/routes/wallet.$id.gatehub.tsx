import type { LoaderFunctionArgs } from 'react-router'

import { data } from 'react-router'
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from 'react-router'
import { Error, GridCard, GridCardError } from '~/components'
import { GetGatehubUser } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const user = await GetGatehubUser(request, params.id as string)

  return data({ user })
}

export default function Page() {
  const { user } = useLoaderData<typeof loader>()

  return (
    <GridCard
      className='sticky top-4 col-span-full lg:col-span-4'
      options={user}
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
