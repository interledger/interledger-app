import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardLink, Icon, Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  let response = await grpc.listRafikiGrants(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return json({ grants: response.grants })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
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
      <Card>
        {grants.length > 0 &&
          grants.map((grant) => (
            <CardLink
              to={route('/settings/grants/:grantId', {
                grantId: grant.id
              })}
              className='flex items-center justify-between'
              key={grant.id}
            >
              <div key={grant.id} className='flex-col'>
                <p className='font-medium text-medium'>{grant.client}</p>
                <p className='mt-2 text-sm text-medium'>
                  Added {grant.finalizationReason}
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
    </>
  )
}
