import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Card, Icon, Layouts, Router, Snackbar } from '~/components'
import { getKycStatus, getLinkedAccounts } from '~/lib/wallet.server'
import { KycStatus } from '~/routes/index'
import { useState } from 'react'
import { getSnackbar } from '~/lib/snackbar.server'

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

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Linked accounts'
  }
}

export default function Page() {
  const { snackbar, linkedAccounts, canTopUp, canWithdraw, kycStatus } =
    useLoaderData<typeof loader>()

  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  return (
    <>
      <Card>
        <h1 className='font-display text-2xl font-medium'>Linked accounts</h1>
        {linkedAccounts.length == 0 && (
          <p className='mt-6'>You do not have any linked accounts.</p>
        )}

        {linkedAccounts &&
          linkedAccounts.length > 0 &&
          linkedAccounts.map((method) => (
            <Router
              key={method.id}
              to={route('/settings/linked-accounts/:accountId', {
                accountId: method.id
              })}
              className='mt-4 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
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
        {linkedAccounts.length > 0 && (
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/linked-account/:type', { type: 'new' })}
          >
            Link another account
          </Router>
        )}
      </Card>
      {kycStatus == KycStatus.Unknown && (
        <Card className='col-span-full space-y-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
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
        </Card>
      )}
      {kycStatus == KycStatus.Verified && !canTopUp && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
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
              {/*<Router*/}
              {/*  className='text-sm font-medium text-primary'*/}
              {/*  to={route('/linked-account/:type/widget', { type: 'card' })}*/}
              {/*>*/}
              {/*  Add a debit card*/}
              {/*</Router>*/}
            </div>
          </div>
        </Card>
      )}
      {kycStatus == KycStatus.Verified && !canWithdraw && (
        <Card className='col-span-full space-y-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
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
              {/*<Router*/}
              {/*  className='text-sm font-medium text-primary'*/}
              {/*  to={route('/linked-account/:type/widget', { type: 'bank' })}*/}
              {/*>*/}
              {/*  Add a bank account*/}
              {/*</Router>*/}
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
    </>
  )
}
