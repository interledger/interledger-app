import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * Hook for countdown timer
 * @returns Object with start function, isActive flag, and remainingSeconds
 *
 * Usage:
 * const { start, isActive, remainingSeconds } = useCountdown()
 *
 * <button onClick={() => start(60)} disabled={isActive}>
 *   {isActive ? `Wait ${remainingSeconds}s` : 'Click me'}
 * </button>
 */
export const useCountdown = () => {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const [remainingSeconds, setRemainingSeconds] = useState(0)
  const endTimeRef = useRef<number | null>(null)

  const updateRemainingTime = useCallback(() => {
    if (!endTimeRef.current) return

    const remaining = Math.max(
      0,
      Math.ceil((endTimeRef.current - Date.now()) / 1000)
    )
    setRemainingSeconds(remaining)

    if (remaining <= 0) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
      endTimeRef.current = null
    }
  }, [])

  const cleanup = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    endTimeRef.current = null
    setRemainingSeconds(0)
  }, [])

  useEffect(() => {
    return cleanup
  }, [cleanup])

  const start = useCallback(
    (durationMs: number) => {
      cleanup()

      endTimeRef.current = Date.now() + durationMs
      setRemainingSeconds(Math.ceil(durationMs / 1000))

      timeoutRef.current = setTimeout(cleanup, durationMs)
      intervalRef.current = setInterval(updateRemainingTime, 1000)
      updateRemainingTime()
    },
    [cleanup, updateRemainingTime]
  )

  return {
    start,
    isActive: remainingSeconds > 0,
    remainingSeconds
  }
}
