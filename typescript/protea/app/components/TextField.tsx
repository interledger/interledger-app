import { forwardRef } from 'react'

interface TextFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string
}

export const TextField = forwardRef<any, TextFieldProps>(
  ({ className, label, errorMessage, ...inputProps }, ref) => {
    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={inputProps.id}
          className='ml-2 block text-sm font-medium text-medium'
        >
          {label}
        </label>
        <input
          ref={ref}
          {...inputProps}
          className='mt-1 block h-12 w-full rounded-xl border-2 border-base focus:border-focus focus:ring-0'
        />
        <div className='h-7 pt-2 pl-2'>
          {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
        </div>
      </div>
    )
  }
)

TextField.displayName = 'TextField'
