import { Grid } from '~/components'
import { NavLink, Outlet, useParams, href } from 'react-router'
import type { FC, ReactNode } from 'react'

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
          <TabItem to={href('/review/:id/details', { id: id as string })}>
            Details
          </TabItem>
        </div>
      </div>
      <Outlet />
    </Grid>
  )
}
