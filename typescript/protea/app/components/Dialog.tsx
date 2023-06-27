import { Dialog as HeadlessDialog, Transition } from '@headlessui/react'
import clsx from 'clsx'
import type { Dispatch, FC, ReactNode, SetStateAction } from 'react'
import { Fragment } from 'react'

type DialogProps = {
  children?: ReactNode
  open: boolean
  unmount?: boolean
  grow?: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
}

export const Dialog: FC<DialogProps> = ({
  children,
  open,
  unmount = true,
  grow = false,
  setOpen
}) => {
  return (
    <Transition.Root unmount={unmount} show={open} as={Fragment}>
      <HeadlessDialog
        unmount={unmount}
        as='div'
        className='relative z-50'
        onClose={setOpen}
      >
        <Transition.Child
          as={Fragment}
          unmount={unmount}
          enter='ease-in-out duration-200'
          enterFrom='opacity-0'
          enterTo='opacity-100'
          leave='ease-in-out duration-200'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div className='fixed inset-0 bg-scrim/75 backdrop-blur-sm transition-opacity' />
        </Transition.Child>

        <div className='pointer-events-none fixed inset-0 flex items-center justify-center overflow-hidden'>
          <Transition.Child
            as={Fragment}
            unmount={unmount}
            enter='ease-in-out duration-200'
            enterFrom='opacity-0 scale-75'
            enterTo='opacity-100 scale-100'
            leave='ease-in-out duration-200'
            leaveFrom='opacity-100 scale-100'
            leaveTo='opacity-0 scale-75'
          >
            <HeadlessDialog.Panel
              className={clsx(
                'pointer-events-auto mx-6 flex w-full flex-col space-y-4 rounded-3xl bg-container-strong p-6 shadow-lg',
                grow ? 'sm:max-w-min' : 'sm:max-w-xs'
              )}
            >
              {children}
            </HeadlessDialog.Panel>
          </Transition.Child>
        </div>
      </HeadlessDialog>
    </Transition.Root>
  )
}

Dialog.displayName = 'Dialog'
