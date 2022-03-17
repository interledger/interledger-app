import type { LoaderFunction } from 'remix'
import { requireUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function ConnectPage() {
  return (
    <div className='flex h-screen w-full items-center justify-center font-display text-5xl font-medium text-medium'>
      Coming soon <span className='text-primary'>.</span>
    </div>
  )
}
