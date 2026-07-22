import type { PlainMessage } from '@bufbuild/protobuf'
import { data, href, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Alert, AlertBody, Card, CardLink, Icon, Layouts } from '~/components'
import type { RafikiGrant } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/settings.grants'

export async function loader({ request }: Route.LoaderArgs) {
  const response = await grpc.listRafikiGrants(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return data({ grants: response.grants })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Grants'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Grants'
  }
])

export default function Page() {
  const { grants } = useLoaderData()

  return (
    <>
      {grants.length == 0 && (
        <Alert>
          <Icon>error</Icon>
          <AlertBody>You don't have any grants yet.</AlertBody>
        </Alert>
      )}

      {grants.length > 0 && (
        <Card>
          {grants.map((grant: PlainMessage<RafikiGrant>) => (
            <CardLink
              to={href('/settings/grants/:grantId', {
                grantId: grant.id
              })}
              className='flex items-center justify-between'
              key={grant.id}
            >
              <div key={grant.id} className='flex-col'>
                <p className='font-medium text-medium'>{grant.client}</p>
                <p className='mt-2 text-sm text-medium'>
                  Created {grant.createdAt}
                </p>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          ))}
        </Card>
      )}
    </>
  )
}
