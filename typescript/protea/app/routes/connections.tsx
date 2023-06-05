import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, Icon, Layouts, Router, Snackbar } from '~/components'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { getSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  let connections = await grpcClient
    .listConnections(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((resp) => resp.response.keys)
    .catch(StatusError)

  if (isGrpcError(connections)) {
    throw json({}, httpMapping(connections.code))
  }

  let snackbar = await getSnackbar(request)

  return json({ connections, snackbar })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Connections'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connections'
  }
}

export default function Page() {
  const { connections, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    snackbar.show ?? false
  )

  useEffect(() => {
    setShowSnackbar(snackbar.show ?? false)
  }, [snackbar])

  return (
    <>
      <Card>
        <p className='text-medium'>Add and manage your connections.</p>
      </Card>

      {connections.length > 0 && (
        <Card>
          <h1 className='font-display text-2xl font-medium'>Public keys</h1>
          {connections.map((conn) => (
            <Router
              to={route('/connections/:connectionId', {
                connectionId: conn.id
              })}
              className='mt-6 flex justify-between rounded-xl bg-nav p-3'
              key={conn.id}
            >
              <div className='flex-col'>
                <p className='font-medium text-medium'>
                  {conn.applicationName}
                </p>
                <p className='mt-2 text-sm text-medium'>
                  Added {conn.createdAt}
                </p>
                {/*TODO: implement last used*/}
                {/*<p className='mt-1 text-sm text-purple-500'>
                                      Last used {conn.lastUsedAt}
                                    </p>*/}
              </div>
              <Icon>navigate_next</Icon>
            </Router>
          ))}
        </Card>
      )}

      <Card>
        <div className='flex items-start space-x-4'>
          <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
            <Icon>key</Icon>
          </div>
          <div className='flex flex-col space-y-2'>
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
