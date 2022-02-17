import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return await requireUserSession(request)
}

export default function ActivityTransactionPage() {
  return (
    // <WalletLayout backRoute={Routes.activity} header='Transaction' hideNav>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
