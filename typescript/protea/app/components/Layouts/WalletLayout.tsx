import { Outlet } from '@remix-run/react'
import { route } from 'routes-gen'
import { useState } from 'react'
import { Icon, IconButton, Logo, Router } from '~/components'
import { NavDrawer } from './NavDrawer'

export function WalletLayout() {
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  return (
    <div className='inset-0 flex min-h-screen flex-col lg:flex-row'>
      <div className='fixed top-0 hidden lg:flex'>
        <NavDrawer>
          <NavDrawer.List>
            <div className='ml-4'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            {/*TODO FAB*/}
            <div className='mt-10 mb-2 flex w-full space-x-3 rounded-2xl bg-primary p-4 text-white'>
              <Icon>attach_money</Icon>
              <span className='font-display font-medium'>Pay</span>
            </div>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
            <span className='text-xs font-medium text-medium'>
              &copy;Fynbos
            </span>
            <Router
              className='text-xs font-medium text-primary'
              to={route('/legal/privacy-policy')}
            >
              Privacy &amp; Terms
            </Router>
          </footer>
        </NavDrawer>
      </div>
      <div className='w-full'>
        <header className='fixed top-0 z-50 flex h-16 w-full select-none items-center justify-start space-x-4 bg-app p-4 sm:min-w-full lg:hidden'>
          <IconButton
            className='lg:hidden'
            onClick={() => setOpenNavModal(true)}
            aria-label='Open menu'
          >
            menu
          </IconButton>
          <Router to={route('/')} aria-label='Fynbos logo'>
            <Logo className='h-8' />
          </Router>
        </header>
        <div className='my-16 lg:my-[5.5rem] lg:ml-64'>
          <div className='relative mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
            <Outlet />
          </div>
        </div>
      </div>
      <NavDrawer.Modal open={openNavModal} setOpen={setOpenNavModal}>
        <NavDrawer>
          <NavDrawer.List>
            <div className='relative ml-1 mb-8 flex items-center space-x-4'>
              <IconButton
                onClick={() => setOpenNavModal(!openNavModal)}
                aria-label='Close menu'
              >
                menu_open
              </IconButton>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
            <span className='text-xs font-medium text-medium'>
              &copy;Fynbos
            </span>
            <Router
              className='text-xs font-medium text-primary'
              to={route('/legal/privacy-policy')}
            >
              Privacy &amp; Terms
            </Router>
          </footer>
        </NavDrawer>
      </NavDrawer.Modal>
    </div>
  )
}
