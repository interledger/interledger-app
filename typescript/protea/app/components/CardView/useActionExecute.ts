import { useCallback, useState } from 'react'

export const DELAY_AFTER_ACTION = 3000
export type ActionStatus = 'idle' | 'loading' | 'success' | 'error'

export const useActionExecute = () => {
  const [actionStatus, setActionStatus] = useState<ActionStatus>('idle')

  const resetStatusDelayed = useCallback(
    (delay: number) => {
      setTimeout(() => setActionStatus('idle'), delay)
    },
    [setActionStatus]
  )

  const resetStatus = useCallback(() => {
    setActionStatus('idle')
  }, [])

  const errorStatus = useCallback(() => {
    setActionStatus('error')
    resetStatusDelayed(DELAY_AFTER_ACTION)
  }, [resetStatusDelayed, setActionStatus])

  const executeAction = useCallback(
    async ({
      execute,
      onSuccess,
      onError
    }: {
      execute: () => Promise<void>
      onSuccess?: () => void
      onError?: (error: any) => void
    }) => {
      try {
        setActionStatus('loading')

        await execute()
        onSuccess?.()

        setActionStatus('success')
        resetStatusDelayed(DELAY_AFTER_ACTION)
      } catch (error) {
        onError?.(error)

        setActionStatus('error')
        resetStatusDelayed(DELAY_AFTER_ACTION)
      }
    },
    [resetStatusDelayed, setActionStatus]
  )

  return {
    actionStatus,
    executeAction,
    setActionStatus,
    resetStatus,
    errorStatus
  }
}
