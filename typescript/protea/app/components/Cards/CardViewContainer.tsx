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
        'relative h-card-height-small w-card-width-small overflow-hidden rounded-xl bg-gradient-to-r from-mint-light to-mint-dark font-sans font-titillium text-white shadow-lg xs:h-card-height xs:w-card-width',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
