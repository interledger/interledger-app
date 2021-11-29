import { Router, Routes } from './Routes'
import { FC, useState } from 'react'
import { Header } from './Header'
import { Footer } from './Footer'
import { useRouter } from 'next/router'
import { Logo } from './Logo'
import { Select } from './Select'
import {
  AccountIcon,
  AddIcon,
  DashboardIcon,
  GatewayIcon,
  IntegrationIcon,
  SettingsIcon,
  WalletIcon
} from './icons'
import { Switch } from './Switch'
import type { OrgsForDashboard } from 'lib/dashboard'

type DashboardProps = {
  className?: string
  route?: Routes
  orgsForDashboard: OrgsForDashboard
  isTest?: boolean
}

export const Dashboard: FC<DashboardProps> = ({
  children,
  route,
  orgsForDashboard,
}) => {
  return (
    <>
      <div className='hidden sm:flex'>
        <SideNav
          route={route}
          orgsForDashboard={orgsForDashboard}
        />
        <div className='flex flex-col max-w-3xl p-8 w-full h-screen overflow-auto'>
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
}

// TODO: Get options from backend.
const options = [
  {
    id: '1234',
    name: 'GateHub'
  },
  {
    id: '2341',
    name: "Bob's biscuits"
  }
]

const SideNav: FC<SideNavProps> = ({ route, orgsForDashboard }) => {
  const router = useRouter()
  const [enabled, setEnabled] = useState(false)
  const [selected, setSelected] = useState(
    options.filter((item) => item.id === orgId)[0]
  )

  return (
    <div
      className={`${
        enabled ? 'theme-test' : ''
      } font-display select-none bg-base sticky top-0 flex flex-col justify-between min-w-[300px] min-h-screen p-4`}
    >
      <NavList>
        <Router href={Routes.home} aria-label='Fynbos logo'>
          <Logo className='h-8 mb-4 ml-2 mt-2' />
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
            <li className='flex justify-start items-center pt-4 px-2 text-medium uppercase font-medium text-xs'>
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
              <li className='flex justify-between items-center h-12 p-2 hover:bg-base-hover text-medium cursor-pointer'>
                Add organisation <AddIcon />
              </li>
            </Router>
          </>
        )}
      </NavList>
      <NavList>
        {orgsForDashboard.currentOrg && (
          <li className='flex items-center h-12 p-2 text-medium justify-between'>
            Test data <Switch enabled={enabled} onChange={setEnabled} />
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
        className={`flex justify-start items-center h-12 p-2 hover:bg-base-hover cursor-pointer ${
          route == pathname ? 'text-primary' : 'text-medium'
        }`}
      >
        {icon && <div className='mr-2'>{icon}</div>}
        {children}
      </li>
    </Router>
  )
}
