import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Layouts, Router, WalletGrid } from '~/components'
import { getKycStatus, getLinkedAccounts } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/index'

export async function loader({ request }: LoaderArgs) {
  const { linkedAccounts, canTopUp, canWithdraw } = await getLinkedAccounts(
    request
  )
  const kycStatus = await getKycStatus(request)
  return json({
    linkedAccounts: linkedAccounts.filter(
      (account) => account.type != 'wallet'
    ),
    kycStatus: kycStatus.kycStatus,
    canTopUp,
    canWithdraw
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { linkedAccounts, canTopUp, canWithdraw, kycStatus } =
    useLoaderData<typeof loader>()

  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Linked accounts</h1>
        {linkedAccounts.length == 0 && (
          <p className='mt-6'>You do not have any linked accounts.</p>
        )}

        {linkedAccounts &&
          linkedAccounts.length > 0 &&
          linkedAccounts.map((method) => (
            <div
              key={method.id}
              className='mt-4 flex w-full items-center justify-between  space-x-3 rounded-xl bg-container p-3 first-of-type:mt-6'
            >
              <div className='flex items-center space-x-3 text-medium'>
                {method.icon && <Icon>{method.icon}</Icon>}
                <div className='flex flex-col text-sm'>
                  <span>{method.name} (My nickname)</span>
                </div>
              </div>
            </div>
          ))}
        {linkedAccounts.length > 0 && (
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/linked-account/:type', { type: 'new' })}
          >
            Link another account
          </Router>
        )}
      </div>
      {kycStatus == KycStatus.Unknown && (
        <div className='col-span-full flex flex-col space-y-4 rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <h1 className='font-display text-lg font-medium'>Next step</h1>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
              <Icon>attach_money</Icon>
            </div>
            <div className='flex flex-col space-y-4'>
              <p className='text-sm text-medium'>
                Your payment pointer is reserved, we just need a few more
                details to activate it.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/personal-details')}
              >
                Activate payment pointer
              </Router>
            </div>
          </div>
        </div>
      )}
      {kycStatus == KycStatus.Verified && !canTopUp && (
        <div className='col-span-full flex flex-col space-y-6 rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
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
                to={route('/linked-account/:type/widget', { type: 'card' })}
              >
                Add a debit card
              </Router>
            </div>
          </div>
        </div>
      )}
      {kycStatus == KycStatus.Verified && !canWithdraw && (
        <div className='col-span-full flex flex-col space-y-6 rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
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
                to={route('/linked-account/:type/widget', { type: 'bank' })}
              >
                Add a bank account
              </Router>
            </div>
          </div>
        </div>
      )}
    </WalletGrid>
  )
}
