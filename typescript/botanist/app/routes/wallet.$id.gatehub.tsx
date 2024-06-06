import type { LoaderArgs } from '@remix-run/node'

import { json } from '@remix-run/node'
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from '@remix-run/react'
import { Error, GridCard, GridCardError } from '~/components'
import { GetGatehubUser } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const user = await GetGatehubUser(request, params.id as string)

  return json({ user })
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
