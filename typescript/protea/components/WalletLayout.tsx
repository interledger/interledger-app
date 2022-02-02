import { Router, Routes } from './Routes'
import React, { FC } from 'react'
import {
  BackIcon,
  ConnectIcon,
  SettingsIcon,
  TransactIcon,
  WalletIcon
} from './icons'
import { Logo, LogoIcon } from './Logo'
import { Button } from './Button'
import { useRouter } from 'next/router'
import { FAB } from 'components'

type WalletLayoutProps = {
  className?: string
  route?: Routes
  backRoute?: Routes
  actionButton?: {
    text: string
    route: Routes
    icon?: React.ReactNode
  }
  header: string
  settings?: boolean
  hideNav?: boolean
}

/**
 * WalletLayout provides structure for application pages:
 * - Navigation: https://www.figma.com/file/uvoEXzq2nCCLqv2ifPHAjN/Components?node-id=6%3A925
 * - A grid layout for content: https://www.figma.com/file/uvoEXzq2nCCLqv2ifPHAjN/Components?node-id=5%3A1339
 *
 * @param children The children to be placed in the content grid.
 * @param route - The current Route. Used to highlight the correct NavItem.
 * @param backRoute - [Header] The href of the back button. Shows button when defined.
 * @param button - [Header] Shows button when defined, provides necessary info and routing. TODO replace with FAB and include mobile.
 * @param header - [Header] The text rendered in the header.
 * @param settings - [Header] Whether to render the settings icon. Will only show on mobile.
 * @param hideNav - Whether to render the NavX components.
 */
export const WalletLayout: FC<WalletLayoutProps> = ({
  children,
  route,
  backRoute,
  actionButton,
  header,
  settings,
  hideNav
}) => {
  const router = useRouter()
  return (
    <>
      <div className='flex flex-col text-medium sm:flex-row'>
        {!hideNav && (
          <NavRail>
            <NavList>
              <Router href={Routes.home} aria-label='Fynbos logo'>
                <LogoIcon className='mx-auto mb-4 h-8' />
              </Router>
              <NavListItem
                icon={<WalletIcon />}
                pathname={Routes.walletHome}
                route={route}
              >
                Home
              </NavListItem>
              <NavListItem
                icon={<TransactIcon />}
                pathname={Routes.transact}
                route={route}
              >
                Transact
              </NavListItem>
              <NavListItem
                icon={<ConnectIcon />}
                pathname={Routes.connect}
                route={route}
              >
                Connect
              </NavListItem>
              <NavListItem
                icon={<SettingsIcon />}
                pathname={Routes.settings}
                route={route}
              >
                Settings
              </NavListItem>
            </NavList>
          </NavRail>
        )}
        {!hideNav && (
          <NavDrawer>
            <NavList>
              <Router href={Routes.home} aria-label='Fynbos logo'>
                <Logo className='mb-6 ml-4 h-8' />
              </Router>
              <NavListItem
                icon={<WalletIcon />}
                pathname={Routes.walletHome}
                route={route}
              >
                Home
              </NavListItem>
              <NavListItem
                icon={<TransactIcon />}
                pathname={Routes.transact}
                route={route}
              >
                Transact
              </NavListItem>
              <NavListItem
                icon={<ConnectIcon />}
                pathname={Routes.connect}
                route={route}
              >
                Connect
              </NavListItem>
            </NavList>
            <NavList>
              <NavListItem
                icon={<SettingsIcon />}
                pathname={Routes.settings}
                route={route}
              >
                Settings
              </NavListItem>
            </NavList>
          </NavDrawer>
        )}
        <div className='w-full'>
          {/* Header */}
          <header
            className={`sticky top-0 flex h-16 select-none items-center justify-between bg-white p-4 text-medium ${
              hideNav
                ? 'mx-auto w-full sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'
                : 'min-w-full'
            }`}
          >
            <div className='flex items-center justify-start font-display text-2xl font-medium'>
              {backRoute && (
                <Router href={backRoute}>
                  <div className='-ml-3 p-3 text-medium'>
                    <BackIcon />
                  </div>
                </Router>
              )}
              {header}
            </div>
            {settings && (
              <Router className='sm:hidden' href={Routes.settings}>
                <div className='-mr-3 p-3 text-medium'>
                  <SettingsIcon />
                </div>
              </Router>
            )}
            {actionButton && (
              <div className='hidden lg:flex'>
                <Button
                  onClick={() => router.push(actionButton.route)}
                  icon={actionButton.icon}
                >
                  {actionButton.text}
                </Button>
              </div>
            )}
          </header>
          {/* Body */}
          <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
            {children}
          </div>
        </div>
        {!hideNav && (
          <NavBar>
            <NavList>
              <NavListItem
                icon={<WalletIcon />}
                pathname={Routes.walletHome}
                route={route}
              >
                Home
              </NavListItem>
              <NavListItem
                icon={<TransactIcon />}
                pathname={Routes.transact}
                route={route}
              >
                Transact
              </NavListItem>
              <NavListItem
                icon={<ConnectIcon />}
                pathname={Routes.connect}
                route={route}
              >
                Connect
              </NavListItem>
            </NavList>
          </NavBar>
        )}
      </div>
      {actionButton && (
        <FAB
          hasNav={!hideNav}
          onClick={() => router.push(actionButton.route)}
          icon={actionButton.icon}
        >
          {actionButton.text}
        </FAB>
      )}
    </>
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
    <div className='sticky bottom-0 flex h-20 min-w-full select-none justify-between bg-container-primary p-4 pt-3 font-display text-xs sm:hidden'>
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
  pathname: Routes
  route?: Routes
}

const NavListItem: FC<NavListItemProps> = ({
  children,
  icon,
  route,
  pathname
}) => {
  return (
    <Router className='w-full' href={pathname}>
      <li
        className={`flex w-full flex-col items-center justify-between space-y-1 rounded-full sm:w-14 sm:space-y-2 lg:h-12 lg:w-56 lg:flex-row lg:justify-start lg:space-y-0 lg:space-x-3 lg:px-4 lg:hover:bg-container-hover ${
          route == pathname ? 'ring-focus lg:ring-2' : ''
        }`}
      >
        <div
          className={`flex h-8 w-16 items-center justify-center rounded-full sm:w-14 lg:h-6 lg:w-6 ${
            route == pathname
              ? 'border-2 border-focus text-primary lg:border-0'
              : 'text-medium'
          }`}
        >
          {icon}
        </div>
        <div>{children}</div>
      </li>
    </Router>
  )
}
