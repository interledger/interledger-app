import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function ActivityFilterPage() {
  return (
    // <WalletLayout backRoute={Routes.activity} header='Filter' hideNav>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
