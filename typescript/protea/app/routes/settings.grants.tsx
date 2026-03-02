import type { LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data } from 'react-router';
import { useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Alert, AlertBody, Card, CardLink, Icon, Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  let response = await grpc.listRafikiGrants(request, {})

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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Grants'
  }
])

export default function Page() {
  const { grants } = useLoaderData<typeof loader>()

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
          {grants.map((grant) => (
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
