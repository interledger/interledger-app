import type { MetaFunction } from '@remix-run/node'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  ButtonRouter,
  Card,
  CardContent,
  Layouts,
  SuccessShapes
} from '~/components'
import { mergeMeta } from '~/lib/meta'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Success'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Waitlist | Success'
  }
])

export default function Page() {
  return (
    <>
      <Card>
        <CardContent>
          <SuccessShapes />
          <span className='mt-6 text-medium'>
            You have successfully joined the waitlist.
          </span>
          <span className='mt-2 text-medium'>
            We will let you know via email once you are able to transact.
          </span>
        </CardContent>
      </Card>
      <ButtonRouter to={route('/')}>Close</ButtonRouter>
    </>
  )
}
