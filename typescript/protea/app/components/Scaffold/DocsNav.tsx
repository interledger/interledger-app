import { NavLink, useLocation } from '@remix-run/react'
import { motion } from 'framer-motion'
import type { FC, ReactNode } from 'react'
import type { DocRecord } from '~/generated/dato-cms-graphql'

type ListItemProps = {
  children?: ReactNode
  to: string
}

const ListItem: FC<ListItemProps> = ({ children, to }) => {
  let location = useLocation()
  console.log('location', location)
  console.log('to', to)
  // Need to use custom NavLink to get the active state
  // Navlink doesn't include anchor tags
  return (
    <NavLink
      prefetch='intent'
      end
      className='mt-4 w-full rounded-xl hover:bg-nav-hover focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <li className='relative flex w-56 items-center rounded-xl p-4'>
          <span className='z-10'>{children}</span>
          {isActive && (
            <motion.div
              className='absolute -ml-4 h-full w-full rounded-xl bg-nav-active'
              layoutId='nav-active'
            />
          )}
        </li>
      )}
    </NavLink>
  )
}

ListItem.displayName = 'ListItem'

type ListGroupItemsProps = {
  children?: ReactNode
  to: string
}

const ListGroupItems: FC<ListGroupItemsProps> = ({ children, to }) => {
  let location = useLocation()
  console.log('location', location)
  console.log('to', to)
  return (
    <NavLink
      prefetch='intent'
      end
      className='mt-4 w-full rounded-xl hover:bg-nav-hover focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <li className='relative flex w-56 items-center rounded-xl p-4'>
          <span className='z-10'>{children}</span>
          {isActive && (
            <motion.div
              className='absolute -ml-4 h-full w-full rounded-xl bg-nav-active'
              layoutId='nav-active'
            />
          )}
        </li>
      )}
    </NavLink>
  )
}

ListGroupItems.displayName = 'ListGroupItems'

type ListGroupProps = {
  children?: ReactNode
  group: DocRecord
}

const ListGroup: FC<ListGroupProps> = ({ group, children }) => {
  const location = useLocation()
  // Should the VisibleSectionHighlight show?
  let isActiveGroup = location.pathname?.includes(group.slug || '')
  return <ul className='flex w-full flex-col'>{children}</ul>
}

ListGroup.displayName = 'ListGroup'

type ListProps = {
  children?: ReactNode
}

const List: FC<ListProps> = ({ children }) => {
  return <ul className='flex w-full flex-col'>{children}</ul>
}

List.displayName = 'List'

type NavDrawerRootProps = {
  children?: ReactNode
}

const NavDrawerRoot: FC<NavDrawerRootProps> = ({ children }) => {
  return (
    <ul className='flex h-full min-w-max select-none flex-col justify-between bg-container-strong px-3 py-4 lg:h-screen lg:bg-page'>
      {children}
    </ul>
  )
}

NavDrawerRoot.displayName = 'NavDrawerRoot'
