import type { FC, HTMLAttributes, ReactNode } from 'react'

interface GridProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode
}

export const Grid: FC<GridProps> = ({ children }) => {
  return (
    <div className='grid w-full grid-cols-4 content-start gap-4 gap-y-6 px-4 sm:grid-cols-8 sm:px-0 lg:grid-cols-12'>
      {children}
    </div>
  )
}

Grid.displayName = 'WalletGrid'
