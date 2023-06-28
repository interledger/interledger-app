import { NavLink } from '@remix-run/react'
import type { RemixNavLinkProps } from '@remix-run/react/dist/components'
import clsx from 'clsx'
import type { ButtonHTMLAttributes, HTMLAttributes, RefAttributes } from 'react'
import { forwardRef } from 'react'

const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          'flex w-full flex-col rounded-[1.25rem] bg-container-strong',
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
      className={clsx('flex w-full justify-between p-4 pb-0', className)}
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
    <div ref={ref} className={clsx('p-4', className)} {...props} />
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
          'mx-2 my-1 flex rounded-xl p-3 first-of-type:mt-2 last-of-type:mb-2 focus-visible:outline-2 focus-visible:outline-focus',
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

const CardButton = forwardRef<
  HTMLButtonElement,
  ButtonHTMLAttributes<HTMLButtonElement>
>(({ className, ...props }, ref) => {
  return (
    <button
      ref={ref}
      className={clsx(
        'mx-2 my-1 flex rounded-xl p-3 first-of-type:mt-2 last-of-type:mb-2 hover:bg-nav focus-visible:outline-2 focus-visible:outline-focus active:bg-nav-hover',
        className
      )}
      {...props}
    />
  )
})
CardButton.displayName = 'CardButton'

export {
  Card,
  CardIcon,
  CardContent,
  CardHeader,
  CardTitle,
  CardLink,
  CardButton
}
