import type { LoaderArgs } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  return requireUserSession(request)
}

export default function Page() {
  return (
    // <WalletLayout backRoute={Routes.activity} header='Filter' hideNav>
    {
      /* TODO insert content */
    }
    // </WalletLayout>
  )
}
