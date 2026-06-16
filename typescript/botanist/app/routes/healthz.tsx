import type { LoaderFunctionArgs } from 'react-router'
import { getBackendHealth } from '~/lib/backend-health.server'

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const backendHealth = await getBackendHealth()

  if (!backendHealth.ok) {
    console.error(
      `Health check failed: Backend is unavailable: ${backendHealth.error}`
    )
    return new Response('Backend unavailable', { status: 503 })
  }

  return new Response('OK', { status: 200 })
}
