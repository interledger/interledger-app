import clsx from 'clsx'
import type { HTMLAttributes } from 'react'
import { forwardRef } from 'react'

interface AlertProps extends HTMLAttributes<HTMLDivElement> {
  success?: boolean
}

const Alert = forwardRef<HTMLDivElement, AlertProps>(
  ({ success = false, className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'flex w-full gap-x-2 rounded-lg p-4',
          className,
          success ? 'bg-alert-success' : 'bg-alert-slate'
        )}
        {...props}
      />
    )
  }
)
Alert.displayName = 'Alert'

const AlertTitle = forwardRef<
  HTMLParagraphElement,
  HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => {
  return (
    // eslint-disable-next-line jsx-a11y/heading-has-content -- heading content is supplied by the consumer via {...props}
    <h2
      ref={ref}
      className={clsx('font-medium text-medium', className)}
      {...props}
    />
  )
})
AlertTitle.displayName = 'AlertTitle'

const AlertBody = forwardRef<
  HTMLParagraphElement,
  HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => {
  return <p ref={ref} className={clsx('text-medium', className)} {...props} />
})
AlertBody.displayName = 'AlertBody'

const AlertContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={clsx('flex w-full flex-col gap-y-2', className)}
      {...props}
    />
  )
)
AlertContent.displayName = 'AlertContent'

export { Alert, AlertBody, AlertContent, AlertTitle }
