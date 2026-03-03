import type { Route } from './+types/live'

export const loader = async ({ request }: Route.LoaderArgs) => {
  return new Response(JSON.stringify({ status: 'alive' }), {
    status: 200,
    headers: {
      'Content-Type': 'application/json'
    }
  })
}
