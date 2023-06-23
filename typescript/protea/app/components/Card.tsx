import { NavLink } from '@remix-run/react'
import type { RemixNavLinkProps } from '@remix-run/react/dist/components'
import clsx from 'clsx'
import type { HTMLAttributes, ReactNode, RefAttributes } from 'react'
import { forwardRef } from 'react'

const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'flex w-full flex-col rounded-[1.25rem] bg-container-strong p-4',
          className
        )}
        {...props}
      />
    )
  }
)
Card.displayName = 'Card'

const CardHeader = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={clsx('flex w-full justify-between', className)}
      {...props}
    />
  )
)
CardHeader.displayName = 'CardHeader'

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
        className={clsx(
          'mt-4 flex w-full justify-between first-of-type:mt-0',
          className
        )}
        {...props}
      />
    )
  }
)
CardRow.displayName = 'CardRow'

const CardColumn = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'mt-4 flex w-full flex-col gap-y-2 first-of-type:mt-0',
          className
        )}
        {...props}
      />
    )
  }
)
CardColumn.displayName = 'CardColumn'

const CardLink = forwardRef<
  any,
  RemixNavLinkProps & RefAttributes<HTMLAnchorElement>
>(({ children, className, ...props }, ref) => {
  return (
    <NavLink
      ref={ref}
      className={clsx(
        'group relative mt-4 flex w-full rounded-xl py-1 first-of-type:mt-6 focus-visible:outline-2 focus-visible:outline-focus',
        className
      )}
      {...props}
    >
      {({ isActive }) => (
        <>
          <div
            className={clsx(
              'absolute -inset-2 flex rounded-xl',
              isActive ? 'bg-nav-hover' : 'group-hover:bg-nav'
            )}
          />
          <div className={clsx('z-10 flex w-full', className)}>
            {children as ReactNode}
          </div>
        </>
      )}
    </NavLink>
  )
})
CardLink.displayName = 'CardLink'

export { Card, CardRow, CardColumn, CardTitle, CardLink }
