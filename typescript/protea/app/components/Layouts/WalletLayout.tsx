import { Outlet, useMatches } from '@remix-run/react'
import { route } from 'routes-gen'
import { useState } from 'react'
import {
  Icon,
  IconButton,
  Logo,
  Router,
  WalletGrid,
  WalletShapes
} from '~/components'
import { NavDrawer } from './NavDrawer'

export function WalletLayout() {
  const [openNavModal, setOpenNavModal] = useState<boolean>(false)
  const matches = useMatches()
  const title = matches[matches.length - 1].handle?.title
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
            <Router
              to={route('/pay')}
              className='mt-10 mb-2 flex w-full space-x-3 rounded-2xl bg-primary p-4 text-white'
            >
              <Icon>attach_money</Icon>
              <span className='font-display font-medium'>Pay</span>
            </Router>
            <NavDrawer.ListItem to={route('/')}>Home</NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/transactions')}>
              Transactions
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/support')}>
              Support
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
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
        </NavDrawer>
      </div>
      <div className='w-full'>
        <header className='fixed top-0 z-50 flex h-16 w-full select-none items-center justify-start space-x-4 bg-page p-4 sm:min-w-full lg:hidden'>
          <IconButton
            className='lg:hidden'
            onClick={() => setOpenNavModal(true)}
            aria-label='Open menu'
          >
            menu
          </IconButton>
          {title && <h1 className='text-xl font-medium'>{title}</h1>}
          {!title && <Logo className='h-8' />}
        </header>
        <div className='mt-16 mb-32 lg:my-[5.5rem] lg:ml-64'>
          <div className='relative mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-[59rem]'>
            {title && (
              <WalletGrid>
                <div className='col-span-full hidden justify-between px-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4 lg:flex'>
                  <h1 className='text-2xl font-medium'>{title}</h1>
                  <WalletShapes />
                </div>
              </WalletGrid>
            )}
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
            <NavDrawer.ListItem to={route('/transactions')}>
              Transactions
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/settings')}>
              Settings
            </NavDrawer.ListItem>
            <NavDrawer.ListItem to={route('/support')}>
              Support
            </NavDrawer.ListItem>
          </NavDrawer.List>
          <footer className='flex w-full space-x-3 pl-4 pb-2'>
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
        </NavDrawer>
      </NavDrawer.Modal>
      <FAB />
    </div>
  )
}

function FAB() {
  return (
    <Router to={route('/pay')}>
      <div className='fixed bottom-4 right-4 flex h-[6rem] w-[6rem] items-center justify-center rounded-[1.75rem] bg-primary shadow-lg lg:hidden'>
        <svg
          width='36'
          height='36'
          viewBox='0 0 36 36'
          fill='none'
          xmlns='http://www.w3.org/2000/svg'
        >
          <mask
            id='mask0_1852_4585'
            style={{ maskType: 'alpha' }}
            maskUnits='userSpaceOnUse'
            x='0'
            y='0'
            width='36'
            height='36'
          >
            <rect width='36' height='36' fill='#D9D9D9' />
          </mask>
          <g mask='url(#mask0_1852_4585)'>
            <path
              d='M16.5375 31.5V28.275C15.2125 27.975 14.069 27.4 13.107 26.55C12.144 25.7 11.4375 24.5 10.9875 22.95L13.7625 21.825C14.1375 23.025 14.694 23.9375 15.432 24.5625C16.169 25.1875 17.1375 25.5 18.3375 25.5C19.3625 25.5 20.2315 25.269 20.9445 24.807C21.6565 24.344 22.0125 23.625 22.0125 22.65C22.0125 21.775 21.7375 21.081 21.1875 20.568C20.6375 20.056 19.3625 19.475 17.3625 18.825C15.2125 18.15 13.7375 17.344 12.9375 16.407C12.1375 15.469 11.7375 14.325 11.7375 12.975C11.7375 11.35 12.2625 10.0875 13.3125 9.1875C14.3625 8.2875 15.4375 7.775 16.5375 7.65V4.5H19.5375V7.65C20.7875 7.85 21.819 8.306 22.632 9.018C23.444 9.731 24.0375 10.6 24.4125 11.625L21.6375 12.825C21.3375 12.025 20.9125 11.425 20.3625 11.025C19.8125 10.625 19.0625 10.425 18.1125 10.425C17.0125 10.425 16.175 10.669 15.6 11.157C15.025 11.644 14.7375 12.25 14.7375 12.975C14.7375 13.8 15.1125 14.45 15.8625 14.925C16.6125 15.4 17.9125 15.9 19.7625 16.425C21.4875 16.925 22.794 17.7185 23.682 18.8055C24.569 19.8935 25.0125 21.15 25.0125 22.575C25.0125 24.35 24.4875 25.7 23.4375 26.625C22.3875 27.55 21.0875 28.125 19.5375 28.35V31.5H16.5375Z'
              fill='white'
            />
          </g>
        </svg>
      </div>
    </Router>
  )
}
