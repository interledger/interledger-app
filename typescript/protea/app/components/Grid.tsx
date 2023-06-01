import type { HTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'

interface WalletGridProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode
}

export const WalletGrid = forwardRef<any, WalletGridProps>(
  ({ children }, ref) => {
    return (
      <div
        ref={ref}
        className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-6 sm:max-w-lg sm:grid-cols-8 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'
      >
        {children}
      </div>
    )
  }
)

WalletGrid.displayName = 'WalletGrid'
