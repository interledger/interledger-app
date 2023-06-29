import { useLocation } from '@remix-run/react'
import clsx from 'clsx'
import type { MotionProps } from 'framer-motion'
import { AnimatePresence, motion } from 'framer-motion'
import type { ReactNode } from 'react'
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

// Client-only components - stops the shapes' classname from changing on initial load
// https://remix.run/docs/en/1.18.0/guides/migrating-react-router-app#client-only-components
let isHydrating = true

export function WalletShapes() {
  const location = useLocation()

  const [isHydrated, setIsHydrated] = useState(!isHydrating)

  useEffect(() => {
    isHydrating = false
    setIsHydrated(true)
  }, [])

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

    const commonShape = {
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

    return Array(3)
      .fill({ ...commonShape })
      .map((shape: MotionShapeProps, index: number) => {
        const colourIndex = Math.floor(Math.random() * colours.length)
        const radiiIndex = Math.floor(Math.random() * radii.length)

        return {
          key: `shape${index}${location.pathname}`,
          color: colours[colourIndex],
          radius: radii[radiiIndex],
          ...shape
        }
      })
  }, [location.pathname])

  if (!isHydrated) {
    return null
  }

  return (
    <div className='flex items-center justify-center'>
      <AnimatePresence mode='popLayout'>
        {shapes.map(({ key, ...props }) => (
          <MotionShape key={key} {...props} />
        ))}
      </AnimatePresence>
    </div>
  )
}

export function LoadingShapes() {
  const [isHydrated, setIsHydrated] = useState(!isHydrating)

  useEffect(() => {
    isHydrating = false
    setIsHydrated(true)
  }, [])

  if (!isHydrated) {
    return null
  }

  const container = {
    hidden: {},
    show: {
      transition: {
        delayChildren: 0.5,
        staggerChildren: 0.1,
        repeat: Infinity
      }
    }
  }

  const item = {
    hidden: { opacity: 0, scale: 0.5 },
    show: {
      opacity: 1,
      scale: 1
    }
  }

  return (
    <div className='flex items-center justify-center'>
      <motion.div
        className='flex'
        variants={container}
        initial='hidden'
        animate='show'
      >
        <motion.div
          className='h-6 w-6 rounded-full bg-rose-500'
          variants={item}
          transition={{
            type: 'spring',
            stiffness: 400,
            damping: 20,
            repeat: Infinity,
            repeatDelay: 0.35,
            repeatType: 'reverse'
          }}
        />
        <motion.div
          className='h-6 w-6 rounded-bl-full rounded-tr-full bg-lime-500'
          variants={item}
          transition={{
            type: 'spring',
            stiffness: 400,
            damping: 20,
            repeat: Infinity,
            repeatDelay: 0.35,
            repeatType: 'reverse'
          }}
        />
        <motion.div
          className='h-6 w-6 rounded-l-full bg-sky-500'
          variants={item}
          transition={{
            type: 'spring',
            stiffness: 400,
            damping: 20,
            repeat: Infinity,
            repeatDelay: 0.35,
            repeatType: 'reverse'
          }}
        />
      </motion.div>
    </div>
  )
}
