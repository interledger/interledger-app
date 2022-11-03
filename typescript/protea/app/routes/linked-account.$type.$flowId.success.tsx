import { ButtonRouter, Layouts, SuccessShapes } from '~/components'
import { route } from 'routes-gen'
import { useParams } from '@remix-run/react'

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const params = useParams()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <SuccessShapes />

      <span className='mt-6 font-display text-2xl font-medium'>Success</span>
      <span className='mt-6 text-medium'>
        Your {params.type == 'card' ? 'debit card' : 'bank account'} has been
        added.
      </span>

      <div className='flex justify-end pt-12'>
        <ButtonRouter to={route('/settings/linked-accounts')}>
          Close
        </ButtonRouter>
      </div>
    </div>
  )
}
