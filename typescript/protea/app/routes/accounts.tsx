import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, Icon, Layouts, Router, Snackbar, WalletGrid } from '~/components'
import { getSnackbar } from '~/lib/snackbar.server'
import { getKycStatus, getLinkedAccounts } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/_index/route'

export async function loader({ request }: LoaderArgs) {
  const { linkedAccounts, canTopUp, canWithdraw } = await getLinkedAccounts(
    request
  )
  const kycStatus = await getKycStatus(request)

  const snackbar = await getSnackbar(request)

  return json({
    snackbar,
    linkedAccounts: linkedAccounts.filter(
      (account) => account.type != 'wallet'
    ),
    kycStatus: kycStatus.kycStatus,
    canTopUp,
    canWithdraw
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Accounts'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Accounts'
  }
}

export default function Page() {
  const { snackbar, linkedAccounts, canTopUp, canWithdraw, kycStatus } =
    useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <WalletGrid>
      {linkedAccounts && linkedAccounts.length > 0 && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          {linkedAccounts.map((method) => (
            <Router
              key={method.id}
              to={route('/accounts/:accountId', {
                accountId: method.id
              })}
              className='mt-4 flex items-center justify-between rounded-xl bg-nav p-3 text-medium hover:bg-nav-hover'
            >
              <div className='flex space-x-3'>
                {method.icon && <Icon>{method.icon}</Icon>}
                <span>
                  {method.name} {method.nickname && '(' + method.nickname + ')'}
                </span>
              </div>
              <Icon>navigate_next</Icon>
            </Router>
          ))}
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/link-account')}
          >
            Connect another account
          </Router>
        </Card>
      )}
      {linkedAccounts && linkedAccounts.length == 0 && (
        <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <p>You do not have any connected accounts.</p>
        </Card>
      )}
      {kycStatus == KycStatus.Unknown && (
        <Card className='col-span-full space-y-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h1 className='font-display text-lg font-medium'>Next step</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <Icon>attach_money</Icon>
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                Your wallet address is reserved, we just need a few more details
                to activate it.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/personal-details')}
              >
                Activate payment pointer
              </Router>
            </div>
          </div>
        </Card>
      )}
      {kycStatus == KycStatus.Verified && !canTopUp && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <Icon>credit_card</Icon>
            </div>
            <div className='flex flex-col space-y-2'>
              <h1 className='font-medium text-medium'>Add a debit card</h1>
              <p className='text-sm text-medium'>
                Add a debit card to easily send payments or top up your cash
                balance.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/link-account/card')}
              >
                Add a debit card
              </Router>
            </div>
          </div>
        </Card>
      )}
      {kycStatus == KycStatus.Verified && !canWithdraw && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
              <Icon>account_balance</Icon>
            </div>
            <div className='flex flex-col space-y-2'>
              <h1 className='font-medium text-medium'>Add a bank account</h1>
              <p className='text-sm text-medium'>
                Add a bank account to securely withdraw from your cash balance
                at any time.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/link-account/bank')}
              >
                Add a bank account
              </Router>
            </div>
          </div>
        </Card>
      )}
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
