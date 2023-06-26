import { useLoaderData } from '@remix-run/react'
import { Fragment, useState } from 'react'
import { route } from 'routes-gen'
import {
  AnimatedSchedule,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Chip,
  ChipColor,
  GridColumn,
  Icon,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
import { usePusher } from '~/lib/usePusher'
import type { loader } from './route'
import { KycStatus } from './route'

export function AppPage() {
  const { paymentPointer, snackbar, transactions, kycStatus, pusherArgs } =
    useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>(snackbar)
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    snackbar.show ?? false
  )

  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <WalletGrid>
      <GridColumn className='col-span-full lg:col-span-6'>
        <Card>
          <CardHeader>
            <CardTitle>Wallet</CardTitle>
            {kycStatus == KycStatus.InProgress && (
              <Chip color={ChipColor.orange}>Pending</Chip>
            )}
            {kycStatus == KycStatus.Unknown && (
              <Chip color={ChipColor.orange}>Reserved</Chip>
            )}
          </CardHeader>
          <CardContent>
            {kycStatus == KycStatus.InProgress && (
              <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                <h2 className='font-display text-lg font-medium'>
                  Activation pending
                </h2>
                <p className='mt-4'>
                  Just a moment, we are verifying your details.
                </p>
              </Card>
            )}
            {kycStatus == KycStatus.Suspended && (
              <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                <h2 className='font-display text-lg font-medium'>
                  Activation failed
                </h2>
                <p className='mt-4'>
                  We could not verify your identity, please contact support to
                  continue.
                </p>
                <Router
                  className='mt-4 text-sm font-medium text-primary'
                  to={route('/support')}
                >
                  Contact support
                </Router>
              </Card>
            )}
            {kycStatus == KycStatus.Unknown && (
              <div className='flex items-start space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <Icon>account_balance_wallet</Icon>
                </div>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Your wallet is reserved, we just need a few more details to
                    activate it.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route('/personal-details')}
                  >
                    Activate wallet
                  </Router>
                </div>
              </div>
            )}
            {kycStatus == KycStatus.Verified && (
              <button
                type='button'
                onClick={async () => {
                  if (typeof navigator.clipboard == 'undefined') {
                    setSnackbar({
                      message: "Couldn't copy to clipboard.",
                      icon: 'close',
                      show: true
                    })
                    setShowSnackbar(true)
                  } else
                    navigator.clipboard
                      .writeText(paymentPointer.formatted)
                      .then(
                        () => {
                          setSnackbar({
                            message: 'Payment pointer copied to clipboard.',
                            icon: 'close',
                            show: true
                          })
                          setShowSnackbar(true)
                        },
                        () => {
                          setSnackbar({
                            message: "Couldn't copy to clipboard.",
                            icon: 'close',
                            show: true
                          })
                          setShowSnackbar(true)
                        }
                      )
                }}
                className='mt-4 flex items-center justify-between rounded-xl bg-nav p-4 hover:bg-nav-hover'
              >
                <span className='font-medium text-medium'>
                  {paymentPointer.formatted}
                </span>
                <Icon className='text-medium'>content_copy</Icon>
              </button>
            )}
          </CardContent>
        </Card>
        {kycStatus == KycStatus.Verified && (
          <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <CardHeader>
              <CardTitle>Latest transactions</CardTitle>
              <Router className='flex max-h-fit' to={route('/transactions')}>
                <Icon className='text-medium'>read_more</Icon>
              </Router>
            </CardHeader>
            {transactions.length == 0 && (
              <div className='mt-4 flex flex-col space-y-4'>
                <span className='text-sm text-medium'>
                  Your payment activity will appear here once you start using
                  your payment pointer.
                </span>
                <Router
                  to={route('/pay')}
                  className='text-sm font-medium text-primary'
                >
                  Send or receive payments now
                </Router>
              </div>
            )}
            {transactions.map((transaction, index) => (
              <Fragment key={transaction.id}>
                {(index == 0 ||
                  transaction.date != transactions[index - 1].date) && (
                  <span className='mt-6 text-xs text-medium'>
                    {transaction.date}
                  </span>
                )}
                <Router
                  to={`/transaction/${transaction.type}/${transaction.id}`}
                  className='mt-2 flex w-full justify-between'
                >
                  <div className='flex space-x-1'>
                    {transaction.icon == 'schedule' && (
                      <div className='mt-0.5'>
                        <AnimatedSchedule />
                      </div>
                    )}
                    {transaction.icon != 'schedule' && (
                      <Icon className='mt-0.5 text-medium'>
                        {transaction.icon}
                      </Icon>
                    )}
                    <div className='flex flex-col space-y-2'>
                      <span className='text-medium'>{transaction.title}</span>
                      <span className='text-xs text-medium'>
                        {transaction.time}
                      </span>
                    </div>
                  </div>
                  <span className='font-medium'>{transaction.total}</span>
                </Router>
              </Fragment>
            ))}
          </Card>
        )}
      </GridColumn>
      <GridColumn className='col-span-full lg:col-span-6'>
        <Card>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                <Icon>credit_card</Icon>
              </div>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>Add card?</p>
                <Router
                  className='text-sm font-medium text-primary'
                  to={route('/link-account/card')}
                >
                  Something
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      </GridColumn>

      <Snackbar
        message={snackbarState.message}
        action={snackbarState.action}
        icon={snackbarState.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setShowSnackbar(false)}
      />
    </WalletGrid>
  )
}
