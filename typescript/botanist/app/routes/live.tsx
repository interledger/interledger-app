import type { LoaderArgs } from '@remix-run/node'

export const loader = async ({ request }: LoaderArgs) => {
  return new Response(JSON.stringify({ status: 'alive' }), {
    status: 200,
    headers: {
      'Content-Type': 'application/json'
    }
  })
}
