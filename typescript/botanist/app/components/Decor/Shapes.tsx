import type { FC } from 'react'
import clsx from 'clsx'
import { Icon } from '~/components'
import type { ReactNode } from 'react'

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

export type ShapeProps = {
  width?: string
  radius: Radius
  color: string
  children?: ReactNode
}

export const Shape: FC<ShapeProps> = ({ children, width, radius, color }) => {
  return (
    <div
      className={clsx(
        'flex aspect-square items-center justify-center',
        width || 'w-full',
        radius,
        color
      )}
    >
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

export function SuccessShapes() {
  const shapes: ShapeProps[][] = [
    [
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-14',
        radius: 'rounded-tl-full',
        color: 'bg-slate-100'
      },
      {
        width: 'w-14',
        radius: 'rounded-br-full',
        color: 'bg-green-50'
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-transparent'
      }
    ],
    [
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-slate-50'
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-green-200'
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-green-500'
      },
      {
        width: 'w-14',
        radius: 'rounded-tl-full',
        color: 'bg-green-100'
      },
      {
        width: 'w-14',
        radius: 'rounded-tr-full',
        color: 'bg-slate-50'
      }
    ],
    [
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-transparent'
      },
      {
        width: 'w-14',
        radius: 'rounded-l-full',
        color: 'bg-green-50'
      },
      {
        width: 'w-14',
        radius: 'rounded-none',
        color: 'bg-green-50'
      },
      {
        width: 'w-14',
        radius: 'rounded-br-full',
        color: 'bg-green-200'
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-transparent'
      }
    ]
  ]
  return (
    <div>
      {shapes.map((shapeRow, outerIndex, outerArray) => (
        <div className='flex justify-center' key={`shapeRow${outerIndex}`}>
          {shapeRow.map((shape, index, array) => (
            <Shape key={`shape${outerIndex}${index}`} {...shape}>
              {Math.floor(outerArray.length / 2) == outerIndex &&
                Math.floor(array.length / 2) == index && (
                  <Icon className='text-white'>check</Icon>
                )}
            </Shape>
          ))}
        </div>
      ))}
    </div>
  )
}

export const BlogShapes: FC<{ slug: string; large?: boolean }> = ({
  slug,
  large
}) => {
  if (typeof blogShapes[slug] == 'undefined')
    throw new Error(`Shapes for ${slug} blog post not found`)
  const shapes: ShapeProps[][] = blogShapes[slug]
  return (
    <div>
      {shapes.map((shapeRow, outerIndex) => (
        <div className='flex' key={`shapeRow${outerIndex}`}>
          {shapeRow.map((shape, index) => (
            <Shape
              key={`shape${outerIndex}${index}`}
              width={clsx(large ? 'w-20' : 'w-10')}
              {...shape}
            />
          ))}
        </div>
      ))}
    </div>
  )
}

const blogShapes: {
  [key: string]: ShapeProps[][]
} = {
  'connecting-the-internet-economy': [
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-rose-400'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-600'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-200'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-yellow-100'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-orange-900'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-orange-600'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-yellow-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-400'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-yellow-600'
      }
    ]
  ],
  'card-payments': [
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-slate-100'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-100'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-blue-500'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-yellow-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-400'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-orange-600'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-300'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-sky-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-blue-500'
      }
    ]
  ]
}
