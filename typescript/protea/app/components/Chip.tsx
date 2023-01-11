import type { FC, ReactNode } from 'react'
import clsx from 'clsx'

export enum ChipColor {
  green = 'bg-green-200 text-green-800',
  purple = 'bg-purple-200 text-purple-800',
  orange = 'bg-orange-200 text-orange-800',
  yellow = 'bg-yellow-200 text-yellow-800',
  blue = 'bg-blue-200 text-blue-800'
}

export type ChipProps = {
  children?: ReactNode
  color: ChipColor
}

export const Chip: FC<ChipProps> = ({ children, color }) => {
  return (
    <div
      className={clsx(
        'flex items-center justify-center rounded-lg px-3 py-1.5 text-xs',
        color
      )}
    >
      {children}
    </div>
  )
}
