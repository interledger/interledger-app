import React, { FC } from 'react'

interface TextFieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // The **isValid** value of `formState` from `useForm`. Sets the input into error state.
  isValid?: boolean
  // The message from errors produced by form validation.
  errorMessage?: string
}

export const TextField = React.forwardRef<any, TextFieldProps>(
  ({ className, label, isValid, errorMessage, ...inputProps }, ref) => {
    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={inputProps.id}
          className='block text-sm font-medium text-medium ml-2'
        >
          {label}
        </label>
        <input
          ref={ref}
          {...inputProps}
          className='focus:ring-primary focus:border-primary h-12 block w-full border-base border-2 mt-1'
        />
        <div className='h-7 pt-2 pl-2'>
          {!isValid && <p className='text-sm text-error'>{errorMessage}</p>}
        </div>
      </div>
    )
  }
)

TextField.displayName = 'TextField'
