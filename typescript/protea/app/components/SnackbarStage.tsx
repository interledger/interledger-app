import clsx from 'clsx'
import { motion } from 'framer-motion'
import { forwardRef, useEffect, useRef, useState } from 'react'
import { href, useNavigate } from 'react-router'
import { IconButton, TextButton } from '~/components/Buttons'
import type { SnackbarAction, SnackbarType } from '~/lib/useScaffoldStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

const Stage = forwardRef<any>(({ ...motionProps }, ref) => {
  const [snackbar, setSnackbar] = useState<SnackbarType | null>(null)
  let dismissRef = useRef<NodeJS.Timeout>()
  const navigate = useNavigate()

  const [snackbars, shiftSnackbar] = useScaffoldStore((state) => [
    state.snackbars,
    state.shiftSnackbar
  ])

  const actionReducer = (action?: SnackbarAction) => {
    switch (action) {
      case 'Contact support':
        shiftSnackbar()
        navigate(href('/support'))
        break
      case 'View cards':
        shiftSnackbar()
        navigate(href('/cards'))
        break
      default:
        shiftSnackbar()
    }
  }

  useEffect(() => {
    const visibleSnackbar = snackbars.find((s) => s.canShow)
    if (visibleSnackbar) {
      setSnackbar(visibleSnackbar)
    } else {
      setSnackbar(null)
    }
  }, [shiftSnackbar, snackbar?.id, snackbars])

  useEffect(() => {
    if (snackbar?.id) {
      dismissRef.current = setTimeout(
        () => {
          shiftSnackbar()
        },
        snackbar.action ? 10000 : 4000
      )
    }
    return () => {
      clearTimeout(dismissRef.current)
    }
  }, [shiftSnackbar, snackbar?.action, snackbar?.id])

  return (
    <>
      {snackbar && (
        <motion.div
          ref={ref}
          {...motionProps}
          key={snackbar.id}
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
            'z-50 flex w-full overflow-hidden rounded-xl bg-snackbar px-4 py-3 text-left shadow-lg sm:max-w-xs',
            snackbar.action && snackbar.icon
              ? 'flex-col items-start gap-y-3'
              : 'items-center gap-x-3'
          )}
        >
          <p className='whitespace-pre-line text-sm text-inverted'>
            {snackbar.message}
          </p>
          <div className='ml-auto flex items-center gap-x-3'>
            {snackbar.action && (
              <TextButton onClick={() => actionReducer(snackbar?.action)}>
                {snackbar.action}
              </TextButton>
            )}
            {snackbar.icon && (
              <IconButton
                className='text-inverted'
                onClick={() => shiftSnackbar()}
              >
                {snackbar.icon}
              </IconButton>
            )}
          </div>
        </motion.div>
      )}
    </>
  )
})

Stage.displayName = 'SnackbarStage'

export const SnackbarStage = motion(Stage, { forwardMotionProps: true })
