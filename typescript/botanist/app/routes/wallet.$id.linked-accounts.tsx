import type { LoaderFunctionArgs } from 'react-router'
import { data } from 'react-router'
import { useLoaderData } from 'react-router'
import { ListLinkedAccounts } from '~/lib/wallet.server'
import { GridCard } from '~/components'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const linkedAccounts = await ListLinkedAccounts(request, params.id as string)

  return data({
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
