import { AnimatePresence, motion } from 'framer-motion'
import { forwardRef } from 'react'

interface TextAreaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
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
          className='mt-1 block h-36 w-full resize-y rounded-xl border-2 border-base bg-transparent focus:border-focus focus:ring-0'
        />
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

TextArea.displayName = 'TextArea'
