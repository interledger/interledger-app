import { Combobox, Transition } from '@headlessui/react'
import type { FC } from 'react'
import { Fragment } from 'react'
import { Icon } from '.'

type AutocompleteOptions = {
  id: string
  name: string
}

interface AutocompleteProps {
  id: string
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  button?: boolean
  disabled?: boolean
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string
  value?: AutocompleteOptions
  options: AutocompleteOptions[]
  onChange(value: AutocompleteOptions): void
  onQuery(query: string): void
}

/**
 * Autocomplete component
 * https://headlessui.dev/react/combobox
 * NOTE: a hidden input is required to submit to remix.
 * @param param0
 * @returns
 */
export const Autocomplete: FC<AutocompleteProps> = ({
  id,
  className,
  label,
  button = true,
  errorMessage,
  value,
  onChange,
  onQuery,
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
      <Combobox disabled={disabled} value={value} onChange={onChange}>
        <div className='relative'>
          <div className='mt-1 h-12 w-full rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0'>
            <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
              <Combobox.Input
                autoComplete='off'
                type='text'
                className='w-full overflow-hidden border-none focus:ring-0'
                displayValue={(value: AutocompleteOptions) => value?.name}
                onChange={(event) => onQuery(event.target.value)}
              />
              {button && (
                <Combobox.Button className='flex h-full items-center bg-container px-4 text-medium'>
                  <Icon>unfold_more</Icon>
                </Combobox.Button>
              )}
            </div>
          </div>
          <Transition
            as={Fragment}
            leave='transition ease-in duration-100'
            leaveFrom='opacity-100'
            leaveTo='opacity-0'
            afterLeave={() => onQuery('')}
          >
            <Combobox.Options className='absolute z-10 mt-2 max-h-60 w-full overflow-auto rounded-xl bg-container py-1 shadow-lg focus:outline-none sm:text-sm'>
              {options.length > 0 &&
                options.map((option, index) => (
                  <Combobox.Option
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
                  </Combobox.Option>
                ))}
              {options.length === 0 && (
                <div className='relative flex h-12 select-none items-center justify-between pl-4 pr-3 text-medium'>
                  Nothing found.
                </div>
              )}
            </Combobox.Options>
          </Transition>
        </div>
      </Combobox>
      <div className='h-7 pl-2 pt-2'>
        {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
      </div>
    </div>
  )
}

Autocomplete.displayName = 'Autocomplete'
