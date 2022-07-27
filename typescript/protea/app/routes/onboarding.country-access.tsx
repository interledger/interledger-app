import { Logo } from '~/components'

export default function Page() {
  return (
    <div className='w-full'>
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-between bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <div className='flex items-center'>
          <div className='flex items-center justify-start font-display text-2xl font-medium'>
            <Logo className='h-8' />
          </div>
        </div>
      </header>
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-4xl font-medium'>
            We aren't in your country yet
          </span>
          <span className='text-medium'>
            We unfortunately don't support the country you reside in.
          </span>
          <span className='text-medium'>
            You can sign up to our waitlist, and we will let you know when
            Fynbos becomes available to you.
          </span>
        </div>
        <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          {/* TODO: Signup to waitlist button */}
        </div>
      </div>
    </div>
  )
}

// TODO: action to sign user up to the waitlist
