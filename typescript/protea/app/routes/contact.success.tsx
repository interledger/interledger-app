import { ButtonRouter, Card, Layouts, SuccessShapes } from '~/components'
import { route } from 'routes-gen'
import type { MetaFunction } from '@remix-run/node'

export const handle = {
  title: 'Success',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Contact us | Success'
  }
}

export default function Page() {
  return (
    <Card>
      <SuccessShapes />

      <span className='mt-6 font-display text-2xl font-medium'>Thank you</span>
      <span className='mt-6 text-medium'>Your message has been sent.</span>
      <span className='mt-2 text-medium'>
        One of our team members will get back to you in due course.
      </span>

      <div className='flex justify-end pt-12'>
        <ButtonRouter to={route('/')}>Close</ButtonRouter>
      </div>
    </Card>
  )
}
