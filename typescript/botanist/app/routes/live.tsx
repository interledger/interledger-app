import type { LoaderFunctionArgs } from 'react-router'

export const loader = async ({ request }: LoaderFunctionArgs) => {
  return new Response(JSON.stringify({ status: 'alive' }), {
    status: 200,
    headers: {
      'Content-Type': 'application/json'
    }
  })
}
