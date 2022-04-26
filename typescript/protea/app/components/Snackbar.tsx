import { Transition } from '@headlessui/react'
import type { FC } from 'react'
import { Fragment } from 'react'

interface SnackbarProps {
  id: string
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  show?: boolean
  // The label value.
  message?: string
  action?: string
  onClose(): void
}

export const Snackbar: FC<SnackbarProps> = ({
  id,
  className,
  message,
  action,
  onClose,
  show = false
}) => {
  return (
    <Transition
      id={id}
      appear
      show={show}
      as={'div'}
      className='fixed bottom-0 z-10 w-full overflow-y-auto sm:left-20 sm:w-[calc(100vw-5rem)] lg:left-[15.5rem] lg:w-[calc(100vw-15.5rem)]'
    >
      <div className='flex justify-center px-4 pb-6 text-center'>
        <Transition.Child
          as={Fragment}
          enter='ease-out duration-300'
          enterFrom='opacity-0 scale-95'
          enterTo='opacity-100 scale-100'
          leave='ease-in duration-200'
          leaveFrom='opacity-100 scale-100'
          leaveTo='opacity-0 scale-95'
        >
          <div className='flex w-full transform items-center justify-between space-x-2 overflow-hidden rounded-xl bg-snackbar py-1 pl-4 pr-2 text-left align-middle shadow-lg transition-all sm:max-w-sm lg:max-w-lg xl:max-w-xl'>
            <h3 className='font-sans text-sm font-normal text-white'>
              {message}
            </h3>
            <button
              className='h-10 items-center px-2 font-display text-sm font-medium capitalize text-primary focus-visible:outline-none'
              type='button'
              onClick={() => onClose()}
            >
              {action}
            </button>
          </div>
        </Transition.Child>
      </div>
    </Transition>
  )
}
