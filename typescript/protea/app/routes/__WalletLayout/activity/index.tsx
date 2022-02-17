import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return await requireUserSession(request)
}

export default function ActivityPage() {
  return (
    // <WalletLayout
    //   backRoute={Routes.walletHome}
    //   header='Activity'
    //   hideNav
    //   actionButton={{
    //     text: 'Filter',
    //     route: Routes.activityFilter,
    //     icon: <FilterIcon />
    //   }}
    // >
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
