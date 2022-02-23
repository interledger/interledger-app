import { Link, NavLink, Outlet, useLocation } from 'remix'
import { route } from 'routes-gen'
import React, { FC } from 'react'
import {
  ConnectIcon,
  SettingsIcon,
  TransactIcon,
  WalletIcon,
  Logo,
  LogoIcon
} from '~/components'

// Allow us to show/hide the NavRail and NavDrawer on certain pages.
function showNav(pathname: string) {
  return (
    ['/home', '/transact', '/connect', '/settings'].findIndex(
      (val) => val === pathname
    ) >= 0
  )
}

// Allow us to show/hide the NavBar on certain pages.
function showNavBar(pathname: string) {
  return (
    ['/home', '/transact', '/connect'].findIndex((val) => val === pathname) >= 0
  )
}

export default function WalletLayout() {
  // We useLocation here as loader only runs on first load of WalletLayout.
  const location = useLocation()
  return (
    <div className='flex flex-col text-medium sm:flex-row'>
      {showNav(location.pathname) && (
        <NavRail>
          <NavList>
            <Link to={route('/')} aria-label='Fynbos logo'>
              <LogoIcon className='mx-auto mb-4 h-8' />
            </Link>
            <NavListItem icon={<WalletIcon />} to={route('/home')}>
              Home
            </NavListItem>
            <NavListItem icon={<TransactIcon />} to={route('/transact')}>
              Transact
            </NavListItem>
            <NavListItem icon={<ConnectIcon />} to={route('/connect')}>
              Connect
            </NavListItem>
            <NavListItem icon={<SettingsIcon />} to={route('/settings')}>
              Settings
            </NavListItem>
          </NavList>
        </NavRail>
      )}
      {showNav(location.pathname) && (
        <NavDrawer>
          <NavList>
            <Link to={route('/')} aria-label='Fynbos logo'>
              <Logo className='mb-6 ml-4 h-8' />
            </Link>
            <NavListItem icon={<WalletIcon />} to={route('/home')}>
              Home
            </NavListItem>
            <NavListItem icon={<TransactIcon />} to={route('/transact')}>
              Transact
            </NavListItem>
            <NavListItem icon={<ConnectIcon />} to={route('/connect')}>
              Connect
            </NavListItem>
          </NavList>
          <NavList>
            <NavListItem icon={<SettingsIcon />} to={route('/settings')}>
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
            <NavListItem icon={<WalletIcon />} to={route('/home')}>
              Home
            </NavListItem>
            <NavListItem icon={<TransactIcon />} to={route('/transact')}>
              Transact
            </NavListItem>
            <NavListItem icon={<ConnectIcon />} to={route('/connect')}>
              Connect
            </NavListItem>
          </NavList>
        </NavBar>
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
  icon?: React.ReactNode
  to: string
}

const NavListItem: FC<NavListItemProps> = ({ children, icon, to }) => {
  return (
    <NavLink prefetch='render' className='w-full' to={to}>
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
            {icon}
          </div>
          <div>{children}</div>
        </li>
      )}
    </NavLink>
  )
}
