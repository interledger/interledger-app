import { Transition, Dialog } from '@headlessui/react'
import { NavLink, useNavigation } from 'react-router'
import type { Dispatch, FC, SetStateAction, ReactNode } from 'react'
import { Fragment, useEffect } from 'react'

type ListItemProps = {
  to: string
  children: ReactNode
}

const ListItem: FC<ListItemProps> = ({ children, to }) => {
  return (
    <NavLink
      end
      prefetch='intent'
      className='mt-4 w-full rounded-xl focus-visible:outline-2 focus-visible:outline-focus'
      to={to}
    >
      {({ isActive }) => (
        <li
          className={`flex w-56 items-center rounded-xl p-4 ${
            isActive ? 'bg-container-hover' : 'hover:bg-container'
          }`}
        >
          <span className='font-display'>{children}</span>
        </li>
      )}
    </NavLink>
  )
}

ListItem.displayName = 'ListItem'

const List: FC<{
  children: ReactNode
}> = ({ children }) => {
  return <ul className='flex w-full flex-col'>{children}</ul>
}

List.displayName = 'List'

const NavDrawerRoot: FC<{
  children: ReactNode
}> = ({ children }) => {
  return (
    <ul className='flex h-full min-w-max select-none flex-col justify-between bg-app py-4 px-3 lg:h-screen lg:bg-app'>
      {children}
    </ul>
  )
}

NavDrawerRoot.displayName = 'NavDrawerRoot'

type ModalProps = {
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
  children: ReactNode
}

const Modal: FC<ModalProps> = ({ children, open, setOpen }) => {
  const navigation = useNavigation()

  useEffect(() => {
    if (navigation.state === 'loading') setOpen(false)
  }, [navigation.state, setOpen])

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
          <div className='fixed inset-0 bg-slate-600/75 backdrop-blur-sm transition-opacity' />
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
