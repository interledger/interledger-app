import { FC, Fragment, useEffect, useState } from 'react'
import { Listbox, Transition } from '@headlessui/react'
import { AddIcon, CheckIcon, SelectIcon } from './icons'
import { OrgsForDashboard } from 'lib/dashboard'
import { useRouter } from 'next/router'
import { Routes } from 'components'

type SelectProps = {
  route?: Routes
  orgsForDashboard: OrgsForDashboard
}

export const Select: FC<SelectProps> = ({ route, orgsForDashboard }) => {
  const router = useRouter()
  const temp = orgsForDashboard.organisations.find(
    (item: any) => item?.id == orgsForDashboard.currentOrg!.id
  )
  const [currentOrg, setCurrentOrg] = useState(temp)

  // implicitly set slected on routing change.
  useEffect(() => {
    setCurrentOrg(
      orgsForDashboard.organisations.find(
        (item: any) => item?.id == orgsForDashboard.currentOrg!.id
      )
    )
  }, [orgsForDashboard])

  return (
    <Listbox
      value={currentOrg}
      onChange={(select: any) => {
        if (select.id === 'add-organisation') {
          router.push(Routes.organisation)
        } else {
          router.push({
            pathname: route,
            query: { orgId: select.id }
          })
        }
      }}
    >
      <div className='relative font-display'>
        <Listbox.Button className='focus:outline-none relative h-12 w-full cursor-pointer bg-base text-left text-strong hover:bg-base-hover focus-visible:ring-2 focus-visible:ring-primary'>
          <span className='flex items-center'>
            <span className='ml-2 block truncate'>{currentOrg?.name}</span>
          </span>
          <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2'>
            <SelectIcon />
          </span>
        </Listbox.Button>

        <Transition
          as={Fragment}
          leave='transition ease-in duration-100'
          leaveFrom='opacity-100'
          leaveTo='opacity-0'
        >
          <Listbox.Options className='focus:outline-none absolute z-10 mt-1 max-h-56 w-full overflow-auto bg-base py-1 text-medium shadow-2xl ring-1 ring-black ring-opacity-5'>
            {orgsForDashboard.organisations.map((option) => (
              <Listbox.Option
                key={option?.id}
                className={({ active }) =>
                  classNames(
                    active ? 'bg-base-hover' : 'bg-base',
                    'relative flex h-12 cursor-pointer select-none items-center py-2 pl-1 pr-9'
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
                        {option?.name}
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
                  'relative flex h-12 cursor-pointer select-none items-center py-2 pl-1 pr-9'
                )
              }
              value={{
                id: 'add-organisation'
              }}
            >
              <div className='flex items-center'>
                <span className='ml-3 block truncate text-medium'>
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
