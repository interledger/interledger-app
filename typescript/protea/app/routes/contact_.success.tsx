import type { MetaFunction } from '@remix-run/node'
import { route } from 'routes-gen'
import { ButtonRouter, Card, Layouts, SuccessShapes } from '~/components'

export const handle = {
  title: 'Success',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Contact us | Success'
  }
}

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
