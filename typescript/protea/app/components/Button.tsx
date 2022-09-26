import clsx from 'clsx'
import type { ReactNode, ButtonHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import { Icon } from '.'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: string
  outline?: boolean
}

export const Button = forwardRef<any, ButtonProps>(
  ({ children, icon, outline, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        className={clsx(
          'flex h-12 w-full items-center justify-center rounded-full font-display font-medium focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500 disabled:cursor-not-allowed disabled:bg-disabled disabled:text-disabled',
          icon ? 'pl-4 pr-6' : 'px-6',
          outline
            ? 'border-2 border-blue-500 bg-app text-primary active:ring-active hover:enabled:border-hover'
            : 'border border-transparent bg-primary text-white active:ring-blue-400 hover:enabled:bg-blue-400',
          buttonProps.className
        )}
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
            : `bg-container-primary shadow-lg hover:bg-container-primary-hover focus-visible:outline-2 focus-visible:outline-focus active:bg-container-primary-active`
        }`}
      >
        <Icon>{icon}</Icon>
        {children && <div>{children}</div>}
      </button>
    )
  }
)

FAB.displayName = 'Floating action button'
