import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { ListLinkedAccounts } from '~/lib/wallet.server'
import { GridCard } from '~/components'

export async function loader({ request, params }: LoaderArgs) {
  const linkedAccounts = await ListLinkedAccounts(request, params.id as string)

  return json({
    linkedAccounts
  })
}

export default function Page() {
  const { linkedAccounts } = useLoaderData<typeof loader>()

  return (
    <>
      {linkedAccounts.map((linkedAccount) => (
        <GridCard
          key={linkedAccount.id}
          className='col-span-full lg:col-span-4'
          title={linkedAccount.name}
          options={linkedAccount}
        />
      ))}
    </>
  )
}
