import { AnimatePresence, motion } from 'framer-motion'

interface ActionMessageProps {
  className?: string
  message?: string
}

export function ActionMessage({ className = 'min-w-full', message }: ActionMessageProps) {
  return (
    <div className={className}>
      <div className="min-h-[1.75rem] pl-2 pt-2">
        <AnimatePresence>
          {message && (
            <motion.p
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ type: 'spring', stiffness: 400, damping: 20 }}
              className="text-sm text-error"
            >
              {message}
            </motion.p>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}