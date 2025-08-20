import clsx from 'clsx'
import type { ComponentProps } from 'react'

export type CardViewContainerProps = ComponentProps<'div'>

export const CardViewContainer = ({
  children,
  className,
  ...props
}: CardViewContainerProps) => {
  return (
    <div
      className={clsx(
        'relative h-52 w-80 overflow-hidden rounded-xl bg-gradient-to-br from-slate-800 to-slate-900 px-5 py-4 font-sans text-white shadow-lg',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
