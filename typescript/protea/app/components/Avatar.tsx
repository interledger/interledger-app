import clsx from 'clsx'
import type { FC, ReactNode } from 'react'

const colors: {
  [K in AvatarColor]: string
} = {
  green: 'bg-green-100',
  purple: 'bg-purple-100',
  orange: 'bg-orange-100',
  yellow: 'bg-yellow-100',
  sky: 'bg-sky-100',
  rose: 'bg-rose-100',
  indigo: 'bg-indigo-100'
}

type AvatarColor =
  | 'green'
  | 'purple'
  | 'orange'
  | 'yellow'
  | 'sky'
  | 'rose'
  | 'indigo'

const listColors: AvatarColor[] = [
  'green',
  'purple',
  'orange',
  'yellow',
  'sky',
  'rose',
  'indigo'
]

export type AvatarProps = {
  children?: ReactNode
  color?: AvatarColor
  index?: number
}

export const Avatar: FC<AvatarProps> = ({
  children,
  color = 'rose',
  index
}) => {
  let indexColor = color
  if (typeof index != 'undefined')
    indexColor = listColors[index % listColors.length]
  return (
    <div
      className={clsx(
        'flex aspect-square h-12 w-12 items-center justify-center rounded-full font-medium capitalize text-medium',
        colors[indexColor]
      )}
    >
      {children}
    </div>
  )
}
