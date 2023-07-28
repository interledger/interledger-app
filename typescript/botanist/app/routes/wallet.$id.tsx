import { Grid } from '~/components'
import { NavLink, Outlet, useParams } from '@remix-run/react'
import type { FC, ReactNode } from 'react'
import { route } from 'routes-gen'

type TabItemProps = {
  to: string
  children: ReactNode
}

const TabItem: FC<TabItemProps> = ({ children, to }) => {
  return (
    <NavLink prefetch='none' className='flex' to={to}>
      {({ isActive }) => (
        <li
          className={`flex items-center rounded-lg px-4 py-2 ${
            isActive ? 'bg-container-hover' : 'hover:bg-container'
          }`}
        >
          <span className='font-display text-xs text-medium lg:text-base'>
            {children}
          </span>
        </li>
      )}
    </NavLink>
  )
}

export default function Page() {
  const { id } = useParams()

  return (
    <Grid>
      <div className='col-span-full flex'>
        <div className='flex space-x-1 rounded-xl bg-page p-1'>
          <TabItem to={route('/wallet/:id/profile', { id: id as string })}>
            Profile
          </TabItem>
          <TabItem to={route('/wallet/:id/transactions', { id: id as string })}>
            Transactions
          </TabItem>
          <TabItem to={route('/wallet/:id/audit', { id: id as string })}>
            Audit log
          </TabItem>
          <TabItem
            to={route('/wallet/:id/linked-accounts', { id: id as string })}
          >
            Linked accounts
          </TabItem>
        </div>
      </div>
      <Outlet />
    </Grid>
  )
}
