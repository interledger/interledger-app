import clsx from 'clsx'
import type { HTMLAttributes } from 'react'
import { forwardRef } from 'react'

const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'flex w-full flex-col rounded-2xl bg-container-strong p-4',
          className
        )}
        {...props}
      />
    )
  }
)
Card.displayName = 'Card'

const CardTitle = forwardRef<
  HTMLParagraphElement,
  HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => {
  return (
    <h2
      ref={ref}
      className={clsx('text-lg font-medium text-strong', className)}
      {...props}
    />
  )
})
CardTitle.displayName = 'CardTitle'

const CardRow = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx('flex w-full justify-between', className)}
        {...props}
      />
    )
  }
)
CardRow.displayName = 'CardRow'

export { Card, CardRow, CardTitle }
