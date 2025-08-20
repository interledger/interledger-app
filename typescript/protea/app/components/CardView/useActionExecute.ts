import { useState } from 'react'

export const DELAY_AFTER_ACTION = 3000
export type ActionStatus = 'idle' | 'loading' | 'success' | 'error'

export const useActionExecute = () => {
  const [actionStatus, setActionStatus] = useState<ActionStatus>('idle')

  const resetStatusDelayed = (delay: number) => {
    setTimeout(() => setActionStatus('idle'), delay)
  }

  const resetStatus = () => {
    setActionStatus('idle')
  }

  const executeAction = async ({
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
  }

  return {
    actionStatus,
    executeAction,
    resetStatus
  }
}
