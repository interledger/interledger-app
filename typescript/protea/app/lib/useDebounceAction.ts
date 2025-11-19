import { useCallback, useEffect, useRef } from 'react'

/**
 * Hook to debounce an action - prevents action from firing until delay has passed
 * @param delay - The delay in milliseconds (default: 60000 = 1 minute)
 * @returns Object with execute function and isDebouncing flag
 *
 * Usage:
 * const { execute, isDebouncing } = useDebounceAction(60000)
 *
 * const handleClick = () => {
 *   execute(() => {
 *     console.log("Action executed!")
 *   })
 * }
 *
 * <button onClick={handleClick} disabled={isDebouncing}>
 *   Click me
 * </button>
 */
export const useDebounceAction = (delay = 60000) => {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current)
    }
  }, [])

  const execute = useCallback(
    <T extends (...args: any[]) => void>(action: T, ...args: Parameters<T>) => {
      if (timeoutRef.current) return

      action(...args)

      timeoutRef.current = setTimeout(() => {
        timeoutRef.current = null
      }, delay)
    },
    [delay]
  )

  return execute
}
