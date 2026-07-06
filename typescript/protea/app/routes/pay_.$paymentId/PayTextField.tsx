import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import type { InputHTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'
import { Router } from '~/components/Router'

interface PayTextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // Appended text to the label (Footer notes)
  labelSuffix?: string
  // Text right aligned to the label.
  labelLink?: string
  // Where the labelLink should go.
  labelLinkTo?: string
  // The message from errors produced by form validation.
  errorMessage?: string
  // Any success messages produced by form validation.
  successMessage?: string

  // Prefix text to show in front of the input.
  prefix?: string
  prefixIcon?: ReactNode
  appendIcon?: ReactNode
}

export const PayTextField = forwardRef<HTMLInputElement, PayTextFieldProps>(
  (
    {
      className,
      label,
      labelSuffix,
      labelLink,
      labelLinkTo,
      errorMessage,
      successMessage,
      prefix,
      prefixIcon,
      appendIcon,
      readOnly,
      ...inputProps
    },
    ref
  ) => {
    return (
      <div className={className || 'min-w-full'}>
        {label && (
          <div className='mb-1 flex justify-between'>
            <label
              htmlFor={inputProps.id}
              className='ml-2 block text-sm font-medium text-medium'
            >
              {label} {labelSuffix && <sup>{labelSuffix}</sup>}
            </label>
            {labelLink && labelLinkTo && (
              <Router
                to={labelLinkTo}
                className='mr-2 text-sm font-medium text-primary'
              >
                {labelLink}
              </Router>
            )}
          </div>
        )}
        <div
          className={clsx(
            'block h-14 w-full rounded-xl border-2 border-base',
            readOnly
              ? 'bg-app'
              : 'focus-within:border-focus focus-within:ring-0'
          )}
        >
          <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
            {prefixIcon && (
              <div className='-mr-4 flex h-full items-center px-4'>
                {prefixIcon}
              </div>
            )}
            <input
              ref={ref}
              readOnly={readOnly}
              {...inputProps}
              className='z-0 h-full w-full overflow-hidden border-none bg-transparent px-4 text-2xl focus:ring-0'
            />
          </div>
        </div>
        <AnimatePresence>
          {(errorMessage || successMessage) && (
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
              {successMessage && !errorMessage && (
                <p className='text-sm text-success'>{successMessage}</p>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    )
  }
)

PayTextField.displayName = 'PayTextField'
