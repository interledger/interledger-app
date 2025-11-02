import { Menu, Transition } from '@headlessui/react'
import clsx from 'clsx'
import { Fragment, forwardRef } from 'react'
import { Icon } from '~/components/Icon'
import { Card, CardContent } from '../Card'

interface CardActionsProps {
  showBack: boolean
  isPhysical: boolean
  flip: () => void
  isSensitiveDataVisible: boolean
  isPinVisible: boolean
  isFrozen: boolean
  toggleSensitiveDataOn: () => void
  toggleFreeze: () => void
  toggleUnfreeze: () => void
  toggleViewPin: () => void
  toggleBlock: () => void
  toggleChangePin: () => void
}

export const CardActions = ({
  isPhysical,
  showBack,
  flip,
  isFrozen,
  toggleSensitiveDataOn,
  toggleFreeze,
  toggleUnfreeze,
  toggleViewPin,
  toggleBlock,
  toggleChangePin
}: CardActionsProps) => {
  return (
    <Card>
      <CardContent>
        <div className='flex items-center justify-center space-x-2'>
          {!isFrozen && (
            <>
              <ActionButton
                onClick={() => {
                  if (showBack) {
                    flip()
                  } else {
                    toggleSensitiveDataOn()
                  }
                }}
                icon={showBack ? 'visibility_off' : 'visibility'}
                text='Details'
              />
            </>
          )}
          <ActionButton
            onClick={isFrozen ? toggleUnfreeze : toggleFreeze}
            icon={isFrozen ? 'mode_cool_off' : 'mode_cool'}
            text={isFrozen ? 'Unfreeze' : 'Freeze'}
          />

          <Menu>
            {({ open }) => (
              <>
                <Menu.Button
                  text='Settings'
                  icon='settings'
                  onClick={() => {}}
                  open={open}
                  as={ActionButton}
                />
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
                  <Menu.Items className='absolute right-0 top-full z-10 mt-2 w-56 origin-bottom-right overflow-hidden rounded-lg border border-base bg-nav shadow-lg focus:outline-none'>
                    {isPhysical && (
                      <>
                        <MenuButton
                          onClick={toggleChangePin}
                          icon='lock'
                          text='Change PIN'
                        />
                        <MenuButton
                          onClick={toggleViewPin}
                          icon='password'
                          text='View PIN'
                        />
                      </>
                    )}
                    <MenuButton
                      onClick={toggleBlock}
                      icon='delete'
                      text='Terminate'
                      dangerousAction
                    />
                  </Menu.Items>
                </Transition>
              </>
            )}
          </Menu>
        </div>
      </CardContent>
    </Card>
  )
}

interface ActionButtonProps {
  className?: string
  open?: boolean
  onClick?: () => void
  icon: string
  text: string
}

const ActionButton = forwardRef<any, ActionButtonProps>(function ActionButton(
  { open = false, onClick, icon, text, className },
  ref
) {
  return (
    <button
      ref={ref}
      className={clsx(
        open ? 'bg-nav/40 hover:bg-nav/40' : 'hover:bg-transparent',
        'group flex w-28 flex-col items-center space-y-2 rounded-xl p-3 px-2 active:bg-nav/40',
        className
      )}
      onClick={onClick}
    >
      <span
        className={clsx(
          open ? 'bg-container-hover' : 'bg-nav',
          'flex h-auto w-auto rounded-full p-2 leading-6 group-hover:bg-container-hover'
        )}
      >
        <Icon>{icon}</Icon>
      </span>
      <span className='inline-block text-xs xs:text-sm'>{text}</span>
    </button>
  )
})

interface MenuButtonProps {
  onClick: () => void
  icon: string
  text: string
  dangerousAction?: boolean
}

const MenuButton = ({
  onClick,
  icon,
  text,
  dangerousAction = false
}: MenuButtonProps) => {
  return (
    <Menu.Item>
      {({ active }) => (
        <button
          onClick={onClick}
          className={`flex w-full items-center space-x-3 px-4 py-2 text-left text-sm ${
            active ? 'bg-nav-hover' : ''
          }`}
        >
          <Icon className={dangerousAction ? 'text-error' : ''}>{icon}</Icon>
          <span>{text}</span>
        </button>
      )}
    </Menu.Item>
  )
}
