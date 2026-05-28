import { Outlet, useNavigate, href } from 'react-router'
import { IconButton, Logo, Router } from '~/components'

export function FocusLayout() {
  const navigate = useNavigate()
  return (
    <div className='relative mx-auto grid min-h-screen w-full grid-cols-4 grid-rows-[auto_1fr_auto] content-start gap-4 gap-y-2 bg-app sm:max-w-lg sm:grid-cols-8 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
      <header className='absolute top-0 col-span-full flex h-16 select-none items-center justify-start space-x-4 p-4 lg:top-16 lg:col-span-6 lg:col-start-4'>
        <IconButton onClick={() => navigate(-1)} aria-label='Back'>
          arrow_back
        </IconButton>
        <Router to={href('/')} aria-label='Fynbos logo'>
          <Logo className='h-8' />
        </Router>
      </header>
      <div className='col-span-full mt-16 px-4 sm:px-0 lg:col-span-6 lg:col-start-4 lg:mt-36'>
        <Outlet />
      </div>
      <footer className='col-span-full flex space-x-3 self-end overflow-hidden px-4 py-6 sm:px-0 lg:col-span-6 lg:col-start-4'>
        <span className='text-xs font-medium text-medium'>&copy;Fynbos</span>
      </footer>
    </div>
  )
}
