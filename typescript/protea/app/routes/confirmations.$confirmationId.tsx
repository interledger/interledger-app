import { useNavigate, useParams } from '@remix-run/react'
import { useEffect, useMemo } from 'react'
import { route, type RouteParams } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Icon,
  Layouts,
  OutlineButton
} from '~/components'
import { Label } from '~/components/Label'
import { usePendingConfirmationActions } from '~/lib/usePendingConfirmationActions'
import {
  usePendingConfirmations,
  type StorablePendingConfirmation
} from '~/lib/usePendingConfirmations'

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/confirmations'),
      title: 'Confirmation'
    },
    isNested: true
  }
}

export default function Page() {
  const navigate = useNavigate()
  const { pendingConfirmations, hasFetched, removeConfirmation } =
    usePendingConfirmations()
  const { confirmationId } =
    useParams<RouteParams['/confirmations/:confirmationId']>()

  const confirmation = useMemo(
    () =>
      confirmationId
        ? pendingConfirmations.find((c) => c.transactionId === confirmationId)
        : undefined,
    [pendingConfirmations, confirmationId]
  )

  useEffect(() => {
    if (hasFetched && (!confirmation || !confirmation.transactionId)) {
      navigate('/confirmations')
    }
  }, [hasFetched, confirmation, confirmationId, navigate])

  if (!confirmation) return null

  return (
    <ConfirmationView
      confirmation={confirmation}
      removeConfirmation={removeConfirmation}
    />
  )
}

function ConfirmationView({
  confirmation,
  removeConfirmation
}: {
  confirmation: StorablePendingConfirmation
  removeConfirmation: (id: string) => void
}) {
  const { decide, actionStatus } = usePendingConfirmationActions(confirmation)

  useEffect(() => {
    if (actionStatus === 'success') {
      removeConfirmation(confirmation.transactionId)
    }
  }, [actionStatus, confirmation.transactionId, removeConfirmation])

  return (
    <>
      <Card>
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              {confirmation.purchaseAmount} {confirmation.purchaseCurrency}
            </h2>
            <div className='flex flex-col items-end space-y-1'>
              <span className='text-sm font-medium text-medium'>
                {confirmation.formattedDate}
              </span>
              <span className='text-xs text-weak'>
                {confirmation.formattedTime}
              </span>
            </div>
          </div>
        </CardContent>
        <Label className='mt-2'>Merchant</Label>
        <div className='my-1 flex space-x-2 rounded-xl bg-nav p-3'>
          <div className='flex w-full items-center justify-between text-medium'>
            <div className='flex space-x-2'>
              <Icon>store</Icon>
              <span>{confirmation.merchantName}</span>
            </div>
          </div>
        </div>
      </Card>
      <Button
        disabled={actionStatus === 'loading'}
        onClick={() => decide('true')}
      >
        <span className='mx-auto font-medium'>Approve</span>
      </Button>
      <OutlineButton
        disabled={actionStatus === 'loading'}
        onClick={() => decide('false')}
      >
        <span className='mx-auto font-medium'>Decline</span>
      </OutlineButton>
    </>
  )
}
