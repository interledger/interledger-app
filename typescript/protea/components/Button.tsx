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
      className={`focus:outline-none flex h-10 items-center rounded-full font-display text-sm font-medium text-medium ${
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
