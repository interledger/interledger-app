import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import {
  Chip,
  ChipColor,
  Icon,
  Layouts,
  Router,
  WalletGrid
} from '~/components'
import { getLinkedAccounts } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  return json(await getLinkedAccounts(request))
}

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { linkedAccounts, canTopUp, canWithdraw } =
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
            // TODO: Create route for individual linked account display
            <Router
              to={route('/settings/linked-accounts/:accountId', {
                accountId: method.id
              })}
              key={method.id}
              className='mt-4 flex w-full items-center justify-between  space-x-3 rounded-xl bg-container p-3 first-of-type:mt-6 hover:bg-container-hover'
            >
              <div className='flex items-center space-x-3 text-medium'>
                {method.icon && <Icon>{method.icon}</Icon>}
                <div className='flex flex-col text-sm'>
                  <span>{method.name}</span>
                </div>
              </div>
              <div className='flex items-center space-x-3 '>
                {method.type == 'card' && (
                  <Chip color={ChipColor.orange}>Send</Chip>
                )}
                {method.type == 'bank' && (
                  <Chip color={ChipColor.purple}>Withdraw</Chip>
                )}
                <Icon>navigate_next</Icon>
              </div>
            </Router>
          ))}
        {linkedAccounts.length > 0 && (
          // TODO: Create route for choosing what type of linked account to add
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/linked-account/:type/widget', { type: 'new' })}
          >
            Link another account
          </Router>
        )}
      </div>
      {!canTopUp && (
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
                to={route('/linked-account/:type', { type: 'card' })}
              >
                Add a debit card
              </Router>
            </div>
          </div>
        </div>
      )}
      {!canWithdraw && (
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
                to={route('/linked-account/:type', { type: 'bank' })}
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
