import { Form, useNavigation } from '@remix-run/react'
import { Button, Card, CardContent, Icon, TextField } from '~/components'
import { Label } from '~/components/Label'

export const TotpChallenge = ({
  flowId,
  csrfToken,
  actionData
}: {
  flowId: string
  csrfToken: string
  actionData: any
}) => {
  const navigation = useNavigation()
  const isSubmitting = navigation.state === 'submitting'
  return (
    <>
      <Form method='post'>
        <input type='hidden' name='csrf_token' value={csrfToken} />
        <input type='hidden' name='flow' value={flowId} />
        <Card>
          <CardContent>
            <p>
              Enter the 6-digit code from your authenticator app to continue.
            </p>
          </CardContent>
          <Label className='mt-2'>Authenticator App</Label>
          <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
            <Icon>shield</Icon>
            <span>Time-based One-Time Password (TOTP)</span>
          </div>
          <TextField
            id='totp'
            label='Authenticator Code'
            name='totp_code'
            type='text'
            className='mt-4'
            aria-invalid={Boolean(actionData?.errors?.totp_code) || undefined}
            aria-describedby={
              actionData?.errors?.totp_code ? 'totp-error' : undefined
            }
            required
            errorMessage={actionData?.errors?.totp_code}
          />
        </Card>
        <Button type='submit' className='mt-4' disabled={isSubmitting}>
        {isSubmitting ? 'Verifying…' : 'Verify'}
        </Button>
      </Form>
    </>
  )
}
