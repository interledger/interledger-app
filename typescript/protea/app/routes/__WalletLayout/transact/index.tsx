import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return await requireUserSession(request)
}

export default function TransactPage() {
  return (
    // <WalletLayout route={Routes.transact} header='Transact' settings>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
