import type { LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { requireNoUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  // Redirect to the signup flow.
  return redirect(
    route('/flows/:flowId/signup/about', {
      flowId: 'init'
    })
  )
}
