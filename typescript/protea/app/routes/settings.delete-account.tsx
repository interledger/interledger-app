import { Code } from '@bufbuild/connect'
import { href, redirect, useFetcher, useNavigate } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  Layouts,
  TextButton
} from '~/components'
import { getFeatures } from '~/data/wallet.server'
import { AccountDeletionRequestStatus } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useTotpChallenge } from '~/lib/useTotpChallenge'
import type { Route } from './+types/settings.delete-account'

export async function loader({ request }: Route.LoaderArgs) {
  const features = await getFeatures(request)
  if (!features.deleteAccountEnabled) {
    throw redirect(href('/settings'))
  }
  const statusResponse = await grpc.getAccountDeletionStatus(request, {})
  if (isConnectError(statusResponse)) {
    throw await redirectWithSnackbar(request, href('/settings'), {
      message:
        'Could not verify account deletion status. Please try again shortly.',
      icon: 'warning'
    })
  }
  if (statusResponse.status !== AccountDeletionRequestStatus.UNSPECIFIED) {
    throw await redirectWithSnackbar(request, href('/settings'), {
      message:
        'You already have an account deletion request on file. Check Settings for status.',
      icon: 'close'
    })
  }
  return null
}

export type AccountDeletionActionResponse = { error?: string }

export async function action({ request }: Route.ActionArgs) {
  const features = await getFeatures(request)
  if (!features.deleteAccountEnabled) {
    throw redirect(href('/settings'))
  }
  const response = await grpc.requestAccountDeletion(request, {})
  if (isConnectError(response)) {
    if (response.code === Code.AlreadyExists) {
      return { error: 'Account deletion is already pending.' }
    }
    if (response.code === Code.FailedPrecondition) {
      return {
        error:
          'Two-factor authentication must be configured before deleting your account.'
      }
    }
    return {
      error:
        'Could not submit account deletion request. Please try again or contact support.'
    }
  }
  return redirectWithSnackbar(request, href('/settings'), {
    message:
      'Account deletion request submitted. Our team will be in touch shortly.',
    icon: 'check'
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Delete account'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [{ title: 'Delete account' }])

export default function Page() {
  const navigate = useNavigate()
  const fetcher = useFetcher<AccountDeletionActionResponse>()
  const { withTotpChallenge } = useTotpChallenge()
  const isSubmitting = fetcher.state !== 'idle'

  const handleDelete = () => {
    withTotpChallenge(() => {
      fetcher.submit(null, {
        method: 'post',
        action: href('/settings/delete-account')
      })
    })
  }

  return (
    <Card>
      <CardHeader>
        <h1 className='text-xl font-medium text-error'>Delete your account</h1>
      </CardHeader>
      <CardContent>
        <p className='text-medium'>
          Are you sure you want to delete your account? This action cannot be
          undone. Once submitted, our team will process your request manually.
        </p>
        <p className='mt-3 text-medium'>
          <strong>
            Please withdraw any remaining balance within the next 2–3 days.
          </strong>{' '}
          After that, funds may not be recoverable.
        </p>
        {fetcher.data?.error && (
          <p role='alert' className='mt-3 text-sm text-error'>
            {fetcher.data.error}
          </p>
        )}
        <div className='mt-6 flex justify-end space-x-4'>
          <TextButton
            type='button'
            className='!text-medium'
            onClick={() => navigate(href('/settings'))}
            disabled={isSubmitting}
          >
            Cancel
          </TextButton>
          <Button
            error
            type='button'
            onClick={handleDelete}
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Submitting…' : 'Delete account'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
