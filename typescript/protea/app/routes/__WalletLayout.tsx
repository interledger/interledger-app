import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useNavigate
} from '@remix-run/react'
import { route } from 'routes-gen'
import type { FC } from 'react'
import { Fragment } from 'react'
import { Icon, Logo, LogoIcon, Router } from '~/components'
import { Menu, Popover, Transition } from '@headlessui/react'

// Allow us to show/hide the NavRail and NavDrawer on certain pages.
function showNav(pathname: string) {
  return (
    ['/', '/activity', '/connect', '/settings'].findIndex(
      (val) => val === pathname
    ) >= 0
  )
}

// Allow us to show/hide the NavBar on certain pages.
function showNavBar(pathname: string) {
  return (
    ['/', '/activity', '/connect'].findIndex((val) => val === pathname) >= 0
  )
}

export default function Page() {
  // We useLocation here as loader only runs on first load.
  const location = useLocation()
  return (
    <div className='flex flex-col text-medium sm:flex-row'>
      {showNav(location.pathname) && (
        <NavRail>
          <NavList>
            <Link to={route('/')} aria-label='Fynbos logo'>
              <LogoIcon className='mx-auto mb-4 h-8' />
            </Link>
            <NavFAB />
            <NavListItem icon='savings' to={route('/')}>
              Home
            </NavListItem>
            <NavListItem icon='history' to={route('/activity')}>
              Activity
            </NavListItem>
            <NavListItem icon='dashboard_customize' to={route('/connect')}>
              Connect
            </NavListItem>
            <NavListItem icon='settings' to={route('/settings')}>
              Settings
            </NavListItem>
          </NavList>
        </NavRail>
      )}
      {showNav(location.pathname) && (
        <NavDrawer>
          <NavList>
            <div className='mb-2 ml-4'>
              <Router to={route('/')} aria-label='Fynbos logo'>
                <Logo className='h-8' />
              </Router>
            </div>
            <div className='pb-6'>
              <NavFAB />
            </div>
            <NavListItem icon='savings' to={route('/')}>
              Home
            </NavListItem>
            <NavListItem icon='history' to={route('/activity')}>
              Activity
            </NavListItem>
            <NavListItem icon='dashboard_customize' to={route('/connect')}>
              Connect
            </NavListItem>
          </NavList>
          <NavList>
            <NavListItem icon='settings' to={route('/settings')}>
              Settings
            </NavListItem>
          </NavList>
        </NavDrawer>
      )}
      {/* The header and body of the page should be inserted here by the page. */}
      <Outlet />
      {showNavBar(location.pathname) && (
        <NavBar>
          <NavList>
            <NavListItem icon='savings' to={route('/')}>
              Home
            </NavListItem>
            <NavListItem icon='history' to={route('/activity')}>
              Activity
            </NavListItem>
            <NavListItem icon='dashboard_customize' to={route('/connect')}>
              Connect
            </NavListItem>
          </NavList>
        </NavBar>
      )}
      {(location.pathname == '/' || location.pathname == '/activity') && (
        <HomeFAB />
      )}
    </div>
  )
}

const NavList: FC = ({ children }) => {
  return (
    <ul className='flex w-full space-x-4 sm:flex-col sm:space-x-0 sm:space-y-6 lg:space-y-2'>
      {children}
    </ul>
  )
}

const NavBar: FC = ({ children }) => {
  return (
    <div className='fixed bottom-0 flex h-20 min-w-full select-none justify-between bg-container-primary p-4 pt-3 font-display text-xs sm:hidden'>
      {children}
    </div>
  )
}

const NavRail: FC = ({ children }) => {
  return (
    <ul className='sticky top-0 hidden h-screen min-w-max select-none flex-col justify-between bg-container p-4 px-3 pt-4 font-display text-xs sm:flex lg:hidden'>
      {children}
    </ul>
  )
}

const NavDrawer: FC = ({ children }) => {
  return (
    <ul className='sticky top-0 hidden h-screen min-w-max select-none flex-col justify-between bg-container p-4 px-3 pt-4 font-display text-base lg:flex'>
      {children}
    </ul>
  )
}

type NavListItemProps = {
  icon?: string
  to: string
}

