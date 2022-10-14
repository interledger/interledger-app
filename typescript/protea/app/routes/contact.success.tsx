import { ButtonRouter, SuccessShapes } from '~/components'
import { route } from 'routes-gen'

export default function Page() {
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <SuccessShapes />

      <span className='mt-6 font-display text-2xl font-medium'>Thank you</span>
      <span className='mt-6 text-medium'>Your message has been sent.</span>
      <span className='mt-2 text-medium'>
        One of our team members will get back to you in due course.
      </span>

      <div className='flex justify-end pt-12'>
        <ButtonRouter to={route('/')}>Close</ButtonRouter>
      </div>
    </div>
  )
}
