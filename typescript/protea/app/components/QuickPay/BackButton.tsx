import { Link, useFetcher, useNavigate } from 'react-router'
import { flushSync } from 'react-dom'
import { useDialPadContext } from '~/lib/context/dialpad'
import type { QuickPaySession } from '~/lib/types'
import { clsx } from 'clsx'

interface BackButtonProps {
    title: string
    to: string
    clearSessionKeys?: Array<keyof QuickPaySession> | 'all'
    className?: string
    resetAmount?: boolean
}

export const BackButton = ({ title, to, clearSessionKeys, className, resetAmount }: BackButtonProps) => {
    const fetcher = useFetcher()
    const { setAmountValue } = useDialPadContext()
    const navigate = useNavigate()
    const handleBackClick = (e: React.MouseEvent, to: string, clearSessionKeys?: Array<keyof QuickPaySession> | 'all') => {
        e.preventDefault()
        if (resetAmount) {
            flushSync(() => {
                setAmountValue('0')
            })
        }
        if (clearSessionKeys) {
            fetcher.submit({ intent: 'updateSession', to, clearSessionKeys }, {
                method: 'POST'
            })
        }
        if (resetAmount && !clearSessionKeys) {
            navigate(to)
        }
    }
    return <div className={clsx('w-full text-left cursor-pointer hover:text-rose-600', className)}>
        {clearSessionKeys || resetAmount ? <Link to={to} onClick={(e) => { handleBackClick(e, to, clearSessionKeys) }}>&laquo; {title}</Link> : <Link to={to}>&laquo; {title}</Link>}
    </div>
}