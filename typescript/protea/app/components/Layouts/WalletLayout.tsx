import { Outlet } from '@remix-run/react'
import { route } from 'routes-gen'
import { useState } from 'react'
import { Icon, Logo, Router } from '~/components'
import { NavDrawer } from './NavDrawer'

export function WalletLayout() {
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  return (
    <div className='inset-0 flex min-h-screen flex-col lg:flex-row'>
      <div className='fixed top-0 hidden lg:flex'>
        <NavDrawer>
          <NavDrawer.List>
            <div className='mb-8'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/activity')}>
              Activity
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/connect')}>
              Connect
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <NavDrawer.List>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
          </NavDrawer.List>
        </NavDrawer>
      </div>
      <div className='w-full lg:py-4'>
        <header className='fixed top-0 flex h-16 w-full select-none items-center justify-start bg-app p-4 sm:min-w-full lg:hidden'>
          <button
            className='-m-3 flex p-3 lg:hidden'
            onClick={() => setOpenNavModal(true)}
          >
            <Icon>menu</Icon>
          </button>
          <div className='ml-4 lg:ml-0'>
            <Router to={route('/')} aria-label='Fynbos logo'>
              <Logo className='h-8' />
            </Router>
          </div>
        </header>
        <div className='mx-4 mt-16 mb-4 min-h-full rounded-2xl bg-page lg:ml-64 lg:mt-0'>
          <Outlet />
        </div>
      </div>
      <NavDrawer.Modal open={openNavModal} setOpen={setOpenNavModal}>
        <NavDrawer>
          <NavDrawer.List>
            <div className='relative mb-8 flex items-center'>
              <button
                className='-m-3 flex p-3'
                onClick={() => setOpenNavModal(!openNavModal)}
              >
                <Icon>menu_open</Icon>
              </button>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='ml-4 h-8' />
              </Router>
            </div>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/activity')}>
              Activity
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/connect')}>
              Connect
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <NavDrawer.List>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
          </NavDrawer.List>
        </NavDrawer>
      </NavDrawer.Modal>
    </div>
  )
}
