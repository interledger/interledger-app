import type { MetaFunction } from '@remix-run/node'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { ButtonRouter, Card, Layouts, SuccessShapes } from '~/components'
import { mergeMeta } from '~/lib/meta'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/contact'),
      title: 'Success'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Contact | Success'
  }
])

export default function Page() {
  return (
    <>
      <Card>
        <SuccessShapes />
        <span className='mt-6 text-medium'>Your message has been sent.</span>
        <span className='mt-2 text-medium'>
          One of our team members will get back to you in due course.
        </span>
      </Card>
      <ButtonRouter to={route('/')}>Close</ButtonRouter>
    </>
  )
}
