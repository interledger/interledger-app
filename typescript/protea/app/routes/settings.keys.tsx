import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export async function loader({ request }: LoaderArgs) {
  let response = await grpc.listConnections(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return json({ keys: response.keys })
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
  const { keys } = useLoaderData<typeof loader>()

  return (
    <>
      <Card>
        <CardContent>
          <p>
            Connect external applications to your wallet by uploading their
            keys.
          </p>
          {keys.length > 0 &&
            keys.map((k) => (
              // <CardLink
              //   to={route('/settings/keys/:keyId', {
              //     keyId: k.id
              //   })}
              //   className='flex justify-between'
              //   key={k.id}
              // >
              <div key={k.id} className='mt-4 flex-col rounded-xl bg-nav p-4'>
                <p className='font-medium text-medium'>{k.applicationName}</p>
                <p className='mt-2 text-sm text-medium'>Added {k.createdAt}</p>
                {/*TODO: implement last used*/}
                {/*<p className='mt-1 text-sm text-purple-500'>
                                      Last used {conn.lastUsedAt}
                                    </p>*/}
              </div>
              //   <Icon>navigate_next</Icon>
              // </CardLink>
            ))}
        </CardContent>
      </Card>

      {/*<ButtonRouter to={route('/settings/keys/add-public')}>*/}
      {/*  Add a public key*/}
      {/*</ButtonRouter>*/}
    </>
  )
}
