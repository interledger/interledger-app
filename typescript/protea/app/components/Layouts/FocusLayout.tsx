import { Outlet, useMatches, useNavigate } from '@remix-run/react'
import { route } from 'routes-gen'
import { IconButton, FynbosLogo, Router } from '~/components'

export function FocusLayout() {
  const matches = useMatches()
  const navigate = useNavigate()
  const match = matches[matches.length - 1]
  const backTo = match?.data?.backTo

  let title
  const titleHandle = match.handle?.title
  if (typeof titleHandle === 'function') title = titleHandle(match)
  else title = titleHandle

  return (
    <div className='relative mx-auto grid min-h-screen w-full grid-cols-4 grid-rows-[auto_1fr_auto] content-start gap-4 gap-y-2 bg-page sm:max-w-lg sm:grid-cols-8 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
      <header className='absolute top-0 col-span-full flex h-16 select-none items-center justify-start space-x-4 p-4 lg:top-16 lg:col-span-6 lg:col-start-4'>
        <IconButton
          onClick={() => {
            navigate(backTo ?? -1)
          }}
          aria-label='Back'
        >
          arrow_back
        </IconButton>
        {title && <h1 className='text-xl font-medium'>{title}</h1>}
        {!title && (
          <Router to={route('/')} aria-label='Fynbos logo'>
            <FynbosLogo className='h-8' />
          </Router>
        )}
      </header>
      <div className='col-span-full mt-16 grid grid-cols-1 gap-y-6 px-4 sm:px-0 lg:col-span-6 lg:col-start-4 lg:mt-36'>
        <Outlet />
      </div>
      <footer className='col-span-full flex space-x-3 self-end overflow-hidden px-4 py-6 sm:px-0 lg:col-span-6 lg:col-start-4'>
        <span className='text-xs font-medium text-medium'>
          &copy;&nbsp;Fynbos
        </span>
        <Router
          className='text-xs font-medium text-primary'
          to={route('/legal')}
        >
          Privacy &amp; Terms
        </Router>
      </footer>
    </div>
  )
}