const NavListItem: FC<NavListItemProps> = ({ children, icon, to }) => {
  return (
    <NavLink
      prefetch='render'
      className='w-full rounded-full focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <li
          className={`flex w-full flex-col items-center justify-between space-y-1 rounded-full sm:w-14 sm:space-y-2 lg:h-12 lg:w-56 lg:flex-row lg:justify-start lg:space-y-0 lg:space-x-3 lg:px-4 lg:hover:bg-container-hover ${
            isActive ? 'ring-focus lg:ring-2' : ''
          }`}
        >
          <div
            className={`flex h-8 w-16 items-center justify-center rounded-full sm:w-14 lg:h-6 lg:w-6 ${
              isActive
                ? 'border-2 border-focus text-primary lg:border-0'
                : 'text-medium'
            }`}
          >
            <Icon>{icon}</Icon>
          </div>
          <div>{children}</div>
        </li>
      )}
    </NavLink>
  )
}

function NavFAB() {
  return (
    <Menu as='div'>
      <div>
        <Menu.Button className='inline-flex cursor-pointer items-center space-x-3 rounded-2xl bg-container-primary p-4 text-medium hover:bg-container-primary-hover focus-visible:outline-2 focus-visible:outline-focus lg:w-full'>
          <Icon>swap_horiz</Icon>
          <span className='hidden lg:flex'>Transact</span>
        </Menu.Button>
      </div>
      <Transition
        as={Fragment}
        enter='transition ease-out duration-100'
        enterFrom='transform opacity-0 scale-95'
        enterTo='transform opacity-100 scale-100'
        leave='transition ease-in duration-75'
        leaveFrom='transform opacity-100 scale-100'
        leaveTo='transform opacity-0 scale-95'
      >
        <Menu.Items className='absolute ml-20 -mt-14 w-56 origin-top-left rounded-xl bg-container shadow-lg focus:outline-none lg:ml-[15.5rem]'>
          <Menu.Item>
            {({ active }) => (
              <Router
                className={`flex w-full items-center space-x-3 px-4 py-3 text-sm first-of-type:rounded-t-xl last-of-type:rounded-b-xl hover:bg-container-hover ${
                  active ? 'bg-container-hover' : ''
                }`}
                to={route('/flows/:flowId/send/to', {
                  flowId: 'init'
                })}
              >
                <Icon>send</Icon>
                <span>Send</span>
              </Router>
            )}
          </Menu.Item>
          <Menu.Item>
            {({ active }) => (
              <Router
                className={`flex w-full items-center space-x-3 px-4 py-3 text-sm first-of-type:rounded-t-xl last-of-type:rounded-b-xl hover:bg-container-hover ${
                  active ? 'bg-container-hover' : ''
                }`}
                to={route('/receive')}
              >
                <Icon>qr_code</Icon>
                <span>Recieve</span>
              </Router>
            )}
          </Menu.Item>
          <Menu.Item>
            {({ active }) => (
              <Router
                className={`flex w-full items-center space-x-3 px-4 py-3 text-sm first-of-type:rounded-t-xl last-of-type:rounded-b-xl hover:bg-container-hover ${
                  active ? 'bg-container-hover' : ''
                }`}
                to={route('/flows/:flowId/deposit/payment-method', {
                  flowId: 'init'
                })}
              >
                <Icon>download</Icon>
                <span>Deposit</span>
              </Router>
            )}
          </Menu.Item>
          <Menu.Item>
            {({ active }) => (
              <Router
                className={`flex w-full items-center space-x-3 px-4 py-3 text-sm first-of-type:rounded-t-xl last-of-type:rounded-b-xl hover:bg-container-hover ${
                  active ? 'bg-container-hover' : ''
                }`}
                to={route('/flows/:flowId/withdraw/payment-method', {
                  flowId: 'init'
                })}
              >
                <Icon>upload</Icon>
                <span>Withdraw</span>
              </Router>
            )}
          </Menu.Item>
        </Menu.Items>
      </Transition>
    </Menu>
  )
}

