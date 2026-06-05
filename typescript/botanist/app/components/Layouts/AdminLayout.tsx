import { Outlet, href } from 'react-router'
import { useState } from 'react'
import { IconButton, Logo, Router } from '~/components'
import { NavDrawer } from './NavDrawer'

export function AdminLayout() {
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  return (
    <div className='inset-0 flex min-h-screen flex-col lg:flex-row'>
      <div className='fixed top-0 hidden lg:flex'>
        <NavDrawer>
          <NavDrawer.List>
            <div className='ml-4'>
              <Router to={href('/')} aria-label='Interledger logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            <NavDrawer.ListItem to={href('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/waitlist')}>
              Waitlist
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/wallets')}>
              Wallets
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/reviews')}>
              Reviews
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
            <span className='text-xs font-medium text-medium'>
              &copy;Interledger
            </span>
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
          <Router to={href('/')} aria-label='Interledger logo'>
            <Logo className='h-8' />
          </Router>
        </header>
        <div className='my-16 lg:my-6 lg:ml-64 lg:mr-6'>
          <Outlet />
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
              <Router to={href('/')} aria-label='Interledger logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            <NavDrawer.ListItem to={href('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/waitlist')}>
              Waitlist
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/wallets')}>
              Wallets
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={href('/reviews')}>
              Reviews
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
            <span className='text-xs font-medium text-medium'>
              &copy;Interledger
            </span>
          </footer>
        </NavDrawer>
      </NavDrawer.Modal>
    </div>
  )
}
