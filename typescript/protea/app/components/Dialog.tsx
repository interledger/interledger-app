import { Transition, Dialog as HeadlessDialog } from '@headlessui/react'
import type { Dispatch, FC, SetStateAction, ReactNode } from 'react'
import { Fragment } from 'react'

type DialogProps = {
  children?: ReactNode
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
}

export const Dialog: FC<DialogProps> = ({ children, open, setOpen }) => {
  return (
    <Transition.Root show={open} as={Fragment}>
      <HeadlessDialog as='div' className='relative z-50' onClose={setOpen}>
        <Transition.Child
          as={Fragment}
          enter='ease-in-out duration-200'
          enterFrom='opacity-0'
          enterTo='opacity-100'
          leave='ease-in-out duration-200'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <div className='bg-scrim/75 fixed inset-0 backdrop-blur-sm transition-opacity' />
        </Transition.Child>

        <div className='pointer-events-none fixed inset-0 flex items-center justify-center overflow-hidden'>
          <Transition.Child
            as={Fragment}
            enter='ease-in-out duration-200'
            enterFrom='opacity-0 scale-75'
            enterTo='opacity-100 scale-100'
            leave='ease-in-out duration-200'
            leaveFrom='opacity-100 scale-100'
            leaveTo='opacity-0 scale-75'
          >
            <HeadlessDialog.Panel className='pointer-events-auto mx-8 flex w-full flex-col space-y-2 rounded-3xl bg-container-strong p-6 shadow-lg sm:max-w-xs'>
              {children}
            </HeadlessDialog.Panel>
          </Transition.Child>
        </div>
      </HeadlessDialog>
    </Transition.Root>
  )
}

Dialog.displayName = 'Dialog'
