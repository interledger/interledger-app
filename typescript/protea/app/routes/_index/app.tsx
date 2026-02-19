import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import { route } from 'routes-gen'
import {
  Alert,
  AlertBody,
  Card,
  CardContent,
  CardCopy,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  DiscordIcon,
  GridColumn,
  Icon,
  InterledgerIcon,
  MasterCardLogo,
  Router,
  SlackIcon,
  TwitterIcon,
  WalletGrid
} from '~/components'
import { Label } from '~/components/Label'
import { usePendingConfirmations } from '~/lib/cards/usePendingConfirmations'
import { usePusher } from '~/lib/usePusher'
import type { appLoader } from './route'
import { KycStatus } from './route'

export function AppPage() {
  const {
    walletInfo,
    features,
    transactions,
    kycStatus,
    pusherArgs,
    balances
  } = useLoaderData<typeof appLoader>()

  const { pendingConfirmations } = usePendingConfirmations()

  usePusher(pusherArgs, ['transaction', 'kyc'])

  return (
    <WalletGrid>
      <GridColumn className='col-span-full lg:col-span-6'>
        {!features.sendEnabled && kycStatus == KycStatus.Approved && (
          <Alert>
            <Icon>south_west</Icon>
            <AlertBody>Receive only account</AlertBody>
            <Router className='ml-auto text-primary' to={route('/pay')}>
              Find out why
            </Router>
          </Alert>
        )}
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
        {(kycStatus == KycStatus.Pending ||
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
        {kycStatus == KycStatus.Denied && (
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
        {kycStatus == KycStatus.Approved && (
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
            <CardCopy
              copyContent={walletInfo.url}
              shareData={{
                title: 'Wallet address',
                text: 'You can pay me using my wallet address.',
                url: walletInfo.url
              }}
              success='Wallet address copied to clipboard.'
              copyError="Couldn't copy to clipboard."
              shareError="Couldn't share wallet address."
            >
              {walletInfo.formattedURL}
            </CardCopy>
          </Card>
        )}
        {kycStatus == KycStatus.Approved && (
          <>
            <div className='contents lg:hidden'>
              <CTACards />
            </div>
            {balances.map((method) => (
              <Card key={method.linkedAccount}>
                <CardContent className='flex items-center justify-between'>
                  <div className='flex items-center space-x-3'>
                    <div className={`flag:${method.countryCode}`} />
                    <span className='text-lg font-medium capitalize text-strong'>
                      {method.currency}
                    </span>
                  </div>
                  <span className='text-2xl'>{method.formattedBalance}</span>
                </CardContent>
              </Card>
            ))}

            {/* CARDS */}
            {features.manageWalletCardsEnabled && (
              <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                <CardContent className='flex items-center justify-between'>
                  <div className='flex items-center space-x-3'>
                    <MasterCardLogo size='sm' />
                    <span className='text-lg font-medium capitalize text-strong'>
                      Cards
                    </span>
                  </div>
                  <Router className='flex max-h-fit' to={route('/cards')}>
                    <Icon className='text-medium'>read_more</Icon>
                  </Router>
                </CardContent>
              </Card>
            )}

            {/* PENDING CONFIRMATIONS */}
            {pendingConfirmations.length > 0 && (
              <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                <Router className='flex max-h-fit' to={route('/confirmations')}>
                  <CardHeader className='mb-2'>
                    <CardTitle className='flex'>
                      <Icon className='mr-2 mt-1 flex w-6 text-orange-500'>
                        priority_high
                      </Icon>{' '}
                      You have pending 3DS confirmations
                    </CardTitle>
                    <Icon className='text-medium'>read_more</Icon>
                  </CardHeader>
                </Router>
              </Card>
            )}

            {/* LATEST PAYMENTS */}
            <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <CardHeader>
                <CardTitle>Latest payments</CardTitle>
                <Router className='flex max-h-fit' to={route('/payments')}>
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
                  to={route('/payments/:paymentId', {
                    paymentId: transaction.id
                  })}
                  className='justify-between'
                >
                  <div className='flex space-x-1'>
                    {transaction.state == 'Pending' && <Icon>schedule</Icon>}
                    {transaction.state == 'Failed' && (
                      <Icon className='text-error'>exclamation</Icon>
                    )}
                    {transaction.state != 'Pending' &&
                      transaction.state != 'Failed' && (
                        <>
                          {transaction.destinationIdentityType == 'wallet' && (
                            <>
                              {transaction.type != 'withdrawal' &&
                                transaction.type != 'deposit' && (
                                  <InterledgerIcon />
                                )}
                              {transaction.type == 'withdrawal' && (
                                <Icon>south_west</Icon>
                              )}
                              {transaction.type == 'deposit' && (
                                <Icon>north_east</Icon>
                              )}
                            </>
                          )}
                          {transaction.destinationIdentityType == 'Twitter' && (
                            <TwitterIcon />
                          )}
                          {transaction.destinationIdentityType == 'Discord' && (
                            <DiscordIcon />
                          )}
                          {transaction.destinationIdentityType == 'Slack' && (
                            <SlackIcon />
                          )}
                          {transaction.type == 'card_transaction' && (
                            <Icon>credit_card</Icon>
                          )}
                        </>
                      )}
                    {transaction.state != 'Pending' &&
                      transaction.state != 'Failed' &&
                      transaction.destinationIdentityType == 'Unknown' && (
                        <Icon>account_circle</Icon>
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
                        transaction.type == 'sent' ||
                          (transaction.type == 'card_transaction' &&
                            transaction.state === 'Completed')
                          ? 'text-error'
                          : 'text-medium'
                      )}
                    >
                      {transaction.type == 'sent' && '- '}
                      {transaction.type == 'card_transaction' &&
                        transaction.state == 'Completed' &&
                        '- '}
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
        {kycStatus == KycStatus.Approved && <CTACards />}
      </GridColumn>
    </WalletGrid>
  )
}

function CTACards() {
  const { features, walletInfo } = useLoaderData<typeof appLoader>()

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
                    Connect bank accounts to easily add or withdraw from your
                    balance.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={route(
                      walletInfo.country === 'US'
                        ? '/connect/bank/us'
                        : '/connect/bank/za'
                    )}
                  >
                    Connect a bank account
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      {features.interacEnabled && (
        <Card>
          <CardContent>
            <div className='flex items-start space-x-4'>
              <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                <Icon>account_balance</Icon>
              </div>
              <div className='flex flex-col space-y-4'>
                <p className='text-sm text-medium'>
                  Connect an Interac account to easily withdraw from your
                  balance.
                </p>
                <Router
                  className='text-sm font-medium text-primary'
                  to={route('/connect/interac')}
                >
                  Connect an Interac account
                </Router>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  )
}
