import clsx from 'clsx'
import type { MotionProps } from 'framer-motion'
import { AnimatePresence, motion } from 'framer-motion'
import type { FC, ReactNode } from 'react'
import { forwardRef, useEffect, useMemo, useState } from 'react'
import { Icon } from '~/components'

export type Radius =
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
  | 'rounded-full rounded-tl-none'
  | 'rounded-full rounded-tr-none'
  | 'rounded-full rounded-bl-none'
  | 'rounded-full rounded-br-none'
  | 'rounded-tr-full rounded-bl-full'
  | 'rounded-tl-full rounded-br-full'

interface MotionShapeProps extends MotionProps {
  children?: ReactNode
  width?: string
  radius?: Radius
  color?: string
  flex?: 'flex-1' | 'flex-auto' | 'flex-initial' | 'flex-none'
}

export const Shape = forwardRef<any, MotionShapeProps>(
  (
    { children, width, radius, color, flex = 'flex-initial', ...motionProps },
    ref
  ) => {
    return (
      <motion.div
        ref={ref}
        {...motionProps}
        className={clsx(
          'flex aspect-square items-center justify-center',
          flex,
          width || 'w-full',
          radius,
          color
        )}
      >
        {children}
      </motion.div>
    )
  }
)

Shape.displayName = 'Shape'

export const MotionShape = motion(Shape, { forwardMotionProps: true })

export function SuccessShapes() {
  const shapes: MotionShapeProps[][] = [
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
        color: 'bg-slate-100',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 1.2, delay: 0 }
      },
      {
        width: 'w-14',
        radius: 'rounded-br-full',
        color: 'bg-green-50',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.3 }
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
        color: 'bg-slate-50',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.3 }
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-green-200',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.6 }
      },
      {
        width: 'w-14',
        radius: 'rounded-full',
        color: 'bg-green-500',
        animate: { opacity: 1, scale: 1 },
        initial: { opacity: 0, scale: 0.5 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.6 }
      },
      {
        width: 'w-14',
        radius: 'rounded-tl-full',
        color: 'bg-green-100',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.6 }
      },
      {
        width: 'w-14',
        radius: 'rounded-tr-full',
        color: 'bg-slate-50',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 1.2 }
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
        color: 'bg-green-50',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.3 }
      },
      {
        width: 'w-14',
        radius: 'rounded-none',
        color: 'bg-green-50',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 0.6, delay: 0.6 }
      },
      {
        width: 'w-14',
        radius: 'rounded-br-full',
        color: 'bg-green-200',
        animate: { opacity: 1 },
        initial: { opacity: 0 },
        transition: { ease: 'easeInOut', duration: 1.2, delay: 0.6 }
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
            <MotionShape key={`shape${outerIndex}${index}`} {...shape}>
              {Math.floor(outerArray.length / 2) == outerIndex &&
                Math.floor(array.length / 2) == index && (
                  <Icon className='text-white'>check</Icon>
                )}
            </MotionShape>
          ))}
        </div>
      ))}
    </div>
  )
}

export function WalletShapes() {
  const radii: Radius[] = useMemo(
    () => [
      'rounded-full',
      'rounded-tl-full',
      'rounded-tr-full',
      'rounded-bl-full',
      'rounded-br-full',
      'rounded-t-full',
      'rounded-b-full',
      'rounded-l-full',
      'rounded-r-full'
    ],
    []
  )
  const colours = useMemo(
    () => [
      'bg-slate-600',
      'bg-rose-300',
      'bg-lime-400',
      'bg-rose-500',
      'bg-lime-300',
      'bg-lime-500',
      'bg-rose-400',
      'bg-slate-300',
      'bg-yellow-200',
      'bg-slate-500',
      'bg-rose-100',
      'bg-yellow-400'
    ],
    []
  )
  const shapes = Array(3)
    .fill({})
    .map((shape: MotionShapeProps, index: number) => {
      const colourIndex = Math.floor(Math.random() * colours.length)
      const radiiIndex = Math.floor(Math.random() * radii.length)
      return {
        color: colours[colourIndex],
        radius: radii[radiiIndex],
        width: 'w-0 lg:w-6',
        animate: { opacity: 1, scale: 1 },
        initial: { opacity: 0, scale: 0.5 },
        exit: {
          opacity: 0,
          scale: 0.5,
          transition: {
            duration: 0.2
          }
        },
        transition: {
          type: 'spring',
          stiffness: 400,
          damping: 20,
          duration: 0.3
        }
      }
    })

  return (
    <div className='flex items-center justify-center'>
      <AnimatePresence mode='popLayout'>
        {shapes.map((shape, index) => (
          <MotionShape key={`shape${index}${shape?.color}`} {...shape} />
        ))}
      </AnimatePresence>
    </div>
  )
}

