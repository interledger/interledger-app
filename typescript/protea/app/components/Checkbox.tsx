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
      <div className={className || 'min-w-full'}>
        <div className='flex h-6 w-6 items-center justify-center'>
          <input
            ref={ref}
            {...inputProps}
            type='checkbox'
            className='h-[1.125rem] w-[1.125rem] cursor-pointer rounded-sm border-2 border-base bg-container-strong text-transparent focus:ring-offset-container-strong focus-visible:ring-focus'
          />
        </div>
        <div className='ml-2 text-sm'>
          <label htmlFor={inputProps.id} className='cursor-pointer text-xs'>
            {children}
          </label>
          <div className='-ml-7 h-7 pt-2'>
            {errorMessage && (
              <p className='text-xs text-error'>{errorMessage}</p>
            )}
          </div>
        </div>
      </div>
    )
  }
)

Checkbox.displayName = 'Checkbox'
