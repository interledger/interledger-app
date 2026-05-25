import { Listbox, Transition } from '@headlessui/react'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import type { InputHTMLAttributes, ReactNode } from 'react'
import { Fragment, forwardRef, useImperativeHandle, useRef } from 'react'
import { href } from 'react-router'
import { Icon, SelectRouter } from '~/components'
import type { FormattedLinkedAccount } from '~/data/accounts.server'

interface PayTextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string

  prefixIcon?: ReactNode
  showConnectAccount?: boolean

  linkedAccount?: FormattedLinkedAccount
  linkedAccountOptions: FormattedLinkedAccount[]
  onChangeLinkedAccount: (value: FormattedLinkedAccount) => void
}

/**
 * Phone text field component with country select.
 * Includes two input fields linked to the form specified in the props.
 *
 * References:
 * https://headlessui.dev/react/Listbox
 * https://gitlab.com/catamphetamine/react-phone-number-input
 */
export const PaySelect = forwardRef<HTMLInputElement, PayTextFieldProps>(
  (
    {
      className,
      label,
      errorMessage,
      prefixIcon,
      linkedAccount,
      linkedAccountOptions,
      onChangeLinkedAccount,
      showConnectAccount,
      ...inputProps
    },
    ref
  ) => {
    const inputRef = useRef<HTMLInputElement>(null)
    useImperativeHandle(ref, () => inputRef.current!, [])

    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={inputProps.id}
          className='ml-2 block text-sm font-medium text-medium'
        >
          {label}
        </label>
        <Listbox
          by={(a, z) => a.id == z.id}
          value={linkedAccount}
          onChange={onChangeLinkedAccount}
        >
          <div className='relative'>
            <div className='mt-1 h-14 w-full rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0'>
              <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
                {prefixIcon && (
                  <div className='-mr-4 flex h-full items-center px-4'>
                    {prefixIcon}
                  </div>
                )}
                <input
                  ref={inputRef}
                  {...inputProps}
                  data-testid='pay-amount-input'
                  className='w-full overflow-hidden border-none bg-transparent px-4 text-2xl focus:ring-0'
                />
                {linkedAccountOptions.length > 0 && (
                  <Listbox.Button
                    className='flex h-full items-center gap-x-2 bg-nav px-4 text-medium focus-visible:bg-nav-active focus-visible:outline-none'
                    data-testid='pay-currency-select'
                  >
                    <Icon>{linkedAccount?.icon}</Icon>
                    {linkedAccount?.mask && <span>{linkedAccount?.mask}</span>}
                  </Listbox.Button>
                )}
              </div>
            </div>
            <Transition
              as={Fragment}
              leave='transition ease-in duration-100'
              leaveFrom='opacity-100'
              leaveTo='opacity-0'
              afterLeave={() => {
                // Ensure the input is put in focus
                inputRef.current!.focus()
              }}
            >
              <Listbox.Options className='absolute z-10 mt-2 max-h-60 w-full overflow-auto rounded-xl bg-nav p-1 shadow-lg focus:outline-none'>
                {linkedAccountOptions.length > 0 &&
                  linkedAccountOptions.map((option, index) => (
                    <Listbox.Option
                      key={index}
                      disabled={!option.enabled}
                      data-testid='pay-currency-option'
                      data-currency-code={option.sendCurrencyCode}
                      data-currency-mask={option.mask}
                      className={({ active, disabled }) =>
                        clsx(
                          'relative flex h-12 select-none items-center gap-x-2 rounded-lg pl-4 pr-3',
                          active ? 'bg-nav-hover' : 'text-medium',
                          disabled
                            ? 'cursor-not-allowed text-disabled'
                            : 'cursor-pointer text-medium'
                        )
                      }
                      value={option}
                    >
                      {({ selected, disabled }) => (
                        <>
                          {option.type != 'wallet' && (
                            <Icon>{option.icon}</Icon>
                          )}
                          {option.type == 'wallet' && (
                            <div
                              className={clsx(
                                disabled && 'grayscale',
                                `flag:${option.receiveCurrencyCountryCode}`
                              )}
                            />
                          )}
                          <span className='block truncate'>
                            {option.title} {option.type != 'wallet' && '****'}{' '}
                            {option.mask}
                          </span>
                          {selected && (
                            <span className='ml-auto flex text-primary'>
                              <Icon>check</Icon>
                            </span>
                          )}
                        </>
                      )}
                    </Listbox.Option>
                  ))}
                {showConnectAccount && (
                  <SelectRouter to={href('/accounts')}>
                    <span>Connect new account</span> <Icon>add</Icon>
                  </SelectRouter>
                )}
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
)

PaySelect.displayName = 'PayTextField'
