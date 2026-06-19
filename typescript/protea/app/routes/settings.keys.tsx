import type { PlainMessage } from '@bufbuild/protobuf'
import { data, href, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  ButtonRouter,
  Card,
  CardContent,
  CardLink,
  Icon,
  Layouts
} from '~/components'
import type { Connection } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/settings.keys'

export async function loader({ request }: Route.LoaderArgs) {
  const response = await grpc.listConnections(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return data({ keys: response.keys })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Keys'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Keys'
  }
])

export default function Page() {
  const { keys } = useLoaderData()

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
          keys.map((key: PlainMessage<Connection>) => (
            <CardLink
              to={href('/settings/keys/:keyId', {
                keyId: key.id
              })}
              className='flex items-center justify-between'
              key={key.id}
            >
              <div key={key.id} className='flex-col'>
                <p className='font-medium text-medium'>{key.applicationName}</p>
                <p className='mt-2 text-sm text-medium'>
                  Added {key.createdAt}
                </p>
                {/*TODO: implement last used*/}
                {/*<p className='mt-1 text-sm text-purple-500'>
                                      Last used {conn.lastUsedAt}
                                    </p>*/}
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          ))}
      </Card>

      <ButtonRouter to={href('/settings/keys/add-public')}>
        Add a public key
      </ButtonRouter>
    </>
  )
}
