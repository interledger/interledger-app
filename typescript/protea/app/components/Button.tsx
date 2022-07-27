import type { ReactNode, ButtonHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import { Icon } from '.'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: string
  outline?: boolean
}

export const Button = forwardRef<any, ButtonProps>(
  ({ children, icon, outline, ...buttonProps }, ref) => {
    const activeClassNames = outline
      ? 'bg-white ring-2 ring-base hover:ring-hover active:ring-active'
      : 'bg-container-primary hover:bg-container-primary-hover active:bg-container-primary-active'
    return (
      <button
        ref={ref}
        {...buttonProps}
        className={`flex h-10 items-center rounded-full font-display text-sm font-medium text-medium ${
          icon ? 'pl-4 pr-6' : 'px-6'
        } ${
          buttonProps.disabled
            ? 'cursor-not-allowed bg-disabled text-disabled'
            : `cursor-pointer focus-visible:outline-2 focus-visible:outline-focus ${activeClassNames}`
        }`}
      >
        {icon && (
          <div className='mr-2'>
            <Icon>{icon}</Icon>
          </div>
        )}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

interface FABProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: ReactNode
  hasNav?: boolean
}

export const FAB = forwardRef<any, FABProps>(
  ({ children, icon, hasNav, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        className={`fixed right-4 flex h-14 w-min items-center space-x-3 rounded-2xl p-4 font-display text-sm font-medium text-medium lg:hidden ${
          hasNav ? 'bottom-24 sm:bottom-4' : 'bottom-4'
        } ${children ? 'pr-5' : ''} ${
          buttonProps.disabled
            ? 'cursor-not-allowed bg-disabled text-disabled'
            : `cursor-pointer bg-container-primary shadow-lg hover:bg-container-primary-hover focus-visible:outline-2 focus-visible:outline-focus active:bg-container-primary-active`
        }`}
      >
        <Icon>{icon}</Icon>
        {children && <div>{children}</div>}
      </button>
    )
  }
)

FAB.displayName = 'Floating action button'
