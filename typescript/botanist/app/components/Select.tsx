import { Listbox, Transition } from '@headlessui/react'
import type { FC } from 'react'
import { Fragment } from 'react'
import { Icon } from '.'

type SelectOptions = {
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
  disabled
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
            className='mt-1 h-12 min-w-full rounded-xl ring-2 ring-base focus:outline-none focus:ring focus:ring-focus focus-visible:ring focus-visible:ring-focus'
          >
            {({ disabled }: { disabled: boolean }) => (
              <div className='flex h-full items-center justify-between overflow-hidden rounded-xl'>
                <span className='ml-4'>{disabled ? 'Hello' : value?.name}</span>
                <div className='flex h-full items-center bg-container px-4 text-medium'>
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
            <Listbox.Options className='absolute mt-2 max-h-60 w-full overflow-auto rounded-xl bg-container py-1 shadow-lg focus:outline-none sm:text-sm'>
              {options.map((option, index) => (
                <Listbox.Option
                  key={index}
                  className={({ active }) =>
                    `relative flex h-12 cursor-pointer select-none items-center justify-between pl-4 pr-3 ${
                      active ? 'bg-container-hover' : 'text-medium'
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
            </Listbox.Options>
          </Transition>
        </div>
      </Listbox>
      <div className='h-7 pl-2 pt-2'>
        {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
      </div>
    </div>
  )
}

Select.displayName = 'Select'
