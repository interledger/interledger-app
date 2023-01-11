import type { FC, ReactNode } from 'react'

type WalletGridProps = {
  children?: ReactNode
}

export const WalletGrid: FC<WalletGridProps> = ({ children }) => {
  return (
    <div className='grid w-full grid-cols-4 content-start gap-4 gap-y-6 px-4 sm:grid-cols-8 sm:px-0 lg:grid-cols-12'>
      {children}
    </div>
  )
}

WalletGrid.displayName = 'WalletGrid'
