import type { HTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'
import { Outlet } from '@remix-run/react'

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

interface FocusGridProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode
}

export const FocusGrid = forwardRef<any, FocusGridProps>(
  ({ children }, ref) => {
    return (
      <div
        ref={ref}
        className='col-span-full mt-16 px-4 space-y-6 sm:px-0 lg:col-span-6 lg:col-start-4 lg:mt-36'
      >
        {children}
      </div>
    )
  }
)

FocusGrid.displayName = 'WalletGrid'
