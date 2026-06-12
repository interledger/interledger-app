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
  shrink?: boolean // sm:max-w-fit
}

export const Button = forwardRef<any, ButtonProps>(
  ({ children, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        className={clsx(
          'flex h-12 w-full items-center justify-center rounded-full border border-transparent bg-primary px-10 font-display font-medium text-white focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500 active:ring-blue-400 hover:enabled:bg-blue-400 disabled:cursor-not-allowed disabled:bg-disabled disabled:text-disabled',
          buttonProps.className
        )}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

type TextButtonProps = ButtonHTMLAttributes<HTMLButtonElement>

export const TextButton = forwardRef<any, TextButtonProps>(
  ({ children, ...buttonProps }, ref) => {
    return (
      <button
        ref={ref}
        {...buttonProps}
        className={clsx(
          'text-sm font-medium text-primary focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500 active:ring-blue-400 disabled:cursor-not-allowed disabled:text-disabled',
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
        className={clsx(
          '-m-3 flex p-3 text-medium focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-blue-500 active:ring-blue-400 disabled:cursor-not-allowed disabled:text-disabled',
          buttonProps.className
        )}
      >
        <Icon>{children}</Icon>
      </button>
    )
  }
)

IconButton.displayName = 'IconButton'
