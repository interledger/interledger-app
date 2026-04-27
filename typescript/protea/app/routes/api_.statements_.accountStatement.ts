import { href } from 'react-router'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/api_.statements_.accountStatement'

const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'

export async function loader({ request }: Route.LoaderArgs) {
  const cookies = request.headers.get('cookie') || ''
  const url = new URL(request.url)
  const year = url.searchParams.get('year') ?? ''
  const month = url.searchParams.get('month') ?? ''

  const backendURL = new URL(
    `${BACKEND_HTTP_URL}/api/core/v1/statements/account-statement`
  )
  backendURL.searchParams.set('year', year)
  backendURL.searchParams.set('month', month)

  const response = await fetch(backendURL.toString(), {
    headers: { cookie: cookies }
  })

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
