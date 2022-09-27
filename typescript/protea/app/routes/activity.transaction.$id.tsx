import type { LoaderArgs } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  return requireUserSession(request)
}

export default function Page() {
  return <div>Coming soon</div>
}
