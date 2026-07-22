import { href } from 'react-router'
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

export const handle: ApplicationProps = {
  layout: Layouts.Focus
}

export const meta = mergeMeta(() => [
  {
    title: 'No access'
  }
])

export default function Page() {
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
        <CardLink end preventScrollReset prefetch='intent' to={href('/logout')}>
          <div className='mr-auto flex space-x-3'>
            <Icon>logout</Icon>
            <span>Log out</span>
          </div>
        </CardLink>
      </Card>
    </>
  )
}
