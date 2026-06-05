import clsx from 'clsx'
import { motion } from 'framer-motion'
import { Icon } from '~/components/Icon'

type StatusType = 'success' | 'error' | 'warning' | 'info'

interface StatusPopupProps {
  type: StatusType
  message: string
  className?: string
}

const statusConfig = {
  success: {
    bgColor: 'bg-green-500',
    textColor: 'text-white',
    icon: 'check_circle'
  },
  error: {
    bgColor: 'bg-red-500',
    textColor: 'text-white',
    icon: 'error'
  },
  warning: {
    bgColor: 'bg-yellow-500',
    textColor: 'text-white',
    icon: 'warning'
  },
  info: {
    bgColor: 'bg-blue-500',
    textColor: 'text-white',
    icon: 'info'
  }
} as const

export const StatusPopup = ({ type, message, className }: StatusPopupProps) => {
  const config = statusConfig[type]

  return (
    <div
      className={clsx(
        'fixed bottom-4 left-0 z-50 mx-auto flex w-full justify-center px-4',
        className
      )}
    >
      <motion.div
        key={type + message}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        initial={{ opacity: 0, scale: 0.5, y: 8 }}
        exit={{
          opacity: 0,
          scale: 0.5,
          y: 8,
          transition: {
            duration: 0.2
          }
        }}
        transition={{
          type: 'spring',
          stiffness: 400,
          damping: 20,
          duration: 0.3
        }}
        className={clsx(
          'flex w-full items-center space-x-3 overflow-hidden rounded-xl px-4 py-3 text-left shadow-lg sm:max-w-xs',
          config.bgColor,
          config.textColor
        )}
      >
        <Icon>{config.icon}</Icon>
        <span className='text-sm'>{message}</span>
      </motion.div>
    </div>
  )
}
