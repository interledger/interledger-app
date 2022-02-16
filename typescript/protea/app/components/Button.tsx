import React, { FC } from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ReactNode
  outline?: boolean
}

export const Button: FC<ButtonProps> = (
  { children, icon, outline, ...buttonProps },
  ref
) => {
  const activeClassNames = outline
    ? 'bg-white ring-2 ring-base hover:ring-hover active:ring-active'
    : 'bg-container-primary hover:bg-container-primary-hover active:bg-container-primary-active'
  return (
    <button
      {...buttonProps}
      className={`flex h-10 items-center rounded-full font-display text-sm font-medium text-medium focus:outline-none ${
        icon ? 'pl-4 pr-6' : 'px-6'
      } ${
        buttonProps.disabled
          ? 'cursor-not-allowed bg-disabled text-disabled'
          : `cursor-pointer focus:ring-2 focus:ring-focus ${activeClassNames}`
      }`}
    >
      {icon && <div className='mr-2'>{icon}</div>}
      {children}
    </button>
  )
}

interface FABProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ReactNode
  hasNav?: boolean
}

export const FAB: FC<FABProps> = (
  { children, icon, hasNav, ...buttonProps },
  ref
) => {
  return (
    <button
      {...buttonProps}
      className={`fixed right-4 flex h-14 w-min items-center space-x-3 rounded-2xl p-4 font-display text-sm font-medium text-medium focus:outline-none lg:hidden ${
        hasNav ? 'bottom-24 sm:bottom-4' : 'bottom-4'
      } ${children ? 'pr-5' : ''} ${
        buttonProps.disabled
          ? 'cursor-not-allowed bg-disabled text-disabled'
          : `cursor-pointer bg-container-primary shadow-lg hover:bg-container-primary-hover focus:ring-2 focus:ring-focus active:bg-container-primary-active`
      }`}
    >
      <div>{icon}</div>
      {children && <div>{children}</div>}
    </button>
  )
}
