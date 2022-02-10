import { FC, Fragment } from 'react'
import { Transition } from '@headlessui/react'

type NotificationProps = {
  // The state of the notification.
  show: boolean
  // The action function.
  setShow(value: boolean): void
  // The message heading.
  header: string
  // The message body.
  body?: string
  // A verb describing the action the button on the notification commits.
  action: string
}

export const Notification: FC<NotificationProps> = ({
  show,
  setShow,
  header,
  body,
  action
}) => {
  return (
    <>
      {/* Global notification live region, render this permanently at the end of the document */}
      <div
        aria-live='assertive'
        className='pointer-events-none fixed inset-0 flex items-end px-4 py-6 sm:items-start sm:p-6'
      >
        <div className='flex w-full flex-col items-center space-y-4 sm:items-end'>
          {/* Notification panel, dynamically insert this into the live region when it needs to be displayed */}
          <Transition
            show={show}
            as={Fragment}
            enter='transform ease-out duration-300 transition'
            enterFrom='translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2'
            enterTo='translate-y-0 opacity-100 sm:translate-x-0'
            leave='transition ease-in duration-100'
            leaveFrom='opacity-100'
            leaveTo='opacity-0'
          >
            <div className='pointer-events-auto flex w-full max-w-md bg-container shadow-lg'>
              <div className='w-0 flex-1 p-4'>
                <div className='flex items-start'>
                  <div className='ml-3 w-0 flex-1'>
                    <p className='font-display font-medium text-strong'>
                      {header}
                    </p>
                    <p className='mt-1 text-sm text-weak'>{body}</p>
                  </div>
                </div>
              </div>
              <div className='flex border-l border-base'>
                <button
                  className='focus:outline-none flex w-full items-center justify-center border border-transparent p-4 text-sm font-medium text-primary focus:ring-2 focus:ring-black'
                  onClick={() => {
                    setShow(false)
                  }}
                >
                  {action}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </>
  )
}
