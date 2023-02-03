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
        className='grid w-full grid-cols-4 content-start gap-4 gap-y-6 px-4 sm:grid-cols-8 sm:px-0 lg:grid-cols-12'
      >
        {children}
      </div>
    )
  }
)

WalletGrid.displayName = 'WalletGrid'
