import type { FC, Key } from 'react'
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
  width?: string
  radius: Radius
  color: string
}

export const Shape: FC<ShapeProps> = ({ children, width, radius, color }) => {
  return (
    <div className={clsx('aspect-square', width || 'w-full', radius, color)}>
      {children}
    </div>
  )
}

export function HomeShapes() {
  const shapes: ShapeProps[][] = [
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-600'
      },
      {
        radius: 'rounded-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-yellow-400'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-rose-300'
      },
      {
        radius: 'rounded-full',
        color: 'bg-lime-400'
      },
      {
        radius: 'rounded-b-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-500'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-lime-300'
      },
      {
        radius: 'rounded-b-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-b-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-tr-full',
        color: 'bg-lime-500'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-b-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-b-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-tl-full',
        color: 'bg-rose-400'
      }
    ],
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-lime-500'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-300'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-yellow-200'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-500'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-transparent'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-100'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-300'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-bl-full',
        color: 'bg-yellow-200'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-full',
        color: 'bg-lime-400'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-tr-full',
        color: 'bg-yellow-400'
      },
      {
        width: 'w-0 lg:w-full',
        radius: 'rounded-tl-full',
        color: 'bg-transparent'
      }
    ]
  ]
  return (
    <div>
      {shapes.map((shapeRow, outerIndex) => (
        <div className='flex' key={`shapeRow${outerIndex}`}>
          {shapeRow.map((shape, index) => (
            <Shape key={`shape${outerIndex}${index}`} {...shape} />
          ))}
        </div>
      ))}
    </div>
  )
}
