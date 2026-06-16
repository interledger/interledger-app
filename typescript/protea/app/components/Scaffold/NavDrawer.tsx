import { Dialog, Transition } from '@headlessui/react'
import { motion } from 'framer-motion'
import type { Dispatch, FC, ReactNode, SetStateAction } from 'react'
import { Fragment, useEffect } from 'react'
import { NavLink, useNavigation } from 'react-router'

type ListItemProps = {
  children?: ReactNode
  to: string
  key?: string
}

const ListItem: FC<ListItemProps> = ({ children, to }) => {
  return (
    <NavLink
      prefetch='intent'
      className='mt-4 w-full rounded-xl hover:bg-nav-hover focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <li className='relative flex w-56 items-center rounded-xl p-4'>
          <span className='z-10 truncate'>{children}</span>
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

type ModalProps = {
  children?: ReactNode
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
}

const Modal: FC<ModalProps> = ({ children, open, setOpen }) => {
  const navigation = useNavigation()

  useEffect(() => {
    if (navigation.state == 'loading') setOpen(false)
  }, [setOpen, navigation.state])

  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog as='div' className='relative z-50' onClose={setOpen}>
        <Transition.Child
          as={Fragment}
          enter='ease-in-out duration-300'
          enterFrom='opacity-0'
          enterTo='opacity-100'
          leave='ease-in-out duration-300'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div className='fixed inset-0 bg-scrim/75 backdrop-blur-sm transition-opacity' />
        </Transition.Child>

        <div className='fixed inset-0 overflow-hidden'>
          <div className='absolute inset-0 overflow-hidden'>
            <div className='pointer-events-none fixed inset-y-0 left-0 flex max-w-full'>
              <Transition.Child
                as={Fragment}
                enter='transform transition ease-in-out duration-300'
                enterFrom='-translate-x-full'
                enterTo='translate-x-0'
                leave='transform transition ease-in-out duration-300'
                leaveFrom='translate-x-0'
                leaveTo='-translate-x-full'
              >
                <Dialog.Panel className='pointer-events-auto min-w-max'>
                  {children}
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  )
}

Modal.displayName = 'NavModal'

export const NavDrawer = Object.assign(NavDrawerRoot, { ListItem, List, Modal })
