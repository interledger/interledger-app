import { forwardRef } from 'react'

interface TextAreaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  // The label value.
  label?: string
  // The message from errors produced by form validation.
  errorMessage?: string
}

export const TextArea = forwardRef<any, TextAreaProps>(
  ({ className, label, errorMessage, ...textAreaProps }, ref) => {
    return (
      <div className={className || 'min-w-full'}>
        <label
          htmlFor={textAreaProps.id}
          className='ml-2 block text-sm font-medium text-medium'
        >
          {label}
        </label>
        <textarea
          ref={ref}
          {...textAreaProps}
          className='mt-1 block h-36 w-full resize-y rounded-xl border-2 border-base focus:border-focus focus:ring-0'
        />
        <div className='h-7 pt-2 pl-2'>
          {errorMessage && <p className='text-sm text-error'>{errorMessage}</p>}
        </div>
      </div>
    )
  }
)

TextArea.displayName = 'TextArea'
