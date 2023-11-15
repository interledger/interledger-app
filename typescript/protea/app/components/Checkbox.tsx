import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import type { InputHTMLAttributes } from 'react'
import { forwardRef } from 'react'

interface CheckboxProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The message from errors produced by form validation.
  errorMessage?: string
}

export const Checkbox = forwardRef<any, CheckboxProps>(
  ({ className, errorMessage, children, ...inputProps }, ref) => {
    return (
      <div className={clsx(className || 'min-w-full')}>
        <div className='flex h-6 w-6 items-center justify-center'>
          <input
            ref={ref}
            {...inputProps}
            type='checkbox'
            className='h-[1.125rem] w-[1.125rem] cursor-pointer rounded-sm border-2 border-base bg-container-strong text-transparent focus:ring-offset-container-strong focus-visible:ring-focus disabled:cursor-not-allowed'
          />
        </div>
        <div className='ml-2'>
          <label
            htmlFor={inputProps.id}
            className={clsx(
              'text-sm text-medium',
              inputProps.disabled ? 'cursor-not-allowed' : 'cursor-pointer'
            )}
          >
            {children}
          </label>
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
                className='-ml-7 h-7 pt-2'
              >
                {errorMessage && (
                  <p className='text-xs text-error'>{errorMessage}</p>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    )
  }
)

Checkbox.displayName = 'Checkbox'
