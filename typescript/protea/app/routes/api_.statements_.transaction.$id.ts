import { href } from 'react-router'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/api_.statements_.transaction.$id'

// TODO: Move to env validation instead of defining it in every resource route.
const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'

export async function loader({ request, params }: Route.LoaderArgs) {
  const cookies = request.headers.get('cookie') || ''
  const { id } = params

  const response = await fetch(
    `${BACKEND_HTTP_URL}/api/core/v1/statements/transaction/${id}`,
    { headers: { cookie: cookies } }
  )

  if (!response.ok) {
    console.log(await response.text())
    return redirectWithSnackbar(
      request,
      href('/payments/:paymentId', { paymentId: id }),
      {
        message: 'Failed to download transaction statement statement.',
        icon: 'close'
      }
    )
  }

  return new Response(response.body, {
    headers: {
      'Content-Type': 'application/pdf',
      'Cache-Control': 'private, max-age=0',
      'Content-Disposition':
        response.headers.get('Content-Disposition') ??
        'inline; filename="transaction-statement.pdf"'
    }
  })
}
