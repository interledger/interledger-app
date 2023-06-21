import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  ButtonRouter,
  Card,
  CardContent,
  CardLink,
  Icon,
  Layouts,
  Snackbar
} from '~/components'
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

  return json({ keys: connections, snackbar })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Keys'
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Keys'
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
      <Card>
        <CardContent>
          <p>
            Connect external applications to your wallet by uploading their
            keys.
          </p>
        </CardContent>
        {keys.length > 0 &&
          keys.map((k) => (
            <CardLink
              to={route('/settings/keys/:keyId', {
                keyId: k.id
              })}
              className='flex justify-between'
              key={k.id}
            >
              <div className='flex-col'>
                <p className='font-medium text-medium'>{k.applicationName}</p>
                <p className='mt-2 text-sm text-medium'>Added {k.createdAt}</p>
                {/*TODO: implement last used*/}
                {/*<p className='mt-1 text-sm text-purple-500'>
                                      Last used {conn.lastUsedAt}
                                    </p>*/}
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          ))}
      </Card>

      <ButtonRouter to={route('/settings/keys/add-public')}>
        Add a public key
      </ButtonRouter>

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
