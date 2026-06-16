import type { InputHTMLAttributes } from 'react'
import { forwardRef } from 'react'

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // Appended text to the label (Footer notes)
  labelSuffix?: string
  // The message from errors produced by form validation.
  errorMessage?: string
  // Any success messages produced by form validation.
  successMessage?: string

  // Prefix text to show in front of the input.
  prefix?: string
  prefixIcon?: JSX.Element
  appendIcon?: JSX.Element
}

export const TextField = forwardRef<any, TextFieldProps>(
  (
    {
      className,
      label,
      labelSuffix,
      errorMessage,
      successMessage,
      prefix,
      prefixIcon,
      appendIcon,
      ...inputProps
    },
    ref
  ) => {
    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={inputProps.id}
          className='ml-2 block text-sm font-medium text-medium'
        >
          {label} {labelSuffix && <sup>{labelSuffix}</sup>}
        </label>
        <div className='mt-1 block h-12 w-full rounded-xl border-2 border-base focus-within:border-focus focus-within:ring-0'>
          <div className='flex h-full items-center justify-between overflow-hidden rounded-[10px]'>
            {prefixIcon && (
              <div className='-mr-3 flex h-full items-center px-3'>
                {prefixIcon}
              </div>
            )}
            {prefix && (
              <span className='z-10 -mr-[0.9375rem] ml-4 text-disabled'>
                {prefix}
              </span>
            )}
            <input
              ref={ref}
              {...inputProps}
              className='z-0 h-full w-full overflow-hidden border-none bg-transparent px-4 focus:ring-0'
            />
            {appendIcon && (
              <div className='-ml-3 flex h-full items-center px-3'>
                {appendIcon}
              </div>
            )}
          </div>
        </div>
        <div className='h-7 pl-2 pt-2'>
          {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
          {successMessage && !errorMessage && (
            <p className='text-sm text-success'>{successMessage}</p>
          )}
        </div>
      </div>
    )
  }
)

TextField.displayName = 'TextField'
