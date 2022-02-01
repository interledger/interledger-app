import { Router, Routes } from './Routes'
import { FC, useEffect, useState } from 'react'
import { Header } from './Header'
import { Footer } from './Footer'
import { Logo } from './Logo'
import { Select } from './Select'
import {
  AccountIcon,
  AddIcon,
  DashboardIcon,
  GatewayIcon,
  IntegrationIcon,
  ListItemActiveIcon,
  SettingsIcon,
  WalletIcon
} from './icons'
import { Switch } from './Switch'
import type { OrgsForDashboard } from 'lib/dashboard'
import { setPreview, usePreview } from 'lib/preview'

type DashboardProps = {
  className?: string
  route?: Routes
  orgsForDashboard: OrgsForDashboard
  preview: boolean
}

export const Dashboard: FC<DashboardProps> = ({
  children,
  route,
  orgsForDashboard,
  preview
}) => {
  return (
    <>
      <div className='hidden sm:flex'>
        <SideNav
          preview={preview}
          route={route}
          orgsForDashboard={orgsForDashboard}
        />
        <div className='flex h-screen w-full max-w-3xl flex-col overflow-auto p-8'>
          {children}
        </div>
      </div>
      {/* Hide the dashboard on mobile. */}
      <div className='flex flex-col sm:hidden'>
        <Header />
        <span className='p-4'>
          The Fynbos dashboard is not available on mobile.
        </span>
        <Footer />
      </div>
    </>
  )
}

type SideNavProps = {
  route?: Routes
  orgsForDashboard: OrgsForDashboard
  preview: boolean
}

const WalletSubPages = [
  {
    name: 'Accounts',
    pathname: Routes.organisationWalletAccounts
  },
  {
    name: 'Transactions',
    pathname: Routes.organisationWalletTransactions
  },
  {
    name: 'Risk',
    pathname: Routes.organisationWalletRisk
  },
  {
    name: 'Operations',
    pathname: Routes.organisationWalletOperations
  }
]

const SideNav: FC<SideNavProps> = ({ route, orgsForDashboard, preview }) => {
  const previewDisabledInitial = orgsForDashboard.currentOrg?.verified

  const [isPreview, setIsPreview] = usePreview(preview)
  const [previewSwitchEnabled, setSwitchPreviewEnabled] = useState<boolean>(
    Boolean(previewDisabledInitial)
  )

  const togglePreview = async () => {
    const prev = await setPreview(!isPreview)
    setIsPreview(prev)
  }

  useEffect(() => {
    setSwitchPreviewEnabled(Boolean(orgsForDashboard.currentOrg?.verified))
  }, [orgsForDashboard.currentOrg?.verified])

  return (
    <div
      className={`${
        isPreview ? 'theme-test' : ''
      } sticky top-0 flex min-h-screen min-w-[300px] select-none flex-col justify-between bg-container p-4 font-display`}
    >
      <NavList>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='mb-4 ml-2 mt-2 h-8' />
        </Router>
        {/* If on an org page */}
        {orgsForDashboard.currentOrg && (
          <>
            <Select route={route} orgsForDashboard={orgsForDashboard} />
            <NavListItem
              orgId={orgsForDashboard.currentOrg.id}
              route={route}
              pathname={Routes.organisationOverview}
              icon={<DashboardIcon />}
            >
              Overview
            </NavListItem>
            <NavListItem
              orgId={orgsForDashboard.currentOrg.id}
              route={route}
              pathname={Routes.organisationIntegration}
              icon={<IntegrationIcon />}
            >
              Integration
            </NavListItem>
            <NavListItem
              orgId={orgsForDashboard.currentOrg.id}
              route={route}
              pathname={Routes.organisationSettings}
              icon={<SettingsIcon />}
            >
              Settings
            </NavListItem>
            <li className='flex items-center justify-start px-2 pt-4 text-xs font-medium uppercase text-medium'>
              products
            </li>
            <NavListItem
              orgId={orgsForDashboard.currentOrg.id}
              route={route}
              pathname={Routes.organisationGateway}
              icon={<GatewayIcon />}
            >
              Gateway
            </NavListItem>
            <NavListItem
              orgId={orgsForDashboard.currentOrg.id}
              route={route}
              pathname={Routes.organisationWallet}
              icon={<WalletIcon />}
            >
              Wallet
            </NavListItem>
            <SubNavList>
              {WalletSubPages.map((item, index) => (
                <SubNavListItem
                  key={index}
                  orgId={orgsForDashboard.currentOrg!.id}
                  route={route}
                  pathname={item.pathname}
                  icon={<ListItemActiveIcon />}
                  lastItem={index == WalletSubPages.length - 1}
                >
                  {item.name}
                </SubNavListItem>
              ))}
            </SubNavList>
          </>
        )}
        {!orgsForDashboard.currentOrg && (
          <>
            {orgsForDashboard.organisations &&
              orgsForDashboard.organisations.map((option) => {
                return (
                  <NavListItem
                    key={option?.id}
                    orgId={option?.id}
                    route={route}
                    pathname={Routes.organisationOverview}
                  >
                    {option?.name}
                  </NavListItem>
                )
              })}
            <Router href={Routes.organisation}>
              <li className='flex h-12 cursor-pointer items-center justify-between p-2 text-medium hover:bg-container-hover'>
                Add organisation <AddIcon />
              </li>
            </Router>
          </>
        )}
      </NavList>
      <NavList>
        {orgsForDashboard.currentOrg && (
          <li className='flex h-12 items-center justify-between p-2 text-medium'>
            Test data{' '}
            <Switch
              disabled={!previewSwitchEnabled}
              checked={isPreview}
              onChange={togglePreview}
            />
          </li>
        )}
        <NavListItem
          route={route}
          pathname={Routes.profile}
          icon={<AccountIcon />}
        >
          Profile
        </NavListItem>
      </NavList>
    </div>
  )
}

