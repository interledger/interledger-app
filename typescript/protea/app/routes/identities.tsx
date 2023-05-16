import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'

export async function loader({ request }: LoaderArgs) {
  // get username from request params
  const url = new URL(request.url)
  const username = url.searchParams.get('username')

  return json({ url, username })
}

export default function Page() {
  const { url, username } = useLoaderData<typeof loader>()

  return [url, username]
}
