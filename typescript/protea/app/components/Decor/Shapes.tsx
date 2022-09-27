import { FC } from 'react'
import clsx from 'clsx'

type Radius =
  | 'rounded-none'
  | 'rounded-full'
  | 'rounded-t-full'
  | 'rounded-tl-full'
  | 'rounded-tr-full'
  | 'rounded-b-full'
  | 'rounded-bl-full'
  | 'rounded-br-full'
  | 'rounded-l-full'
  | 'rounded-r-full'

type ShapeProps = {
  key?: string
  width?: string
  radius: Radius
  color: string
}

export const Shape: FC<ShapeProps> = ({
  children,
  key,
  width,
  radius,
  color
}) => {
  return (
    <div
      key={key}
      className={clsx('aspect-square', width || 'w-full', radius, color)}
    >
      {children}
    </div>
  )
}
