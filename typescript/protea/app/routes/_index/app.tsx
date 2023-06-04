import { useLoaderData } from '@remix-run/react'
import { Fragment, useState } from 'react'
import { route } from 'routes-gen'
import {
  AnimatedSchedule,
  Card,
  HomeShapes,
  Icon,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
import { usePusher } from '~/lib/usePusher'
import type { loader } from './route'
import { KycStatus } from './route'

export function AppPage() {
  const {
    firstName,
    paymentPointer,
    snackbar,
    transactions,
    kycStatus,
    canTopUp,
    nextStep,
    pusherArgs
  } = useLoaderData<typeof loader>()

  const [snackbarState, setSnackbar] = useState<any>(snackbar)
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    snackbar.show ?? false
  )

  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <WalletGrid>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='mt-2'>
          <HomeShapes />
        </div>
        <h1 className='mt-6 text-center font-display text-2xl font-medium'>
          Welcome {firstName}
        </h1>
        {kycStatus != KycStatus.Verified && (
          <p className='mb-2 mt-4 text-center'>
            Thank you for signing up to Fynbos.
          </p>
        )}
        {kycStatus == KycStatus.Verified && !canTopUp && (
          <p className='mb-2 mt-4 text-center'>
            Your payment pointer is activated. You are now able to send and
            receive payments.
          </p>
        )}
        {kycStatus == KycStatus.Verified && canTopUp && (
          <p className='mb-2 mt-4 text-center'>
            Send and receive payments with your payment pointer.
          </p>
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
                navigator.clipboard.writeText(paymentPointer.formatted).then(
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
            className='mt-4 flex flex items-center justify-between rounded-xl bg-nav p-4 hover:bg-nav-hover'
          >
            <span className='font-medium text-medium'>
              {paymentPointer.formatted}
            </span>
            <Icon className='text-medium'>content_copy</Icon>
          </button>
        )}
      </Card>

      {kycStatus == KycStatus.InProgress && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h2 className='font-display text-lg font-medium'>
            Activation pending
          </h2>
          <p className='mt-4'>Just a moment, we are verifying your details.</p>
        </Card>
      )}
      {/*{kycStatus == KycStatus.Retry && (*/}
      {/*  <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>*/}
      {/*    <h2 className='font-display text-lg font-medium'>*/}
      {/*      Activation failed*/}
      {/*    </h2>*/}
      {/*    <p className='mt-4'>*/}
      {/*      Some of the details you provided were not correct. Please fix them*/}
      {/*      and submit again.*/}
      {/*    </p>*/}
      {/*    <Router*/}
      {/*      className='mt-4 text-sm font-medium text-primary'*/}
      {/*      to={route('/personal-details')}*/}
      {/*    >*/}
      {/*      Fix personal details*/}
      {/*    </Router>*/}
      {/*  </Card>*/}
      {/*)}*/}
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
      {/*{kycStatus == KycStatus.ReviewPending && (*/}
      {/*  <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>*/}
      {/*    <h2 className='font-display text-lg font-medium'>*/}
      {/*      Reviewing activation*/}
      {/*    </h2>*/}
      {/*    <p className='mt-4'>*/}
      {/*      We need to manually review your verification details. We will notify*/}
      {/*      you when this process completes.*/}
      {/*    </p>*/}
      {/*  </Card>*/}
      {/*)}*/}

      {nextStep.show && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h1 className='font-display text-lg font-medium'>Next step</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <Icon>{nextStep.icon}</Icon>
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>{nextStep.title}</p>
              <Router
                className='text-sm font-medium text-primary'
                to={nextStep.action.to}
              >
                {nextStep.action.text}
              </Router>
            </div>
          </div>
        </Card>
      )}

      {kycStatus == KycStatus.Verified && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-center justify-between'>
            <h1 className='font-display text-lg font-medium'>
              Latest transactions
            </h1>
            <Router className='flex max-h-fit' to={route('/transactions')}>
              <Icon className='text-medium'>read_more</Icon>
            </Router>
          </div>
          {transactions.length == 0 && (
            <div className='mt-4 flex flex-col space-y-4'>
              <span className='text-sm text-medium'>
                Your payment activity will appear here once you start using your
                payment pointer.
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
