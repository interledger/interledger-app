import { FC, Fragment } from 'react'
import { Listbox, Transition } from '@headlessui/react'
import { AddIcon, CheckIcon, SelectIcon } from './icons'

// TODO: update options to allow routing.
type Option = {
  name: string
  id: string
}

type SelectProps = {
  selected: Option
  className?: string
  options: Option[]
  onChange: any
}

export const Select: FC<SelectProps> = ({
  children,
  className,
  selected,
  options,
  onChange
}) => {
  return (
    <Listbox value={selected} onChange={onChange}>
      <div className='relative font-display'>
        <Listbox.Button className='text-strong relative w-full h-12 bg-base hover:bg-base-hover text-left cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary'>
          <span className='flex items-center'>
            <span className='ml-2 block truncate'>{selected.name}</span>
          </span>
          <span className='absolute inset-y-0 right-0 flex items-center pointer-events-none pr-2'>
            <SelectIcon />
          </span>
        </Listbox.Button>

        <Transition
          as={Fragment}
          leave='transition ease-in duration-100'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <Listbox.Options className='absolute z-10 mt-1 w-full bg-base shadow-2xl max-h-56 py-1 text-medium ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none'>
            {options.map((option) => (
              <Listbox.Option
                key={option.id}
                className={({ active }) =>
                  classNames(
                    active ? 'bg-base-hover' : 'bg-base',
                    'cursor-pointer select-none relative py-2 pl-1 pr-9 h-12 flex items-center'
                  )
                }
                value={option}
              >
                {({ selected }) => (
                  <>
                    <div className='flex items-center'>
                      <span
                        className={classNames(
                          selected ? 'text-primary' : 'text-medium',
                          'ml-3 block truncate'
                        )}
                      >
                        {option.name}
                      </span>
                    </div>
                    {selected && (
                      <span className='absolute inset-y-0 right-3 flex items-center pr-1 text-primary'>
                        <CheckIcon />
                      </span>
                    )}
                  </>
                )}
              </Listbox.Option>
            ))}
            <Listbox.Option
              key='AddOrg'
              className={({ active }) =>
                classNames(
                  active ? 'bg-base-hover' : 'bg-base',
                  'cursor-pointer select-none relative py-2 pl-1 pr-9 h-12 flex items-center'
                )
              }
              value={{
                id: 'add-organisation'
              }}
            >
              <div className='flex items-center'>
                <span className='text-medium ml-3 block truncate'>
                  Add organisation
                </span>
              </div>
              <span className='absolute inset-y-0 right-3 flex items-center pr-1'>
                <AddIcon />
              </span>
            </Listbox.Option>
          </Listbox.Options>
        </Transition>
      </div>
    </Listbox>
  )
}

function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ')
}
