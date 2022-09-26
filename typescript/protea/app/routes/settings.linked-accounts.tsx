import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Router } from '~/components'
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

  const linkedAccounts = response.response.linkedAccounts.map((la) => ({
    id: la?.id,
    name: la?.name,
    description: la?.mask,
    icon: 'account_balance' // TODO: get actual icon from fundingsource subtype
  }))

  return json({
    linkedAccounts
  })
}

export default function Page() {
  const { linkedAccounts } = useLoaderData<typeof loader>()
  console.log(linkedAccounts)

  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-app p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <Link to={route('/settings')}>
          <div className='-ml-3 p-3 text-medium'>
            <Icon>arrow_back</Icon>
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Linked accounts
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {linkedAccounts && linkedAccounts.length == 0 && (
          <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <Icon>tips_and_updates</Icon>
            <span className='text-sm'>
              You need to add a linked account before you can deposit money.
            </span>
          </div>
        )}
        {linkedAccounts &&
          linkedAccounts.length > 0 &&
          linkedAccounts.map((method) => (
            <div
              key={method.id}
              className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
            >
              <div className='flex items-center space-x-3 text-medium'>
                {method.icon && <Icon>{method.icon}</Icon>}
                <div className='flex flex-col'>
                  <span>{method.name}</span>
                  <span className='text-xs text-weak'>
                    {method.description}
                  </span>
                </div>
              </div>
            </div>
          ))}
        <Router
          to={route('/')}
          className='col-span-full mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'
        >
          <span>Add linked account</span>
          <Icon>navigate_next</Icon>
        </Router>
      </div>
    </div>
  )
}
