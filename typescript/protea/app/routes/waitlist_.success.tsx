import { ButtonRouter, Card, Layouts, SuccessShapes } from '~/components'
import { route } from 'routes-gen'
import type { MetaFunction } from '@remix-run/node'

export const handle = {
  title: 'Success',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Waitlist | Success'
  }
}

export default function Page() {
  return (
    <>
      <Card>
        <SuccessShapes />
        <span className='mt-6 text-medium'>
          You have successfully joined the waitlist.
        </span>
        <span className='mt-2 text-medium'>
          We will let you know via email once you are able to transact.
        </span>
      </Card>
      <ButtonRouter to={route('/')}>Close</ButtonRouter>
    </>
  )
}
