import { href } from 'react-router'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/api_.statements_.accountConfirmation'
import { envValue } from '~/env.server'

const BACKEND_HTTP_URL = envValue("BACKEND_HTTP_URL")

export async function loader({ request }: Route.LoaderArgs) {
  const cookies = request.headers.get('cookie') || ''

  const response = await fetch(
    `${BACKEND_HTTP_URL}/api/core/v1/statements/account-confirmation`,
    { headers: { cookie: cookies } }
  )

  if (!response.ok) {
    return redirectWithSnackbar(request, href('/settings/documents'), {
      message: 'Failed to download account confirmation statement.',
      icon: 'close'
    })
  }

  return new Response(response.body, {
    headers: {
      'Content-Type': 'application/pdf',
      'Cache-Control': 'private, max-age=0',
      'Content-Disposition':
        response.headers.get('Content-Disposition') ??
        'inline; filename="account-confirmation.pdf"'
    }
  })
}