function HomeFAB() {
  // TODO Scroll state for extended FAB
  const navigate = useNavigate()
  return (
    <div className='fixed top-16 w-full max-w-sm px-4'>
      <Popover className='relative'>
        {({ open }) => (
          <>
            <Popover.Button
              onClick={(event: any) => {
                if (open) {
                  event.preventDefault()
                  navigate(
                    route('/flows/:flowId/send/to', {
                      flowId: 'init'
                    })
                  )
                }
              }}
              className='fixed right-4 bottom-24 font-display text-sm font-medium transition-all sm:hidden'
            >
              {!open && (
                <div className='flex w-min cursor-pointer items-center space-x-3 rounded-2xl bg-container-primary-active p-4 text-medium shadow-lg focus-visible:outline-2 focus-visible:outline-focus'>
                  <Icon>swap_horiz</Icon>
                  <span>Transact</span>
                </div>
              )}
              {open && (
                <div className='flex items-center space-x-4'>
                  <span>Send</span>
                  <div className='flex w-min cursor-pointer items-center space-x-3 rounded-2xl bg-primary p-4 text-white shadow-lg focus-visible:outline-2 focus-visible:outline-focus'>
                    <Icon>send</Icon>
                  </div>
                </div>
              )}
            </Popover.Button>
            <Transition as={Fragment}>
              <Popover.Panel className='fixed right-6 bottom-[168px] z-10 flex flex-col-reverse space-y-4 space-y-reverse font-display text-sm font-medium '>
                <Transition.Child
                  as={Fragment}
                  enter='transition ease-out duration-200'
                  enterFrom='opacity-0 translate-y-2'
                  enterTo='opacity-100 translate-y-0'
                  leave='transition ease-in duration-150'
                  leaveFrom='opacity-100 translate-y-0'
                  leaveTo='opacity-0 translate-y-2'
                >
                  <div className='flex items-center justify-end space-x-6'>
                    <span>Receive</span>
                    <Router
                      to={route('/receive')}
                      className='flex items-center justify-center rounded-xl bg-container-primary p-2 shadow-lg'
                    >
                      <Icon>qr_code</Icon>
                    </Router>
                  </div>
                </Transition.Child>
                <Transition.Child
                  as={Fragment}
                  enter='transition ease-out duration-200 delay-[25ms]'
                  enterFrom='opacity-0 translate-y-2'
                  enterTo='opacity-100 translate-y-0'
                  leave='transition ease-in duration-150'
                  leaveFrom='opacity-100 translate-y-0'
                  leaveTo='opacity-0 translate-y-2'
                >
                  <div className='flex items-center justify-end space-x-6'>
                    <span>Deposit</span>
                    <Router
                      to={route('/flows/:flowId/deposit/payment-method', {
                        flowId: 'init'
                      })}
                      className='flex items-center justify-center rounded-xl bg-container-primary p-2 shadow-lg'
                    >
                      <Icon>download</Icon>
                    </Router>
                  </div>
                </Transition.Child>
                <Transition.Child
                  as={Fragment}
                  enter='transition ease-out duration-200 delay-[50ms]'
                  enterFrom='opacity-0 translate-y-2'
                  enterTo='opacity-100 translate-y-0'
                  leave='transition ease-in duration-150'
                  leaveFrom='opacity-100 translate-y-0'
                  leaveTo='opacity-0 translate-y-2'
                >
                  <div className='flex items-center justify-end space-x-6'>
                    <span>Withdraw</span>
                    <Router
                      to={route('/flows/:flowId/withdraw/payment-method', {
                        flowId: 'init'
                      })}
                      className='flex items-center justify-center rounded-xl bg-container-primary p-2 shadow-lg'
                    >
                      <Icon>upload</Icon>
                    </Router>
                  </div>
                </Transition.Child>
              </Popover.Panel>
            </Transition>
            <Transition.Child
              as={Fragment}
              enter='ease-out duration-300'
              enterFrom='opacity-0'
              enterTo='opacity-100'
              leave='ease-in duration-200'
              leaveFrom='opacity-100'
              leaveTo='opacity-0'
            >
              <Popover.Overlay className='fixed inset-0 -z-10 bg-app/90' />
            </Transition.Child>
          </>
        )}
      </Popover>
    </div>
  )
}
