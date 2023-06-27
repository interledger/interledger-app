import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import {
  AnimatedSchedule,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
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
        {kycStatus == KycStatus.Unknown && (
          <Card>
            <CardHeader>
              <CardTitle>Wallet</CardTitle>
              <Chip color={ChipColor.orange}>Reserved</Chip>
            </CardHeader>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance_wallet</Icon>
                </CardIcon>
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
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.InProgress && (
          <Card>
            <CardHeader>
              <CardTitle>Activation</CardTitle>
              <Chip color={ChipColor.orange}>Pending</Chip>
            </CardHeader>
            <CardContent>
              <p className='text-sm text-medium'>
                Just a moment, we are verifying your details.
              </p>
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.Suspended && (
          <Card>
            <CardHeader>
              <CardTitle>Activation error</CardTitle>
              <Chip color={ChipColor.red}>Error</Chip>
            </CardHeader>
            <CardContent>
              <p className='text-sm text-medium'>
                We could not verify your identity. Please contact support.
              </p>
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.Verified && (
          <Card>
            <CardHeader>
              <CardTitle>Wallet</CardTitle>

              <Chip color={ChipColor.orange}>Pending</Chip>
            </CardHeader>
            <CardContent>
              <CardButton
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
              </CardButton>
            </CardContent>
          </Card>
        )}
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
              <CardLink
                key={transaction.id}
                to={`/transaction/${transaction.type}/${transaction.id}`}
                className='justify-between'
              >
                {(index == 0 ||
                  transaction.date != transactions[index - 1].date) && (
                  <span className='mt-6 text-xs text-medium'>
                    {transaction.date}
                  </span>
                )}
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
              </CardLink>
            ))}
          </Card>
        )}
      </GridColumn>
      <GridColumn className='hidden lg:col-span-6 lg:flex'>
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
                  to={route('/connect/card')}
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
