import { useFetcher } from '@remix-run/react'
import { useCallback, useEffect } from 'react'
import { route } from 'routes-gen'
import { useActionExecute } from '~/components/CardView/useActionExecute'
import type { PaymentConfirmationResponse } from '~/routes/api_.paymentConfirmation'
import { type StorablePendingConfirmation } from './usePendingConfirmations'
import { useScaffoldStore } from './useScaffoldStore'
import { useTotpChallenge } from './useTotpChallenge'

export const usePendingConfirmationActions = (
  confirmation: StorablePendingConfirmation
) => {
  const fetcher = useFetcher<PaymentConfirmationResponse>()
  const {
    errorStatus,
    successStatus,
    actionStatus,
    resetStatus,
    setActionStatus
  } = useActionExecute()
  const { withTotpChallenge } = useTotpChallenge()
  const pushSnackbar = useScaffoldStore((state) => state.pushSnackbar)

  useEffect(() => {
    resetStatus()
  }, [confirmation.transactionId, resetStatus])

  const decide = useCallback(
    (confirmed: 'true' | 'false') => {
      withTotpChallenge(() => {
        setActionStatus('loading')

        const formData = new FormData()
        formData.append('transactionId', confirmation.transactionId)
        formData.append('confirmed', confirmed)
        fetcher.submit(formData, {
          method: 'post',
          action: route('/api/paymentConfirmation')
        })
      })
    },
    [confirmation.transactionId, fetcher, setActionStatus, withTotpChallenge]
  )

  useEffect(() => {
    if (fetcher.data?.errors) {
      errorStatus()
      pushSnackbar({
        id: crypto.randomUUID(),
        message: 'Could not finish payment confirmation. Please try again'
      })
      return
    }

    if (fetcher.data?.success) {
      successStatus()
      pushSnackbar({
        id: crypto.randomUUID(),
        message: 'Payment ' + fetcher.data?.result + '.'
      })
      return
    }

    return
  }, [pushSnackbar, errorStatus, fetcher.data, setActionStatus, successStatus])

  return {
    decide,
    actionStatus
  }
}
