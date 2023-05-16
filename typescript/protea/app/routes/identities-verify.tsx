import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'

export async function loader({ request }: LoaderArgs) {
  // get proof id and proof url from request params
  return json({})
}

export default function Page() {
  return []
}
