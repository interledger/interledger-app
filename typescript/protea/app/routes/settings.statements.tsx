import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  const cookie = request.headers.get('cookie')
  let res = await grpcClient
    .getStatements(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(res)) {
    throw res
  }

  const statements = res.response.statements

  return json({ statements })
}

export default function Page() {
  const { statements } = useLoaderData<typeof loader>()

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
          Statements
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {statements.length == 0 && (
          <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <Icon>tips_and_updates</Icon>
            <span className='font-sans text-sm font-normal'>
              You don't have any statements.
            </span>
          </div>
        )}
        {statements.length > 0 &&
          statements.map((statement) => (
            <div
              key={statement.id}
              className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
            >
              <div className='flex items-center space-x-3 text-medium'>
                {'account_balance' && <Icon>note</Icon>}
                <div className='flex flex-col'>
                  <span className='font-sans text-base font-normal'>
                    {statement.period}
                  </span>
                  <span className='font-sans text-xs font-normal text-weak'>
                    {statement.accountId}
                  </span>
                </div>
              </div>
              <Icon>file_download</Icon>
            </div>
          ))}
      </div>
    </div>
  )
}
