import { href } from 'react-router'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/api_.statements_.monthly.$year.$month'
import { envValue } from '~/env.server'

const BACKEND_HTTP_URL = envValue("BACKEND_HTTP_URL") || 'http://backend:8080'

export async function loader({ request, params }: Route.LoaderArgs) {
  const cookies = request.headers.get('cookie') || ''
  const { year, month } = params

  const response = await fetch(
    `${BACKEND_HTTP_URL}/api/core/v1/statements/monthly/${year}/${month}`,
    { headers: { cookie: cookies } }
  )

  if (!response.ok) {
    return redirectWithSnackbar(request, href('/settings/documents'), {
      message: 'Failed to download account statement.',
      icon: 'close'
    })
  }

  return new Response(response.body, {
    headers: {
      'Content-Type': 'application/pdf',
      'Cache-Control': 'private, max-age=0',
      'Content-Disposition':
        response.headers.get('Content-Disposition') ??
        'inline; filename="account-statement.pdf"'
    }
  })
}
