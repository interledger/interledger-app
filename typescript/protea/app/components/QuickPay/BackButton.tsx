import { clsx } from 'clsx'
import { Link } from 'react-router'

interface BackButtonProps {
  title: string
  to: string
  className?: string
}

export const BackButton = ({ title, to, className }: BackButtonProps) => {
  return (
    <div
      className={clsx(
        'w-full cursor-pointer text-left hover:text-rose-600',
        className
      )}
    >
      {<Link to={to}>&laquo; {title}</Link>}
    </div>
  )
}
