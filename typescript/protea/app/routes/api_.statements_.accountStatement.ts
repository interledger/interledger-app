import type { Route } from './+types/api_.statements_.accountStatement'

const BACKEND_HTTP_URL = process.env.BACKEND_HTTP_URL || 'http://backend:8080'

export async function loader({ request }: Route.LoaderArgs) {
  const cookies = request.headers.get('cookie') || ''

  const response = await fetch(
    `${BACKEND_HTTP_URL}/api/core/v1/statements/account`,
    { headers: { cookie: cookies } }
  )

  if (!response.ok) {
    return new Response(response.body, {
      status: response.status,
      headers: { 'Content-Type': response.headers.get('Content-Type') ?? 'application/json' }
    })
  }

  return new Response(response.body, {
    headers: {
      'Content-Type': 'application/pdf',
      'Content-Disposition': response.headers.get('Content-Disposition') ?? 'attachment; filename="account-statement.pdf"'
    }
  })
}
