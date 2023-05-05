import type { LoaderArgs } from '@remix-run/node'

import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { GetWalletDetails } from '~/lib/wallet.server'
import { GridCard } from '~/components'

export async function loader({ request, params }: LoaderArgs) {
  const wallet = await GetWalletDetails(request, params.id as string)
  return json({
    wallet
  })
}

export default function Page() {
  const { wallet } = useLoaderData<typeof loader>()

  return (
    <>
      <GridCard
        className='col-span-full lg:col-span-4'
        title='Profile'
        options={wallet}
      />
    </>
  )
}
