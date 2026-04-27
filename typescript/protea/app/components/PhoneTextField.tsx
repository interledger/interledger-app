import { Listbox, Transition } from '@headlessui/react'
import { AnimatePresence, motion } from 'framer-motion'
import type { CountryCode } from 'libphonenumber-js'
import { AsYouType, getCountryCallingCode } from 'libphonenumber-js'
import type { ChangeEventHandler, InputHTMLAttributes } from 'react'
import {
  Fragment,
  forwardRef,
  useCallback,
  useImperativeHandle,
  useRef,
  useState
} from 'react'
import { Icon } from '~/components/Icon'

interface PhoneFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string

  defaultCountry: string
  options: PhoneAutocompleteOptions[]
}

export type PhoneAutocompleteOptions = {
  id: CountryCode
  name: string
}

/**
 * Phone text field component with country select.
 * Includes two input fields linked to the form specified in the props.
 *
 * References:
 * https://headlessui.dev/react/Listbox
 * https://gitlab.com/catamphetamine/react-phone-number-input
 */
export const PhoneTextField = forwardRef<
  HTMLInputElement | undefined,
  PhoneFieldProps
>(
  (
    { className, label, errorMessage, defaultCountry, options, ...inputProps },
    ref
  ) => {
    const inputRef = useRef<HTMLInputElement>(null)
    useImperativeHandle(ref, () => inputRef.current!, [])

    const [country, setCountry] = useState<PhoneAutocompleteOptions>(
      options.find(
        (country: PhoneAutocompleteOptions) => country.id == defaultCountry
      ) as PhoneAutocompleteOptions
    )

    const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
      (event) => {
        const formatter = new AsYouType(country.id)
        formatter.input(event.target.value)

        if (
          typeof formatter.country !== 'undefined' &&
          country.id != formatter.country
        ) {
          setCountry(
            options.find(
              (country: PhoneAutocompleteOptions) =>
                country.id == formatter.country
            ) as PhoneAutocompleteOptions
          )
        }
      },
      [country.id, options]
    )

    const _onChangeCountry = useCallback(
      (newCountry: PhoneAutocompleteOptions) => {
        const currentNumber = inputRef.current!.value

        if (currentNumber.startsWith('+')) {
          inputRef.current!.value = currentNumber.replace(
            `+${getCountryCallingCode(country.id)}`,
            `+${getCountryCallingCode(newCountry.id)}`
          )
        } else {
          inputRef.current!.value = `+${getCountryCallingCode(
            newCountry.id
          )}${currentNumber}`
        }

        setCountry(newCountry)
      },
      [country]
    )

    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={inputProps.id}
          className='ml-2 block text-sm font-medium text-medium'
        >
          {label}
        </label>
        <input
          form={inputProps.form}
          value={String(country?.id)}
          name='country'
          type='hidden'
        />
        <Listbox value={country} onChange={_onChangeCountry}>
          <div className='relative'>
            <div className='mt-1 h-12 w-full rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0'>
              <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
                <Listbox.Button className='flex h-full items-center bg-nav px-4 text-medium focus-visible:bg-nav-active focus-visible:outline-none'>
                  <div className={`flag:${country?.id}`} />
                </Listbox.Button>
                <input
                  ref={inputRef}
                  {...inputProps}
                  defaultValue={
                    (inputProps.defaultValue as string) ||
                    `+${getCountryCallingCode(defaultCountry as CountryCode)}`
                  }
                  type='tel'
                  onChange={_onChangeInput}
                  className='w-full overflow-hidden border-none bg-transparent px-4 focus:ring-0'
                />
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
                // Ensure that the cursor is sent to the end of the input
                inputRef.current!.selectionStart =
                  inputRef.current!.value.length
                inputRef.current!.selectionEnd = inputRef.current!.value.length
              }}
            >
              <Listbox.Options className='absolute z-10 mt-2 max-h-60 w-full overflow-auto rounded-xl bg-nav p-1 text-sm shadow-lg focus:outline-none'>
                {options.length > 0 &&
                  options.map((option, index) => (
                    <Listbox.Option
                      key={index}
                      className={({ active }) =>
                        `relative flex h-12 cursor-pointer select-none items-center justify-between rounded-lg pl-4 pr-3 ${active ? 'bg-nav-hover' : 'text-medium'
                        }`
                      }
                      value={option}
                    >
                      {/*<div className={`flag:${option.id}`} />*/}
                      {({ selected }) => (
                        <>
                          <span className={`block truncate`}>
                            {option.name}
                          </span>
                          {selected && (
                            <span className='flex text-primary'>
                              <Icon>check</Icon>
                            </span>
                          )}
                        </>
                      )}
                    </Listbox.Option>
                  ))}
                {options.length === 0 && (
                  <div className='relative flex h-12 select-none items-center justify-between pl-4 pr-3 text-medium'>
                    Nothing found.
                  </div>
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

PhoneTextField.displayName = 'PhoneTextField'
