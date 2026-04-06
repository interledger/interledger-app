import { Dialog as HeadlessDialog, Transition } from '@headlessui/react'
import { useNavigate } from 'react-router';
import clsx from 'clsx'
import type { FC, ReactNode } from 'react'
import { Fragment, useEffect } from 'react'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

type CommandPaletteProps = {
  children?: ReactNode
  unmount?: boolean
}

export const CommandPalette: FC<CommandPaletteProps> = ({
  children,
  unmount = false
}) => {
  const navigate = useNavigate()
  const [commandPaletteOpen, setCommandPaletteOpen] = useScaffoldStore(
    (state) => [state.commandPalletOpen, state.setCommandPalletOpen]
  )

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setCommandPaletteOpen(!commandPaletteOpen)
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [commandPaletteOpen, navigate, setCommandPaletteOpen])

  return (
    <Transition.Root unmount={unmount} show={commandPaletteOpen} as={Fragment}>
      <HeadlessDialog
        unmount={unmount}
        as='div'
        className='relative z-50'
        onClose={setCommandPaletteOpen}
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

        <div className='pointer-events-none fixed inset-0 flex items-start justify-center overflow-hidden'>
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
                'pointer-events-auto mx-4 mt-4 flex w-full flex-col rounded-[1.25rem] bg-container-strong p-0 shadow-lg sm:mt-[5.5rem] sm:max-w-[29rem]'
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

CommandPalette.displayName = 'CommandPalette'
