import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function ConnectPage() {
  return (
    // <WalletLayout route={Routes.connect} header='Connect' settings>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
