import type { Route } from './+types/healthz'

export const loader = async ({ request }: Route.LoaderArgs) => {
  return new Response('OK', { status: 200 })
}
