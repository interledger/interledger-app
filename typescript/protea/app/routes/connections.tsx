import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { Card, Icon, Layouts, Router, Snackbar } from '~/components'
import { grpcClient } from '~/lib/proto.server'
import { getSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  let keys = await grpcClient
    .listPublicKeys(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((resp) => resp.response.keys)

  let snackbar = await getSnackbar(request)

  return json({ keys, snackbar })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connections'
  }
}

export default function Page() {
  const { keys, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    snackbar.show ?? false
  )

  useEffect(() => {
    setShowSnackbar(snackbar.show ?? false)
  }, [snackbar])

  return (
    <>
      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Connections</h1>
        <p className='mt-6 text-medium'>Add and manage your connections.</p>
      </Card>

      {keys.length > 0 && (
        <>
          <br />
          <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <h1 className='font-display text-2xl font-medium'>Public keys</h1>
            {keys.map((key) => (
              <Card
                key={key.id}
                className='mt-6 bg-slate-100 col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'
              >
                <Router
                  to={route('/connections/:connectionId', {
                    connectionId: key.id
                  })}
                >
                  <div className='flex justify-between space-x-4'>
                    <div className='flex flex-col'>
                      <p className='text-sm text-medium space-y-1'>
                        {key.applicationName}
                      </p>
                      <p className='mt-1 text-xs'>Added {key.createdAt}</p>
                      <p className='text-xs text-purple-500'>
                        Last used {key.lastUsedAt}
                      </p>
                    </div>
                    <div className='flex content-start justify-between rounded-full bg-container text-medium'>
                      <Icon>navigate_next</Icon>
                    </div>
                  </div>
                </Router>
              </Card>
            ))}
          </Card>
        </>
      )}

      <br />

      <Card className='col-span-full space-y-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='flex items-start space-x-4'>
          <div className='flex items-center justify-between rounded-full bg-container p-5 text-medium'>
            <Icon>key</Icon>
          </div>
          <div className='flex flex-col space-y-4'>
            <p className='text-sm text-medium'>
              Add a public key for access and integration to your wallet.
            </p>
            <Router
              className='text-sm font-medium text-primary'
              to={route('/connections/add-a-public-key')}
            >
              Add a public key
            </Router>
          </div>
        </div>
      </Card>

      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        onClose={() => setShowSnackbar(false)}
      />
    </>
  )
}
