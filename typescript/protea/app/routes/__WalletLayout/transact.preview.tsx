import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function TransactPreviewPage() {
  return (
    // <WalletLayout backRoute={Routes.transact} header='Preview' hideNav>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
