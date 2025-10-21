import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import { useEffect, useState } from 'react'
import { Icon } from '~/components/Icon'

interface TimedPinPopupProps {
  pin: string
  isVisible: boolean
  onClose: () => void
  duration?: number // in seconds
  className?: string
}

export const TimedPinPopup = ({
  pin,
  isVisible,
  onClose,
  duration = 7,
  className
}: TimedPinPopupProps) => {
  const [timeRemaining, setTimeRemaining] = useState(duration)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!isVisible) {
      setTimeRemaining(duration)
      setCopied(false)
      return
    }

    const interval = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval)
          onClose()
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [isVisible, duration, onClose])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(pin)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy PIN:', err)
    }
  }

  return (
    <AnimatePresence>
      {isVisible && (
        <div
          className={clsx(
            'fixed inset-0 z-50 flex items-center justify-center bg-scrim/75 px-4 backdrop-blur-sm',
            className
          )}
        >
          <motion.div
            key='pin-popup'
            animate={{ opacity: 1, scale: 1 }}
            initial={{ opacity: 0, scale: 0.75 }}
            exit={{
              opacity: 0,
              scale: 0.75,
              transition: {
                duration: 0.2
              }
            }}
            transition={{
              duration: 0.2,
              ease: 'easeInOut'
            }}
            className='relative w-full max-w-xs rounded-3xl bg-container-strong p-2 shadow-lg'
          >
            {/* Close button */}
            <button
              onClick={onClose}
              className='absolute right-4 top-4 rounded-full p-1 text-medium transition-colors hover:bg-nav hover:text-strong'
              aria-label='Close PIN display'
            >
              <Icon className='text-xl'>close</Icon>
            </button>

            {/* Timer badge */}
            <div className='absolute left-4 top-4 flex items-center space-x-1 rounded-full bg-nav px-2.5 py-1 text-xs font-semibold text-medium'>
              <Icon className='text-sm'>schedule</Icon>
              <span>{timeRemaining}s</span>
            </div>

            {/* PIN content */}
            <div className='mt-12 px-2 pb-2 pt-4 text-center'>
              <h1 className='mb-4 text-xl font-medium text-strong'>Card PIN</h1>
              <button
                onClick={handleCopy}
                className='mb-4 w-full rounded-xl bg-nav px-6 py-4 transition-colors hover:bg-nav-hover focus-visible:outline-2 focus-visible:outline-focus'
                aria-label='Copy PIN to clipboard'
              >
                <div className='flex items-center justify-center space-x-3'>
                  <div className='text-4xl font-bold tracking-[0.5em] text-strong'>
                    {pin}
                  </div>
                  <Icon className='text-2xl text-medium'>
                    {copied ? 'check' : 'content_copy'}
                  </Icon>
                </div>
              </button>
              {copied && (
                <p className='text-sm text-medium'>Copied to clipboard!</p>
              )}
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
