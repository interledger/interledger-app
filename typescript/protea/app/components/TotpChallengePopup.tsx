import { useFetcher } from '@remix-run/react'
import { useEffect } from 'react'
import { Button, Card, CardContent, Icon, TextField } from '~/components'
import { Label } from '~/components/Label'

interface TotpChallengePopupProps {
  flowId: string
  csrfToken: string
  isOpen: boolean
  onClose: () => void
  onSuccess?: () => void
  onError?: (error: string) => void
}

export const TotpChallengePopup = ({
  flowId,
  csrfToken,
  isOpen,
  onClose,
  onSuccess,
  onError
}: TotpChallengePopupProps) => {
  const fetcher = useFetcher<{
    success: boolean
    error?: string
    flowId?: string
  }>()

  const isSubmitting = fetcher.state === 'submitting'

  console.log('🍀 (TotpChallengePopup) props:', {
    flowId,
    csrfToken,
    isOpen,
    onClose,
    onSuccess,
    onError
  })

  // Handle verification result
  useEffect(() => {
    if (fetcher.data) {
      if (fetcher.data.success) {
        console.log('✅ TOTP verified successfully')
        onSuccess?.()
        onClose()
      } else if (fetcher.data.error) {
        console.error('❌ TOTP verification failed:', fetcher.data.error)
        onError?.(fetcher.data.error)
      }
    }
  }, [fetcher.data, onSuccess, onError, onClose])

  if (!isOpen) return null

  return (
    <div className='fixed inset-0 z-50 flex items-center justify-center'>
      {/* Backdrop */}
      <div
        className='absolute inset-0 bg-black/50 backdrop-blur-sm'
        onClick={onClose}
        aria-hidden='true'
      />

      {/* Modal */}
      <div className='relative z-10 w-full max-w-md p-4'>
        <fetcher.Form method='post' action='/api/totp-challenge-verify'>
          <input type='hidden' name='csrf_token' value={csrfToken} />
          <input type='hidden' name='flow' value={flowId} />

          <Card>
            {/* Close button */}
            <div className='flex justify-end'>
              <button
                type='button'
                onClick={onClose}
                className='rounded-lg p-2 text-medium hover:bg-nav'
                aria-label='Close'
              >
                <Icon>close</Icon>
              </button>
            </div>

            <CardContent>
              <h2 className='mb-4 text-xl font-semibold'>
                Two-Factor Authentication
              </h2>
              <p className='text-medium'>
                Enter the 6-digit code from your authenticator app to continue.
              </p>
            </CardContent>

            <Label className='mt-2'>Authenticator App</Label>
            <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
              <Icon>shield</Icon>
              <span>Time-based One-Time Password (TOTP)</span>
            </div>

            <TextField
              id='totp-popup'
              label='Authenticator Code'
              name='totp_code'
              type='text'
              className='mt-4'
              aria-invalid={Boolean(fetcher.data?.error) || undefined}
              aria-describedby={fetcher.data?.error ? 'totp-error' : undefined}
              required
              autoFocus
              errorMessage={fetcher.data?.error}
              disabled={isSubmitting}
            />

            <div className='mt-4 flex gap-2'>
              <Button
                type='button'
                onClick={onClose}
                className='flex-1'
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button
                type='submit'
                className='flex-1'
                disabled={isSubmitting}
                aria-label={isSubmitting ? 'Verifying...' : 'Verify'}
              >
                {isSubmitting ? 'Verifying...' : 'Verify'}
              </Button>
            </div>
          </Card>
        </fetcher.Form>
      </div>
    </div>
  )
}
