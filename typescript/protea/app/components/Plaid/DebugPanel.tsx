import { Card, CardContent, CardHeader, CardTitle } from '~/components'
import type { PlaidState } from '~/lib/plaid.server'
import { usePlaidStore } from '~/lib/usePlaidStore'

interface DebugPanelProps {
  userId: string
  state: PlaidState
  actionData: unknown
}

export function DebugPanel({ userId, state, actionData }: DebugPanelProps) {
  const { linkToken, lastError, lastResponses } = usePlaidStore()

  return (
    <div className="mt-8 flex flex-col gap-6">
      {linkToken && (
        <Card>
          <CardHeader>
            <CardTitle>Pending link</CardTitle>
          </CardHeader>
          <CardContent>
            <p>link_token issued; F5a will hand it to react-plaid-link.</p>
            <code className="block break-all text-xs">{linkToken}</code>
          </CardContent>
        </Card>
      )}

      {lastError && (
        <Card>
          <CardHeader>
            <CardTitle>Last error</CardTitle>
          </CardHeader>
          <CardContent>
            <code className="break-all text-error">{lastError}</code>
          </CardContent>
        </Card>
      )}

      {actionData && (
        <Card>
          <CardHeader>
            <CardTitle>Debug — last action result</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="overflow-x-auto whitespace-pre-wrap break-all text-xs">
              {JSON.stringify(actionData, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Debug — session</CardTitle>
        </CardHeader>
        <CardContent>
          <p>
            user_id: <code className="break-all">{userId || '(unauthenticated)'}</code>
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Debug — loader state</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="overflow-x-auto whitespace-pre-wrap break-all text-xs">
            {JSON.stringify(state, null, 2)}
          </pre>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Debug — cached product responses</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="overflow-x-auto whitespace-pre-wrap break-all text-xs">
            {JSON.stringify(lastResponses, null, 2)}
          </pre>
        </CardContent>
      </Card>
    </div>
  )
}
