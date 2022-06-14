import type { LoaderFunction } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { requireNoUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  await requireNoUserSession(request)
  // Redirect to the signup flow.
  return redirect(
    route('/flows/:flowId/signup/about', {
      flowId: 'init'
    })
  )
}
