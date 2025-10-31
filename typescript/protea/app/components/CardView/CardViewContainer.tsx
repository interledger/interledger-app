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
        'relative h-52 w-[21rem] overflow-hidden rounded-xl bg-gradient-to-r from-mint-light to-mint-dark font-sans text-white shadow-lg',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
