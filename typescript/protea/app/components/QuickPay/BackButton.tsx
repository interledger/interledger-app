import { Link } from 'react-router'
import { clsx } from 'clsx'

interface BackButtonProps {
    title: string
    to: string
    className?: string
}

export const BackButton = ({ title, to, className }: BackButtonProps) => {

    return <div className={clsx('w-full text-left cursor-pointer hover:text-rose-600', className)}>
        {<Link to={to}>&laquo; {title}</Link>}
    </div>
}