import { Transition } from '@headlessui/react'
import clsx from 'clsx'
import type { FC } from 'react'
import { Fragment, useEffect } from 'react'
import { IconButton, TextButton } from './Buttons'

interface SnackbarProps {
  id: string
  show?: boolean
  // The label value.
  message?: string
  action?: string
  icon?: string
  onClose(): void
  // Offset to the right for the WalletLayout on desktop
  xOffset?: boolean
  // Offset upwards for the FAB on mobile
  yOffset?: boolean
  // ms delay after which the snackbar should be aut dismissed.
  dismissAfter?: number
}

export const Snackbar: FC<SnackbarProps> = ({
  id,
  message,
  action,
  icon,
  onClose,
  xOffset,
  yOffset,
  show = false,
  dismissAfter
}) => {
  useEffect(() => {
    let timer: NodeJS.Timeout
    if (dismissAfter && show) {
      timer = setTimeout(() => {
        onClose()
      }, dismissAfter)
    }
    return () => clearTimeout(timer)
  }, [dismissAfter, onClose, show])

  return (
    <Transition
      id={id}
      appear
      show={show}
      as={'div'}
      className={clsx(
        'fixed left-0 z-[100] mx-auto w-full overflow-y-visible lg:bottom-auto lg:top-4',
        xOffset ? 'lg:pl-64' : '',
        yOffset ? 'bottom-32' : 'bottom-4'
      )}
    >
      <div className='flex justify-center text-center'>
        <Transition.Child
          as={Fragment}
          enter='ease-out duration-300'
          enterFrom='opacity-0 scale-95'
          enterTo='opacity-100 scale-100'
          leave='ease-in duration-200'
          leaveFrom='opacity-100 scale-100'
          leaveTo='opacity-0 scale-95'
        >
          <div className='mx-4 flex w-full transform items-center justify-between space-x-3 overflow-hidden rounded-xl bg-snackbar px-4 py-3 text-left align-middle shadow-lg transition-all sm:max-w-[22rem]'>
            <p className='text-sm text-inverted'>{message}</p>
            {action && (
              <TextButton onClick={() => onClose()}>{action}</TextButton>
            )}
            {icon && (
              <div className='-mr-2'>
                <IconButton className='text-inverted' onClick={() => onClose()}>
                  {icon}
                </IconButton>
              </div>
            )}
          </div>
        </Transition.Child>
      </div>
    </Transition>
  )
}
