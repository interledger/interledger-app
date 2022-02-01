import React, { FC } from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {}

export const Button: FC<ButtonProps> = ({ children, ...buttonProps }, ref) => {
  return (
    <button
      {...buttonProps}
      className={`h-12 max-w-max px-6 ${
        buttonProps.disabled
          ? 'focus:outline-none cursor-not-allowed bg-strong text-medium'
          : 'focus:outline-none cursor-pointer bg-primary focus:ring-2 focus:ring-focus'
      }`}
    >
      {children}
    </button>
  )
}
