import type { LoaderArgs } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  return requireUserSession(request)
}

export default function Page() {
  return (
    <div className='flex h-screen w-full items-center justify-center font-display text-5xl font-medium text-medium'>
      Coming soon <span className='text-primary'>.</span>
    </div>
  )
}
