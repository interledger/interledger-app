import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Fab,
  GridColumn,
  Icon,
  Layouts,
  Router,
  Snackbar,
  WalletGrid
} from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus, getLinkedAccounts } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const { bankAccounts, cardAccounts } = await getLinkedAccounts(request)
  const kycStatus = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    snackbar,
    kycStatus: kycStatus.kycStatus,
    bankAccounts,
    cardAccounts,
    hasCard: cardAccounts.length > 0,
    hasBank: bankAccounts.length > 0
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Accounts',
      actions: [{ type: 'shapes' }]
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Accounts'
  }
}

export default function Page() {
  const { snackbar, bankAccounts, cardAccounts, hasCard, hasBank, kycStatus } =
    useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'accounts'}
        className='col-span-full lg:col-span-6'
      >
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
                    prefetch='render'
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
        {cardAccounts && hasCard && (
          <Card>
            <CardHeader>
              <CardTitle>Connected cards</CardTitle>
            </CardHeader>
            {cardAccounts.map((method) => (
              <CardLink
                key={method.id}
                to={route('/accounts/:accountId', {
                  accountId: method.id
                })}
                className='items-center justify-between'
              >
                <div className='flex space-x-3'>
                  <Icon>credit_card</Icon>
                  <span>{method.name}</span>
                </div>
                <div className='flex items-center space-x-2'>
                  {method.canSend && <Chip color={ChipColor.green}>Send</Chip>}
                  {method.canReceive && (
                    <Chip color={ChipColor.purple}>Receive</Chip>
                  )}
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
            <CardContent>
              <Router
                className='mt-4 text-sm font-medium text-primary'
                to={route('/connect/card')}
              >
                Connect another card
              </Router>
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.Verified && !hasCard && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>credit_card</Icon>
                </CardIcon>
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
        {bankAccounts && hasBank && (
          <Card>
            <CardHeader>
              <CardTitle>Connected bank accounts</CardTitle>
            </CardHeader>
            {bankAccounts.map((method) => (
              <CardLink
                key={method.id}
                to={route('/accounts/:accountId', {
                  accountId: method.id
                })}
                className='items-center justify-between'
              >
                <div className='flex space-x-3'>
                  <Icon>account_balance</Icon>
                  <span>
                    {method.name}{' '}
                    {method.nickname && '(' + method.nickname + ')'}
                  </span>
                </div>
                <div className='flex items-center space-x-2'>
                  <Chip color={ChipColor.green}>Send</Chip>
                  <Chip color={ChipColor.purple}>Receive</Chip>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
            <CardContent>
              <Router
                className='mt-4 text-sm font-medium text-primary'
                to={route('/connect/bank')}
              >
                Connect another bank account
              </Router>
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.Verified && !hasBank && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance</Icon>
                </CardIcon>
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
      </GridColumn>
      <GridColumn className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet />
      </GridColumn>
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setSnackbar(false)}
      />
    </WalletGrid>
  )
}
