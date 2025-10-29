import { useEffect, useState } from 'react'

interface TimeoutDisplayProps {
  purchaseDate: string // timestamp in milliseconds
  timeout: string // timeout in seconds
}

export function TimeoutDisplay({ purchaseDate, timeout }: TimeoutDisplayProps) {
  const [timeLeft, setTimeLeft] = useState<number>(0)

  useEffect(() => {
    const calculateTimeLeft = () => {
      const purchaseDateTimestamp = new Date(purchaseDate).valueOf()
      const timeoutSeconds = Number(timeout)
      const now = Date.now()

      // Calculate time passed in seconds
      const timePassed = Math.floor((now - purchaseDateTimestamp) / 1000)

      // Calculate remaining time in seconds
      const remaining = Math.max(0, timeoutSeconds - timePassed)

      return remaining
    }

    // Initial calculation
    setTimeLeft(calculateTimeLeft())

    // Update every second
    const interval = setInterval(() => {
      setTimeLeft(calculateTimeLeft())
    }, 1000)

    return () => clearInterval(interval)
  }, [purchaseDate, timeout])

  // Format as mm:ss
  const minutes = Math.floor(timeLeft / 60)
  const seconds = timeLeft % 60
  const formatted = `${minutes}:${seconds.toString().padStart(2, '0')}`

  return <span className='text-xs text-weak'>{formatted}</span>
}
