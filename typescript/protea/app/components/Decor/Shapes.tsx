import clsx from 'clsx'
import type { MotionProps } from 'framer-motion'
import { motion, stagger, useAnimate, usePresence } from 'framer-motion'
import type { ReactNode } from 'react'
import { forwardRef, useEffect, useMemo, useRef, useState } from 'react'
import { Icon } from '~/components'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

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

// Client-only components - stops the shapes' classname from changing on initial load
// https://remix.run/docs/en/1.18.0/guides/migrating-react-router-app#client-only-components
let isHydrating = true

export function WalletShapes() {
  const [isPresent, safeToRemove] = usePresence()
  const [isHydrated, setIsHydrated] = useState(!isHydrating)

  const [loading] = useScaffoldStore((state) => [state.loading])

  const throttleRef = useRef<NodeJS.Timeout>()

  const [scope, animate] = useAnimate()

  const shapes = useMemo(() => {
    const radii: Radius[] = [
      'rounded-full',
      'rounded-tl-full',
      'rounded-tr-full',
      'rounded-bl-full',
      'rounded-br-full',
      'rounded-t-full',
      'rounded-b-full',
      'rounded-l-full',
      'rounded-r-full',
      'rounded-full rounded-tl-none',
      'rounded-full rounded-tr-none',
      'rounded-full rounded-bl-none',
      'rounded-full rounded-br-none',
      'rounded-tr-full rounded-bl-full',
      'rounded-tl-full rounded-br-full'
    ]
    const colours = [
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
    ]

    return ['', '', ''].map(() => {
      const colourIndex = Math.floor(Math.random() * colours.length)
      const radiiIndex = Math.floor(Math.random() * radii.length)
      return `w-6 h-6 ${colours[colourIndex]} ${radii[radiiIndex]}`
    })
  }, [])

  useEffect(() => {
    isHydrating = false
    setIsHydrated(true)
  }, [])

  useEffect(() => {
    if (isHydrated) {
      if (isPresent) {
        animate(
          'li',
          { opacity: 1, scale: 1 },
          {
            delay: stagger(0.1),
            type: 'spring',
            stiffness: 400,
            damping: 20
          }
        )
      } else {
        animate(
          'li',
          { opacity: 0, scale: 0.5 },
          {
            delay: stagger(0.1),
            type: 'spring',
            stiffness: 400,
            damping: 20
          }
        ).then(() => safeToRemove())
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPresent, isHydrated])

  useEffect(() => {
    const loadingAnimation = () => {
      animate([
        [
          'li',
          { opacity: 0, scale: 0.5 },
          {
            delay: stagger(0.1),
            type: 'spring',
            stiffness: 400,
            damping: 20
          }
        ],
        [
          'li',
          { opacity: 1, scale: 1 },
          {
            delay: stagger(0.1),
            type: 'spring',
            stiffness: 400,
            damping: 20
          }
        ]
      ])
    }

    if (loading) {
      loadingAnimation()
      throttleRef.current = setInterval(async () => {
        loadingAnimation()
      }, 1800)
    } else {
      clearTimeout(throttleRef.current)
    }
    return () => {
      clearTimeout(throttleRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading])

  if (!isHydrated) {
    return null
  }
  return (
    <ul ref={scope} className='flex items-center justify-center'>
      <li
        style={{
          opacity: 0,
          scale: 0.5
        }}
        className={shapes[0]}
      />
      <li
        style={{
          opacity: 0,
          scale: 0.5
        }}
        className={shapes[1]}
      />
      <li
        style={{
          opacity: 0,
          scale: 0.5
        }}
        className={shapes[2]}
      />
    </ul>
  )
}
