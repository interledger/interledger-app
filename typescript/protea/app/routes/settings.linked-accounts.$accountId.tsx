import type { LoaderArgs, MetaFunction } from '@remix-run/node'
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
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  const cookie = String(request.headers.get('cookie'))

  const response = await grpcClient
    .getLinkedAccounts(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const linkedAccounts = response.response.linkedAccounts
    // .filter((account) => account.type != 'wallet')
    .map((linkedAccount) => ({
      id: linkedAccount.id,
      name:
        linkedAccount.type == 'sendCard'
          ? `Card ending ${linkedAccount.mask}`
          : 'Bank account',
      type: linkedAccount.type == 'sendCard' ? 'send' : 'receive',
      icon: linkedAccount.type == 'sendCard' ? 'credit_card' : 'account_balance'
    }))

  return json({
    linkedAccounts: linkedAccounts,
    canTopUp: linkedAccounts.filter(({ type }) => type == 'send').length > 0,
    hasReceive:
      linkedAccounts.filter(({ type }) => type == 'receive').length > 0
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Linked accounts'
  }
}

export default function Page() {
  const { linkedAccounts, canTopUp, hasReceive } =
    useLoaderData<typeof loader>()

  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Linked accounts</h1>
        {linkedAccounts.length == 0 && (
          <>
            <p className='mt-6'>
              You currently do not have any linked accounts.
            </p>

            <div className='mt-6 flex items-center space-x-3 rounded-xl bg-container-secondary p-4 text-medium'>
              <Icon>tips_and_updates</Icon>
              <p className='text-sm'>
                Enable sending or receiving in order to transact.
              </p>
            </div>
          </>
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
                {method.type == 'send' && (
                  <Chip color={ChipColor.orange}>send</Chip>
                )}
                {method.type == 'receive' && (
                  <Chip color={ChipColor.purple}>receive</Chip>
                )}
                <Icon>navigate_next</Icon>
              </div>
            </Router>
          ))}
        {linkedAccounts.length > 0 && (
          // TODO: Create route for choosing what type of linked account to add
          <Router
            className='mt-6 text-sm font-medium text-primary'
            to={route('/linked-account/:type', { type: 'new' })}
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
              <h1 className='font-medium text-medium'>Send money</h1>
              <p className='text-sm text-medium'>
                Easily send money from a debit card.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/linked-account/:type', { type: 'card' })}
              >
                Enable sending
              </Router>
            </div>
          </div>
        </div>
      )}
      {!hasReceive && (
        <div className='col-span-full flex flex-col space-y-6 rounded-2xl bg-page p-4 pb-6 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-start space-x-4'>
            <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
              <Icon>account_balance</Icon>
            </div>
            <div className='flex flex-col space-y-2'>
              <h1 className='font-medium text-medium'>Receive money</h1>
              <p className='text-sm text-medium'>
                Receive money into your bank account securely.
              </p>
              <Router
                className='text-sm font-medium text-primary'
                to={route('/linked-account/:type', { type: 'bank' })}
              >
                Enable receiving
              </Router>
            </div>
          </div>
        </div>
      )}
    </WalletGrid>
  )
}
