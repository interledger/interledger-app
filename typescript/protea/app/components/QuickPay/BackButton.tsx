import { Link, useFetcher, useNavigate } from '@remix-run/react'
import { flushSync } from 'react-dom'
import type { QuickPaySession } from '~/lib/types'
import { commitSession } from '~/session.server'
import { redirect, type Session } from '@remix-run/node'
import { useDialPadContext } from '~/lib/context/dialpad'
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
    return <div className={clsx('w-full cursor-pointer hover:text-rose-600', className)}>
        {clearSessionKeys || resetAmount ? <Link to={to} onClick={(e) => { handleBackClick(e, to, clearSessionKeys) }}>&laquo; {title}</Link> : <Link to={to}>&laquo; {title}</Link>}
    </div>
}
export const handleSessionUpdate = async (session: Session, intent: string, to: string, clearSessionKeys: Array<keyof QuickPaySession> | 'all') => {
    if (intent === 'updateSession') {
        let sessionData = session.get('quickPay')
        if (clearSessionKeys === 'all') {
            sessionData = undefined
        } else {
            for (const key of clearSessionKeys) { sessionData[key] = undefined }
        }
        session.set('quickPay', sessionData)

        return redirect(to, {
            headers: { 'Set-Cookie': await commitSession(session) }
        })
    }
}