import { Menu, Transition } from '@headlessui/react'
import { Fragment } from 'react'
import { Icon } from '~/components/Icon'

interface CardActionsProps {
  showBack: boolean
  setShowBack: (show: boolean) => void
  isSensitiveDataVisible: boolean
  isPinVisible: boolean
  isLocked: boolean
  toggleSensitiveData: () => void
  toggleLock: () => void
  toggleUnlock: () => void
  toggleViewPin: () => void
  toggleBlock: () => void
  toggleTerminate: () => void
}

export const CardActions = ({
  showBack,
  setShowBack,
  isLocked,
  toggleSensitiveData,
  toggleLock,
  toggleUnlock,
  toggleViewPin,
  toggleBlock,
  toggleTerminate
}: CardActionsProps) => {
  return (
    <div className='flex items-center space-x-4'>
      {/* Flip card */}
      <button
        className='flex items-center justify-center space-x-2 rounded-lg bg-blue-500 px-4 py-2 text-white transition-colors hover:bg-blue-600'
        onClick={() => setShowBack(!showBack)}
      >
        <Icon>{showBack ? 'flip_to_front' : 'flip_to_back'}</Icon>
        <span>Flip</span>
      </button>

      {/* View Dropdown */}
      <Menu as='div' className='relative'>
        {({ open }) => (
          <>
            <Menu.Button className='flex items-center justify-center space-x-2 rounded-lg bg-teal-600 px-4 py-2 text-white transition-colors hover:bg-teal-700 focus:outline-none focus:ring-2 focus:ring-teal-500 focus:ring-offset-2'>
              <Icon>visibility</Icon>
              <Icon className='text-lg'>
                {open ? 'expand_less' : 'expand_more'}
              </Icon>
            </Menu.Button>

            <Transition
              as={Fragment}
              show={open}
              enter='transition ease-out duration-100'
              enterFrom='transform opacity-0 scale-95'
              enterTo='transform opacity-100 scale-100'
              leave='transition ease-in duration-75'
              leaveFrom='transform opacity-100 scale-100'
              leaveTo='transform opacity-0 scale-95'
            >
              <Menu.Items className='absolute left-0 z-10 mt-2 w-48 origin-top-left rounded-lg border border-gray-200 bg-white shadow-lg focus:outline-none'>
                <div className='py-1'>
                  {/* Details */}
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        onClick={toggleSensitiveData}
                        className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
                          active ? 'bg-teal-50 text-teal-900' : 'text-gray-700'
                        }`}
                      >
                        <Icon className='text-teal-600'>numbers</Icon>
                        <span>Details</span>
                      </button>
                    )}
                  </Menu.Item>

                  {/* PIN */}
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        onClick={toggleViewPin}
                        className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
                          active
                            ? 'bg-purple-50 text-purple-900'
                            : 'text-gray-700'
                        }`}
                      >
                        <Icon className='text-purple-600'>password</Icon>
                        <span>PIN</span>
                      </button>
                    )}
                  </Menu.Item>
                </div>
              </Menu.Items>
            </Transition>
          </>
        )}
      </Menu>

      {/* More Actions Dropdown */}
      <Menu as='div' className='relative'>
        {({ open }) => (
          <>
            <Menu.Button className='flex items-center justify-center space-x-2 rounded-lg bg-gray-700 px-4 py-2 text-white transition-colors hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2'>
              <Icon>lock</Icon>
              <Icon className='text-lg'>
                {open ? 'expand_less' : 'expand_more'}
              </Icon>
            </Menu.Button>

            <Transition
              as={Fragment}
              show={open}
              enter='transition ease-out duration-100'
              enterFrom='transform opacity-0 scale-95'
              enterTo='transform opacity-100 scale-100'
              leave='transition ease-in duration-75'
              leaveFrom='transform opacity-100 scale-100'
              leaveTo='transform opacity-0 scale-95'
            >
              <Menu.Items className='absolute right-0 z-10 mt-2 w-56 origin-top-right rounded-lg border border-gray-200 bg-white shadow-lg focus:outline-none'>
                <div className='py-1'>
                  {/* Lock/Unlock */}
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        onClick={isLocked ? toggleUnlock : toggleLock}
                        className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
                          active ? 'bg-red-50 text-red-900' : 'text-gray-700'
                        }`}
                      >
                        <Icon className='text-red-600'>lock</Icon>
                        <span>{isLocked ? 'Unlock' : 'Lock'}</span>
                      </button>
                    )}
                  </Menu.Item>

                  {/* Divider */}
                  <div className='my-1 h-px bg-gray-200' />

                  {/* Block */}
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        onClick={toggleBlock}
                        className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
                          active
                            ? 'bg-orange-50 text-orange-900'
                            : 'text-gray-700'
                        }`}
                      >
                        <Icon className='text-orange-600'>block</Icon>
                        <span>Block</span>
                      </button>
                    )}
                  </Menu.Item>

                  {/* Terminate */}
                  <Menu.Item>
                    {({ active }) => (
                      <button
                        onClick={toggleTerminate}
                        className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
                          active ? 'bg-red-50 text-red-900' : 'text-red-700'
                        }`}
                      >
                        <Icon className='text-red-600'>delete</Icon>
                        <span>Terminate</span>
                      </button>
                    )}
                  </Menu.Item>
                </div>
              </Menu.Items>
            </Transition>
          </>
        )}
      </Menu>
    </div>
  )
}
