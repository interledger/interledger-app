import clsx from 'clsx'
import type { ButtonHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import { Icon } from '.'

/**
 * TODO: Button refactor:
 * - Button - this is the default solid blue button.
 * - TextButton - a button styled like a TextRouter.
 * - OutlineButton - the outline style button.
 * - IconButton - a button that houses an Icon - should just take text of the icon as children.
 */

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  error?: boolean
  shrink?: boolean // sm:max-w-fit
}

export const Button = forwardRef<any, ButtonProps>(
  ({ children, error, shrink, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        disabled={buttonProps.disabled}
        className={clsx(
          'flex h-12 w-full items-center justify-center rounded-full border border-transparent px-6 font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2  disabled:cursor-not-allowed disabled:bg-disabled disabled:text-disabled',
          error
            ? 'bg-container-strong text-error focus-visible:outline-error active:ring-error hover:enabled:bg-error-hover'
            : 'bg-primary text-white focus-visible:outline-focus active:ring-blue-400 hover:enabled:bg-blue-400',
          shrink && 'sm:max-w-fit',
          buttonProps.className
        )}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

export const OutlineButton = forwardRef<any, ButtonProps>(
  ({ children, shrink, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        disabled={buttonProps.disabled}
        className={clsx(
          'flex h-12 w-full items-center justify-center rounded-full border border-transparent px-6 font-medium text-primary outline outline-2 -outline-offset-2 outline-blue-500 hover:text-primary-hover hover:outline-hover focus-visible:bg-container-primary disabled:cursor-not-allowed disabled:bg-disabled disabled:text-disabled',
          shrink && 'sm:max-w-fit',
          buttonProps.className
        )}
      >
        {children}
      </button>
    )
  }
)

OutlineButton.displayName = 'OutlineButton'

type TextButtonProps = ButtonHTMLAttributes<HTMLButtonElement>

export const TextButton = forwardRef<any, TextButtonProps>(
  ({ children, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        disabled={buttonProps.disabled}
        className={clsx(
          'rounded text-sm font-medium text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500 active:ring-blue-400 disabled:cursor-not-allowed disabled:text-disabled',
          buttonProps.className
        )}
      >
        {children}
      </button>
    )
  }
)

TextButton.displayName = 'TextButton'

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  shrink?: boolean // sm:max-w-fit
}

export const IconButton = forwardRef<any, IconButtonProps>(
  ({ children, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        disabled={buttonProps.disabled}
        className='-m-3 flex rounded-lg p-3 text-medium focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-4 focus-visible:outline-blue-500 active:ring-blue-400 disabled:cursor-not-allowed disabled:text-disabled'
      >
        <Icon className={buttonProps.className}>{children}</Icon>
      </button>
    )
  }
)

IconButton.displayName = 'IconButton'
