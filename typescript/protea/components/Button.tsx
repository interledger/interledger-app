import React, { FC } from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {}

export const Button: FC<ButtonProps> = ({ children, ...buttonProps }, ref) => {
  return (
    <button
      {...buttonProps}
      className={`max-w-max h-12 px-6 ${
        buttonProps.disabled
          ? 'bg-strong text-medium cursor-not-allowed focus:outline-none'
          : 'bg-primary cursor-pointer focus:outline-none focus:ring-2 focus:ring-black'
      }`}
    >
      {children}
    </button>
  )
}
