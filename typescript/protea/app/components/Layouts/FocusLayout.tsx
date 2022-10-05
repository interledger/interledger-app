import { Outlet, useNavigate } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Logo, Router } from '~/components'

export function FocusLayout() {
  const navigate = useNavigate()
  return (
    <div className='relative mx-auto grid min-h-screen w-full grid-cols-4 grid-rows-[auto_1fr_auto] content-start gap-4 gap-y-2 bg-app sm:max-w-lg sm:grid-cols-8 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
      <header className='absolute top-0 col-span-full flex h-16 select-none items-center justify-start p-4 lg:top-16 lg:col-span-6 lg:col-start-4'>
        <button className='-m-3 flex p-3' onClick={() => navigate(-1)}>
          <Icon>arrow_back</Icon>
        </button>
        <div className='ml-4'>
          <Router to={route('/')} aria-label='Fynbos logo'>
            <Logo className='h-8' />
          </Router>
        </div>
      </header>
      <div className='col-span-full mt-16 px-4 sm:px-0 lg:col-span-6 lg:col-start-4 lg:mt-36'>
        <Outlet />
      </div>
      <footer className='col-span-full flex space-x-3 self-end overflow-hidden px-4 pb-6 sm:px-0 lg:col-span-6 lg:col-start-4'>
        <span>&copy;Fynbos</span>
        <Router className='text-primary' to={route('/legal')}>
          Privacy &amp; Terms
        </Router>
      </footer>
    </div>
  )
}
