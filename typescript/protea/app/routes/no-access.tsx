import { href, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardIcon,
  CardLink,
  Icon,
  Layouts
} from '~/components'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/no-access'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const interactId = url.searchParams.get('interactId')
  const nonce = url.searchParams.get('nonce')
  const logoutUrl = new URL(href('/logout'), url)
  if (interactId && nonce) {
    logoutUrl.searchParams.set('returnTo', `${href('/consent')}${url.search}`)
  }

  return { logoutUrl: logoutUrl.toString() }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus
}

export const meta = mergeMeta(() => [
  {
    title: 'No access'
  }
])

export default function Page() {
  const { logoutUrl } = useLoaderData<typeof loader>()

  return (
    <>
      <Card className='flex !flex-row'>
        <CardIcon className='my-auto h-16'>
          <Icon className='text-red-600'>warning</Icon>
        </CardIcon>
        <CardContent className='ml-2'>
          You can’t complete this request with the account you’re currently
          signed in with.
        </CardContent>
      </Card>
      <Card>
        <CardLink end preventScrollReset to={logoutUrl}>
          <div className='mr-auto flex space-x-3'>
            <Icon>switch_account</Icon>
            <span>Use a different account</span>
          </div>
        </CardLink>
      </Card>
    </>
  )
}