const NavList: FC = ({ children }) => {
  return <ul className='flex flex-col space-y-2'>{children}</ul>
}

type NavListItemProps = {
  icon?: React.ReactNode
  pathname: Routes
  orgId?: string
  route?: Routes
}

const NavListItem: FC<NavListItemProps> = ({
  children,
  icon,
  route,
  pathname,
  orgId
}) => {
  const href =
    typeof orgId == 'string'
      ? { pathname: pathname, query: { orgId: orgId } }
      : pathname
  return (
    <Router href={href}>
      <li
        className={`flex h-12 cursor-pointer items-center justify-start p-2 hover:bg-container-hover ${
          route == pathname ? 'text-primary' : 'text-medium'
        }`}
      >
        {icon && <div className='mr-2'>{icon}</div>}
        {children}
      </li>
    </Router>
  )
}

const SubNavList: FC = ({ children }) => {
  return <ul className='flex flex-col'>{children}</ul>
}

type SubNavListItemProps = {
  icon?: React.ReactNode
  pathname: Routes
  orgId?: string
  route?: Routes
  lastItem?: boolean
}

const SubNavListItem: FC<SubNavListItemProps> = ({
  children,
  icon,
  route,
  pathname,
  orgId,
  lastItem
}) => {
  const href =
    typeof orgId == 'string'
      ? { pathname: pathname, query: { orgId: orgId } }
      : pathname
  return (
    <Router href={href}>
      <li
        className={`relative mb-0 flex h-12 cursor-pointer items-center justify-start p-2 hover:bg-container-hover ${
          route == pathname ? 'text-primary' : 'text-medium'
        }`}
      >
        {icon && (
          <div
            className={`z-20 mr-2 ${
              route == pathname ? 'text-primary' : 'text-transparent'
            }`}
          >
            {icon}
          </div>
        )}
        {children}
        {!lastItem && (
          <span
            className='absolute top-[14px] left-5 z-10 -ml-px h-full w-0.5 bg-gray-300'
            aria-hidden='true'
          />
        )}
        {lastItem && (
          <span
            className='absolute bottom-[14px] left-5 z-10 -ml-px h-full w-0.5 bg-gray-300'
            aria-hidden='true'
          />
        )}
      </li>
    </Router>
  )
}