export function HomeShapes() {
  return (
    <div className='mt-4 flex w-full flex-col items-center justify-center'>
      <div className='flex'>
        <Shape color='bg-yellow-200' radius='rounded-tr-full' width='w-8' />
        <Shape color='bg-rose-500' radius='rounded-full' width='w-8' />
      </div>
      <div className='flex'>
        <Shape color='bg-lime-500' radius='rounded-bl-full' width='w-8' />
        <Shape color='bg-lime-300' radius='rounded-br-full' width='w-8' />
      </div>
    </div>
  )
}

type LoadingShapesProps = {
  // Whether or not to animate the shapes in and out after initial load
  animate?: boolean
  rows?: number
}

export function LoadingShapes({
  animate = true,
  rows = 4
}: LoadingShapesProps) {
  const radii: Radius[] = useMemo(
    () => [
      'rounded-full',
      'rounded-tl-full',
      'rounded-tr-full',
      'rounded-bl-full',
      'rounded-br-full',
      'rounded-t-full',
      'rounded-b-full',
      'rounded-l-full',
      'rounded-r-full'
    ],
    []
  )
  const colours = useMemo(
    () => [
      'bg-slate-600',
      'bg-rose-300',
      'bg-lime-400',
      'bg-rose-500',
      'bg-lime-300',
      'bg-lime-500',
      'bg-rose-400',
      'bg-slate-300',
      'bg-yellow-200',
      'bg-slate-500',
      'bg-rose-100',
      'bg-yellow-400'
    ],
    []
  )
  const [shapes, setShapes] = useState<(MotionShapeProps | null)[][]>(
    Array(rows)
      .fill(Array(12).fill({}))
      .map((row) =>
        row.map((shape: MotionShapeProps, index: number) => {
          const width = index >= 10 ? 'w-0 lg:w-full' : 'w-full'
          const startingPercentage = animate ? 0.8 : 0.2
          if (Math.random() > startingPercentage) {
            const colourIndex = Math.floor(Math.random() * colours.length)
            const radiiIndex = Math.floor(Math.random() * radii.length)
            return {
              color: colours[colourIndex],
              radius: radii[radiiIndex],
              width,
              animate: { opacity: 1, scale: 1 },
              initial: { opacity: 0, scale: 0.5 },
              exit: {
                opacity: 0,
                scale: 0.5,
                transition: {
                  duration: 0.2
                }
              },
              transition: {
                type: 'spring',
                stiffness: 400,
                damping: 20,
                duration: 0.3
              }
            }
          } else
            return {
              color: 'bg-transparent',
              width,
              animate: { opacity: 1, scale: 1 },
              initial: { opacity: 0, scale: 0.5 },
              exit: { opacity: 0, scale: 0.5 }
            }
        })
      )
  )

  useEffect(() => {
    if (animate) {
      const interval: NodeJS.Timeout = setInterval(() => {
        if (document.visibilityState === 'visible') {
          const colourIndex = Math.floor(Math.random() * colours.length)
          const radiiIndex = Math.floor(Math.random() * radii.length)
          const firstIndex = Math.floor(Math.random() * rows)
          const secondIndex = Math.floor(Math.random() * 11)
          const width = secondIndex >= 10 ? 'w-0 lg:w-full' : 'w-full'
          const newShapes = [...shapes]
          if (newShapes[firstIndex][secondIndex]?.color == 'bg-transparent') {
            newShapes[firstIndex][secondIndex] = {
              color: colours[colourIndex],
              radius: radii[radiiIndex],
              width,
              animate: { opacity: 1, scale: 1 },
              initial: { opacity: 0, scale: 0.5 },
              exit: {
                opacity: 0,
                scale: 0.5,
                transition: {
                  duration: 0.2
                }
              },
              transition: {
                type: 'spring',
                stiffness: 400,
                damping: 20,
                duration: 0.3
              }
            }
          } else
            newShapes[firstIndex][secondIndex] = {
              color: 'bg-transparent',
              width,
              animate: { opacity: 1, scale: 1 },
              initial: { opacity: 0, scale: 0.5 },
              exit: { opacity: 0, scale: 0.5 }
            }
          setShapes(newShapes)
        }
      }, 600)

      return () => clearInterval(interval)
    }
  }, [shapes, animate, colours, radii, rows])

  return (
    <div>
      {shapes.map((shapeRow, outerIndex) => (
        <div className='flex justify-center' key={`shapeRow${outerIndex}`}>
          <AnimatePresence mode='popLayout'>
            {shapeRow.map((shape, index) => (
              <MotionShape
                key={`shape${outerIndex}${index}${shape?.color}`}
                {...shape}
              />
            ))}
          </AnimatePresence>
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
  const shapes: MotionShapeProps[][] = blogShapes[slug]
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
  [key: string]: MotionShapeProps[][]
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
  'card-payments-still-suck': [
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
  ],
  'our-fynbos-family-meet-don': [
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-slate-100'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-purple-200'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-orange-500'
      }
    ],
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-purple-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-orange-200'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-slate-200'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-purple-600'
      },
      {
        radius: 'rounded-full',
        color: 'bg-orange-500'
      }
    ]
  ],
  'the-future-digital-wallets-and-payment-pointers': [
    [
      {
        radius: 'rounded-full',
        color: 'bg-yellow-300'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-green-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-green-600'
      }
    ],
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-green-200'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-yellow-200'
      }
    ],
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-slate-200'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-green-500'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-orange-700'
      }
    ]
  ],
  'our-fynbos-family-meet-matt': [
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-blue-400'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-300'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-200'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-blue-200'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-green-500'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-yellow-300'
      }
    ],
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-green-700'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-slate-200'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-blue-500'
      }
    ]
  ],
  'our-fynbos-family-meet-adrian': [
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-blue-300'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-blue-100'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-500'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-200'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-slate-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-yellow-400'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-blue-400'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-600'
      }
    ]
  ],
  'our-fynbos-family-meet-cairin': [
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-orange-200'
      },
      {
        radius: 'rounded-full',
        color: 'bg-purple-300'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-blue-300'
      }
    ],
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-400'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-purple-600'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-blue-600'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-orange-100'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-300'
      }
    ]
  ],
  'our-fynbos-family-meet-justin': [
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-green-100'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-green-600'
      },
      {
        radius: 'rounded-full',
        color: 'bg-green-300'
      }
    ],
    [
      {
        radius: 'rounded-full',
        color: 'bg-yellow-300'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-slate-300'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-orange-200'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-orange-600'
      }
    ]
  ],
  'our-fynbos-family-meet-barnard': [
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-indigo-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-green-400'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-indigo-300'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-blue-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-yellow-200'
      },
      {
        radius: 'rounded-full',
        color: 'bg-orange-500'
      }
    ],
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-600'
      },
      {
        radius: 'rounded-full',
        color: 'bg-green-300'
      },
      {
        radius: 'rounded-full',
        color: 'bg-blue-400'
      }
    ]
  ],
  'why-payment-pointers-are-urls': [
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-slate-600'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-indigo-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-indigo-500'
      }
    ],
    [
      {
        radius: 'rounded-tr-full',
        color: 'bg-rose-400'
      },
      {
        radius: 'rounded-full',
        color: 'bg-indigo-300'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-slate-400'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-indigo-500'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-indigo-100'
      },
      {
        radius: 'rounded-full',
        color: 'bg-rose-300'
      }
    ]
  ],
  'our-fynbos-family-meet-omer': [
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-600'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-orange-400'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-indigo-500'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-yellow-200'
      },
      {
        radius: 'rounded-full',
        color: 'bg-orange-300'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-slate-200'
      }
    ],
    [
      {
        radius: 'rounded-full',
        color: 'bg-yellow-300'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-indigo-400'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-blue-100'
      }
    ]
  ],
  'how-technical-standards-promote-innovation': [
    [
      {
        radius: 'rounded-tl-full',
        color: 'bg-yellow-400'
      },
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-100'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-purple-400'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-200'
      },
      {
        radius: 'rounded-full',
        color: 'bg-green-400'
      },
      {
        radius: 'rounded-tr-full',
        color: 'bg-yellow-400'
      }
    ],
    [
      {
        radius: 'rounded-bl-full',
        color: 'bg-rose-300'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-500'
      },
      {
        radius: 'rounded-tl-full',
        color: 'bg-purple-600'
      }
    ]
  ],
  'joining-the-owf': [
    [
      {
        radius: 'rounded-l-full',
        color: 'bg-blue-600'
      },
      {
        radius: 'rounded-br-full',
        color: 'bg-slate-50'
      },
      {
        radius: 'rounded-full',
        color: 'bg-blue-100'
      }
    ],
    [
      {
        radius: 'rounded-l-full',
        color: 'bg-slate-50'
      },
      {
        radius: 'rounded-none',
        color: 'bg-purple-500'
      },
      {
        radius: 'rounded-r-full',
        color: 'bg-purple-600'
      }
    ],
    [
      {
        radius: 'rounded-br-full',
        color: 'bg-blue-100'
      },
      {
        radius: 'rounded-full',
        color: 'bg-orange-500'
      },
      {
        radius: 'rounded-r-full',
        color: 'bg-yellow-400'
      }
    ]
  ]
}
