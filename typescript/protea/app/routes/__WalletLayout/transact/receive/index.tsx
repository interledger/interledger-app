import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return await requireUserSession(request)
}

export default function TransactReceivePage() {
  return (
    // <WalletLayout backRoute={Routes.transact} header='Receive' hideNav>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
