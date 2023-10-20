import { NavLink } from '@remix-run/react'
import type { RemixNavLinkProps } from '@remix-run/react/dist/components'
import clsx from 'clsx'
import type { ButtonHTMLAttributes, HTMLAttributes, RefAttributes } from 'react'
import { forwardRef } from 'react'

const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        data-testid='card'
        ref={ref}
        className={clsx(
          'flex w-full flex-col rounded-[1.25rem] bg-container-strong p-2',
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
      className={clsx('flex w-full justify-between px-2 pb-0 pt-2', className)}
      {...props}
    />
  )
)
CardHeader.displayName = 'CardHeader'

const CardTitle = forwardRef<
  HTMLHeadingElement,
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

const CardIcon = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={clsx(
        'flex items-center justify-between rounded-full bg-nav p-5 text-medium',
        className
      )}
      {...props}
    />
  )
)
CardIcon.displayName = 'CardIcon'

const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={clsx('p-2', className)} {...props} />
  )
)
CardContent.displayName = 'CardContent'

const CardLink = forwardRef<
  any,
  RemixNavLinkProps & RefAttributes<HTMLAnchorElement>
>(({ children, className, ...props }, ref) => {
  return (
    <NavLink
      ref={ref}
      className={({ isActive }) =>
        clsx(
          'my-1 flex rounded-xl p-3 first:mt-0 last-of-type:mb-0 focus-visible:outline-2 focus-visible:outline-focus',
          isActive ? 'bg-nav-hover' : 'hover:bg-nav',
          className
        )
      }
      {...props}
    >
      {children}
    </NavLink>
  )
})
CardLink.displayName = 'CardLink'

interface CardButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  noHover?: boolean
}

const CardButton = forwardRef<HTMLButtonElement, CardButtonProps>(
  ({ className, noHover, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={clsx(
          noHover && 'bg-nav',
          'my-1 flex rounded-xl p-3 first:mt-0 last-of-type:mb-0 hover:bg-nav focus-visible:outline-2 focus-visible:outline-focus active:bg-nav-hover',
          className
        )}
        {...props}
      />
    )
  }
)
CardButton.displayName = 'CardButton'

export {
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle
}
