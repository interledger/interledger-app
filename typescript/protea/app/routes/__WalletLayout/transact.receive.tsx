import type { LoaderFunction } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
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
