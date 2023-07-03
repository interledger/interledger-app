import clsx from 'clsx'
import type { HTMLAttributes } from 'react'
import { forwardRef } from 'react'

const WalletGrid = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ children }, ref) => {
    return (
      <div
        ref={ref}
        className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 sm:max-w-lg sm:grid-cols-8 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'
      >
        {children}
      </div>
    )
  }
)

WalletGrid.displayName = 'WalletGrid'

interface GridColumnProps extends HTMLAttributes<HTMLDivElement> {
  sticky?: boolean
  hideOnMobile?: boolean
}
const GridColumn = forwardRef<HTMLDivElement, GridColumnProps>(
  ({ sticky, hideOnMobile, className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={clsx(
          sticky && 'sticky top-16',
          hideOnMobile ? 'hidden lg:flex' : 'flex',
          'h-max max-h-max flex-col gap-y-4',
          className
        )}
        {...props}
      />
    )
  }
)

GridColumn.displayName = 'GridColumn'

export { WalletGrid, GridColumn }
