import { useState } from 'react'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  TextButton
} from '~/components'
import type { PlaidState } from '~/lib/plaid.server'
import { usePlaidStore } from '~/lib/usePlaidStore'

interface DebugPanelProps {
  userId: string
  state: PlaidState
  actionData: unknown
}

/**
 * JsonCard — a Card whose JSON body collapses behind a "Show JSON" toggle.
 * Reused by every panel inside the debug surface so the page isn't a wall of
 * pre tags by default.
 */
function JsonCard({
  title,
  payload,
  defaultOpen = false
}: {
  title: string
  payload: unknown
  defaultOpen?: boolean
}) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <Card>
      <CardHeader>
        <div className='flex items-center justify-between'>
          <CardTitle>{title}</CardTitle>
          <TextButton onClick={() => setOpen((v) => !v)} aria-expanded={open}>
            {open ? 'Hide JSON' : 'Show JSON'}
          </TextButton>
        </div>
      </CardHeader>
      {open && (
        <CardContent>
          <pre className='overflow-x-auto whitespace-pre-wrap break-all text-xs'>
            {JSON.stringify(payload, null, 2)}
          </pre>
        </CardContent>
      )}
    </Card>
  )
}

export function DebugPanel({ userId, state, actionData }: DebugPanelProps) {
  const { linkToken, lastError, lastResponses } = usePlaidStore()

  return (
    <div className='mt-8 flex flex-col gap-6'>
      {linkToken && (
        <Card>
          <CardHeader>
            <CardTitle>Pending link</CardTitle>
          </CardHeader>
          <CardContent>
            <p>link_token issued; F5a will hand it to react-plaid-link.</p>
            <code className='block break-all text-xs'>{linkToken}</code>
          </CardContent>
        </Card>
      )}

      {lastError && (
        <Card>
          <CardHeader>
            <CardTitle>Last error</CardTitle>
          </CardHeader>
          <CardContent>
            <code className='break-all text-error'>{lastError}</code>
          </CardContent>
        </Card>
      )}

      {actionData != null && (
        <JsonCard title='Debug — last action result' payload={actionData} />
      )}

      <Card>
        <CardHeader>
          <CardTitle>Debug — session</CardTitle>
        </CardHeader>
        <CardContent>
          <p>
            user_id:{' '}
            <code className='break-all'>
              {userId || '(unauthenticated)'}
            </code>
          </p>
        </CardContent>
      </Card>

      <JsonCard title='Debug — loader state' payload={state} />
      <JsonCard
        title='Debug — cached product responses'
        payload={lastResponses}
      />
    </div>
  )
}
