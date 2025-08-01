import { Listbox, Transition } from '@headlessui/react'
import { Link } from '@remix-run/react'
import type { RemixLinkProps } from '@remix-run/react/dist/components'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import type { FC, ReactNode, RefAttributes } from 'react'
import { Fragment, forwardRef } from 'react'
import { Icon } from '.'

export type SelectOptions = {
  id: string
  name: string
}

interface SelectProps {
  id: string
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  disabled?: boolean
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string
  value?: SelectOptions
  options: SelectOptions[]
  onChange(value: SelectOptions): void
  selectButton?: ReactNode
}

/**
 * Select component should be used in place of <select>
 * NOTE: a hidden input is required to submit to remix.
 * @param param0
 * @returns
 */
export const Select: FC<SelectProps> = ({
  id,
  className,
  label,
  errorMessage,
  value,
  onChange,
  options,
  disabled,
  selectButton
}) => {
  return (
    <div className={className || 'min-w-full'}>
      <label
        htmlFor={id}
        className='ml-2 block text-sm font-medium text-medium'
      >
        {label}
      </label>
      <Listbox disabled={disabled} value={value} onChange={onChange}>
        <div className='relative'>
          <Listbox.Button
            id={id}
            className='mt-1 h-12 min-w-full rounded-xl border-2 border-base focus:border-focus focus:outline-none focus-visible:border-focus focus-visible:ring-0'
          >
            {({ disabled }: { disabled: boolean }) => (
              <div className='flex h-full items-center justify-between overflow-hidden rounded-xl'>
                <span className='px-4 text-left text-sm'>
                  {JSON.stringify(value)}
                  {disabled ? '' : value?.name}
                </span>
                <div className='flex h-full items-center bg-nav px-4 text-medium'>
                  <Icon>unfold_more</Icon>
                </div>
              </div>
            )}
          </Listbox.Button>
          <Transition
            as={Fragment}
            leave='transition ease-in duration-100'
            leaveFrom='opacity-100'
            leaveTo='opacity-0'
          >
            <Listbox.Options className='absolute z-10 mt-2 max-h-60 w-full overflow-auto rounded-xl bg-nav p-1 text-sm shadow-lg focus:outline-none'>
              {options.map((option, index) => (
                <Listbox.Option
                  key={index}
                  className={({ active }) =>
                    `relative flex h-12 cursor-pointer select-none items-center justify-between rounded-lg pl-4 pr-3 ${
                      active ? 'bg-nav-hover' : 'text-medium'
                    }`
                  }
                  value={option}
                >
                  {({ selected }) => (
                    <>
                      <span className={`block truncate`}>{option.name}</span>
                      {selected && (
                        <span className='flex text-primary'>
                          <Icon>check</Icon>
                        </span>
                      )}
                    </>
                  )}
                </Listbox.Option>
              ))}
              {selectButton && selectButton}
            </Listbox.Options>
          </Transition>
        </div>
      </Listbox>
      <AnimatePresence>
        {errorMessage && (
          <motion.div
            animate={{ opacity: 1, y: 0 }}
            initial={{ opacity: 0, y: -8 }}
            exit={{
              opacity: 0,
              y: -8,
              transition: {
                duration: 0.2
              }
            }}
            transition={{
              type: 'spring',
              stiffness: 400,
              damping: 20,
              duration: 0.3
            }}
            className='h-7 pl-2 pt-2'
          >
            {errorMessage && (
              <p className='text-sm text-error'>{errorMessage}</p>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

Select.displayName = 'Select'

interface SelectRouterProps
  extends RemixLinkProps,
    RefAttributes<HTMLAnchorElement> {
  className?: string
  to: string
  children?: ReactNode
}

export const SelectRouter = forwardRef<any, SelectRouterProps>(
  ({ className, to, children, ...props }, ref) => {
    return (
      <Link
        ref={ref}
        to={to}
        className={clsx(
          'relative flex h-12 w-full cursor-pointer select-none items-center justify-between rounded-lg pl-4 pr-3 hover:bg-nav-hover',
          className
        )}
        {...props}
      >
        {children}
      </Link>
    )
  }
)
SelectRouter.displayName = 'SelectRouter'
