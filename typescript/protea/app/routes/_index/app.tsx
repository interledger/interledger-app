import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
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
  FynbosIcon,
  GridColumn,
  Icon,
  Router,
  Snackbar,
  TwitterIcon,
  WalletGrid
} from '~/components'
import { Label } from '~/components/Label'
import { usePusher } from '~/lib/usePusher'
import type { loader } from './route'
import { KycStatus } from './route'

export function AppPage() {
  const { walletInfo, snackbar, transactions, kycStatus, pusherArgs } =
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
        {(kycStatus == KycStatus.InProgress ||
          kycStatus == KycStatus.InReview) && (
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
            </CardHeader>
            <CardContent>
              <p>
                Share your wallet address to get paid, or click the pay button
                to transact.
              </p>
            </CardContent>
            <Label className='mt-2'>Wallet address</Label>
            <CardButton
              noHover
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
                  navigator.clipboard.writeText(walletInfo.url).then(
                    () => {
                      setSnackbar({
                        message: 'Wallet address copied to clipboard.',
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
              className='items-center justify-between'
            >
              <span className='font-medium text-medium'>
                {walletInfo.formattedURL}
              </span>
              <Icon className='text-medium'>content_copy</Icon>
            </CardButton>
          </Card>
        )}
        {kycStatus == KycStatus.Verified && (
          <>
            <div className='contents lg:hidden'>
              <CTACards />
            </div>
            <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <CardHeader>
                <CardTitle>Latest transactions</CardTitle>
                <Router className='flex max-h-fit' to={route('/transactions')}>
                  <Icon className='text-medium'>read_more</Icon>
                </Router>
              </CardHeader>
              {transactions.length == 0 && (
                <CardContent>
                  <div className='mt-4 flex flex-col space-y-4'>
                    <span className='text-sm text-medium'>
                      Your payment activity will appear here once you start
                      transacting.
                    </span>
                    <Router
                      to={route('/pay')}
                      className='text-sm font-medium text-primary'
                    >
                      Send or receive payments now
                    </Router>
                  </div>
                </CardContent>
              )}
              {transactions.map((transaction, index) => (
                <CardLink
                  key={transaction.id}
                  to={route('/transactions/:transactionId', {
                    transactionId: transaction.id
                  })}
                  className='justify-between'
                >
                  <div className='flex space-x-1'>
                    {transaction.state == 'Pending' && <AnimatedSchedule />}
                    {transaction.state != 'Pending' &&
                      transaction.destinationIdentityType == 'wallet' && (
                        <FynbosIcon />
                      )}
                    {transaction.state != 'Pending' &&
                      transaction.destinationIdentityType == 'twitter' && (
                        <TwitterIcon />
                      )}
                    <div className='flex flex-col space-y-1'>
                      <span className='text-medium'>{transaction.title}</span>
                      <span className='text-xs text-weak'>
                        {transaction.formattedDate} -{' '}
                        {transaction.formattedTime}
                      </span>
                    </div>
                  </div>
                  <div className='flex items-center space-x-2'>
                    <span
                      className={clsx(
                        'font-medium',
                        transaction.type.includes('outgoing')
                          ? 'text-error'
                          : 'text-medium'
                      )}
                    >
                      {transaction.type.includes('outgoing') && '- '}
                      {transaction.formattedAmount}
                    </span>
                    <Icon>navigate_next</Icon>
                  </div>
                </CardLink>
              ))}
            </Card>
          </>
        )}
      </GridColumn>
      <GridColumn className='hidden lg:col-span-6 lg:flex'>
        {kycStatus == KycStatus.Verified && <CTACards />}
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

function CTACards() {
  const { features, walletInfo } = useLoaderData<typeof loader>()

  return (
    <>
      {features.banksEnabled &&
        features.cardsEnabled &&
        !walletInfo.hasCard &&
        !walletInfo.hasBank && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <Icon>account_balance</Icon>
                </div>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Connect bank accounts and cards to easily send and receive
                    payments.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route('/accounts')}
                  >
                    Connect a bank or card
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      {features.cardsEnabled &&
        !walletInfo.hasCard &&
        (walletInfo.hasBank || !features.banksEnabled) && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <Icon>credit_card</Icon>
                </div>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Connect cards to easily send and receive payments.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route('/connect/card')}
                  >
                    Connect a card
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      {features.banksEnabled &&
        !walletInfo.hasBank &&
        (walletInfo.hasCard || !features.cardsEnabled) && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                  <Icon>account_balance</Icon>
                </div>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Connect bank accounts to easily send and receive payments.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route('/connect/bank')}
                  >
                    Connect a bank account
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      {features.twitterEnabled && !walletInfo.hasIdentities && (
        <Card>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                <TwitterIcon className='text-medium' />
              </div>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>
                  Connect a Twitter identity to transact with your audience.
                </p>
                <Router
                  className='text-sm font-medium text-primary'
                  to={route('/connect/twitter')}
                >
                  Connect Twitter identity
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}
