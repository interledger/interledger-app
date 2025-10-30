import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import { useState } from 'react'
import { Button, Card, CardContent, Icon, TextField } from '~/components'
import { Label } from '~/components/Label'
import { usePinChangePopupStore } from '~/lib/usePinChangePopupStore'

interface PinChangePopupProps {
  className?: string
}

export const PinChangePopup = ({ className }: PinChangePopupProps) => {
  const {
    isOpen,
    closePopup,
    handleSubmit: onSubmit
  } = usePinChangePopupStore()
  const [pin, setPin] = useState('')
  const [error, setError] = useState<string | undefined>()

  const handlePinChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    // Only allow up to 4 characters
    if (value.length <= 4) {
      setPin(value)
      setError(undefined)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    // Validate: must be exactly 4 digits
    if (pin.length !== 4) {
      setError('PIN must be exactly 4 digits')
      return
    }

    if (!/^\d{4}$/.test(pin)) {
      setError('PIN must contain only digits')
      return
    }

    onSubmit(pin)
    // Reset state
    setPin('')
    setError(undefined)
  }

  const handleClose = () => {
    setPin('')
    setError(undefined)
    closePopup()
  }

  const isSubmitDisabled = pin.length !== 4

  return (
    <AnimatePresence>
      {isOpen && (
        <div
          className={clsx(
            'fixed inset-0 z-50 flex items-center justify-center bg-scrim/75 px-4 backdrop-blur-sm',
            className
          )}
        >
          <motion.div
            key='pin-change-popup'
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
            className='relative w-full max-w-md'
          >
            <form onSubmit={handleSubmit}>
              <Card>
                {/* Close button */}
                <div className='flex justify-end'>
                  <button
                    type='button'
                    onClick={handleClose}
                    className='rounded-lg p-2 text-medium hover:bg-nav'
                    aria-label='Close'
                  >
                    <Icon>close</Icon>
                  </button>
                </div>

                <CardContent>
                  <h2 className='mb-4 text-xl font-semibold'>
                    Change Card PIN
                  </h2>
                  <p className='text-medium'>
                    Enter a new 4-digit PIN for your card.
                  </p>
                </CardContent>

                <Label className='mt-2'>New PIN</Label>
                <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
                  <Icon>lock</Icon>
                  <span>4-digit numeric PIN</span>
                </div>

                <TextField
                  id='pin-change'
                  label='Enter New PIN'
                  name='new_pin'
                  type='password'
                  value={pin}
                  onChange={handlePinChange}
                  className='mt-4'
                  aria-invalid={Boolean(error) || undefined}
                  aria-describedby={error ? 'pin-error' : undefined}
                  required
                  autoFocus
                  errorMessage={error}
                  inputMode='numeric'
                  pattern='[0-9]*'
                  maxLength={4}
                />

                <div className='mt-4 flex gap-2'>
                  <Button
                    type='button'
                    onClick={handleClose}
                    className='flex-1'
                  >
                    Cancel
                  </Button>
                  <Button
                    type='submit'
                    className='flex-1'
                    disabled={isSubmitDisabled}
                    aria-label='Submit new PIN'
                  >
                    Submit
                  </Button>
                </div>
              </Card>
            </form>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
