import clsx from 'clsx'
import { motion } from 'framer-motion'
import type { FC, ReactNode } from 'react'

// TODO: refactor children to use label prop rather.
type IconProps = {
  /**
   * The name of the material icon:
   * https://fonts.google.com/icons
   */
  children?: ReactNode
  className?: string
}

export const Icon: FC<IconProps> = ({ className, children }) => {
  return (
    <span
      className={clsx(
        'inline-block h-6 w-6 select-none whitespace-nowrap fill-current text-center font-icon text-2xl font-normal normal-case not-italic leading-6 tracking-normal antialiased',
        className
      )}
    >
      {children}
    </span>
  )
}

export function AnimatedSchedule() {
  const duration = 2

  const getVariant = (duration: number) => ({
    from: {
      rotate: 0,
      transition: { duration: 0, ease: 'linear' }
    },
    to: {
      rotate: 360,
      transition: { duration, ease: 'linear', repeat: Infinity }
    }
  })

  return (
    <div className='grid h-6 w-6'>
      {/* Minute hand */}
      <motion.div
        style={{
          gridColumn: 1,
          gridRow: 1
        }}
        variants={getVariant(duration)}
        initial={'from'}
        animate={'to'}
      >
        <svg
          width='24'
          height='24'
          viewBox='0 0 24 24'
          fill='none'
          xmlns='http://www.w3.org/2000/svg'
        >
          <g clipPath='url(#clip0_2538_2239)'>
            <rect
              x='11'
              y='6.34155'
              width='2'
              height='5.65845'
              className='fill-current'
            />
          </g>
          <defs>
            <clipPath id='clip0_2538_2239'>
              <rect width='24' height='24' fill='white' />
            </clipPath>
          </defs>
        </svg>
      </motion.div>
      {/* Hour hand*/}
      <motion.div
        style={{
          gridColumn: 1,
          gridRow: 1
        }}
        variants={getVariant(duration * 6)}
        initial={'from'}
        animate={'to'}
      >
        <svg
          width='24'
          height='24'
          viewBox='0 0 24 24'
          fill='none'
          xmlns='http://www.w3.org/2000/svg'
        >
          <g clipPath='url(#clip0_2538_2245)'>
            <rect x='11' y='7' width='2' height='5' className='fill-current' />
          </g>
          <defs>
            <clipPath id='clip0_2538_2245'>
              <rect width='24' height='24' fill='white' />
            </clipPath>
          </defs>
        </svg>
      </motion.div>
      <svg
        style={{
          gridColumn: 1,
          gridRow: 1
        }}
        width='24'
        height='24'
        viewBox='0 0 24 24'
        fill='none'
        xmlns='http://www.w3.org/2000/svg'
      >
        <path
          d='M12 11C11.4477 11 11 11.4477 11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11Z'
          className='fill-current'
        />
        <path
          fillRule='evenodd'
          clipRule='evenodd'
          d='M2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12ZM4 12C4 7.58172 7.58172 4 12 4C16.4183 4 20 7.58172 20 12C20 16.4183 16.4183 20 12 20C7.58172 20 4 16.4183 4 12Z'
          className='fill-current'
        />
      </svg>
    </div>
  )
}
