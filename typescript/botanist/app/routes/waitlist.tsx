import type { LoaderFunctionArgs, ActionFunctionArgs } from 'react-router'

import { Icon, Grid } from '~/components'
import { data } from 'react-router'
import { useLoaderData, useFetcher } from 'react-router'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const signups = await grpcClient.listWaitlistSignups(
    {},
    {
      meta: {
        cookies: String(request.headers.get('cookie')) || ''
      }
    }
  )

  return data({
    signups: signups.response.signups
  })
}

export default function Page() {
  const allowSignup = useFetcher()
  const { signups } = useLoaderData<typeof loader>()
  const onClick = async (id: string, canSignup: boolean) => {
    if (!canSignup) {
      if (allowSignup.state === 'idle') {
        allowSignup.submit(
          { id: id },
          {
            method: 'post',
            action: '/waitlist'
          }
        )
      }
    } else {
      await navigator.clipboard.writeText(
        `https://interledger.app/signup?waitlistSignupId=${id}`
      )
    }
  }

  return (
    <Grid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div className='sm:flex sm:items-center'>
          <div className='sm:flex-auto'>
            <h1 className='text-xl font-semibold text-gray-900'>Waitlist</h1>
            <p className='mt-2 text-sm text-gray-700'>
              All the users currently on the waitlist
            </p>
          </div>
        </div>
        <div className='mt-8 flex flex-col'>
          <div className='-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8'>
            <div className='inline-block min-w-full py-2 align-middle md:px-6 lg:px-8'>
              <div className='overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg'>
                <table className='min-w-full divide-y divide-gray-300'>
                  <thead className='bg-gray-50'>
                    <tr>
                      <th
                        scope='col'
                        className='py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6'
                      >
                        Name
                      </th>
                      <th
                        scope='col'
                        className='px-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                      >
                        Email
                      </th>
                      <th
                        scope='col'
                        className='px-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                      >
                        Country
                      </th>
                      <th
                        scope='col'
                        className='px-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                      >
                        Beta opt in
                      </th>
                      <th
                        scope='col'
                        className='px-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                      >
                        Mug ID
                      </th>
                      <th
                        scope='col'
                        className='px-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                      >
                        Can sign up
                      </th>
                    </tr>
                  </thead>
                  <tbody className='divide-y divide-gray-200 bg-white'>
                    {signups.map((person) => (
                      <tr key={person.id}>
                        <td className='py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6'>
                          {person.name}
                        </td>
                        <td className='px-3 py-4 text-sm text-gray-500'>
                          {person.email}
                        </td>
                        <td className='px-3 py-4 text-sm text-gray-500'>
                          {person.countryCode}
                        </td>
                        <td className='whitespace-nowrap px-3 py-4 text-sm text-gray-500'>
                          {person.betaOptIn ? (
                            <Icon className='text-green-500'>check</Icon>
                          ) : (
                            <Icon className='text-orange-500'>close</Icon>
                          )}
                        </td>
                        <td className='whitespace-nowrap px-3 py-4 text-sm text-gray-500'>
                          {person.mugId}
                        </td>
                        <td
                          onClick={() => onClick(person.id, person.canSignup)}
                          className='whitespace-nowrap px-3 py-4 text-sm text-gray-500'
                        >
                          {person.canSignup ? (
                            <>
                              <Icon className='text-green-500'>check</Icon>
                              <Icon className='cursor-pointer'>
                                content_copy
                              </Icon>
                            </>
                          ) : (
                            <span className='flex cursor-pointer items-center space-x-2 font-medium text-primary'>
                              <Icon>approval_delegation</Icon>
                              <span>Approve</span>
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Grid>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const id = form.get('id') as string

  let response = await grpcClient
    .allowWaitlistSignup(
      {
        id
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw data({}, httpMapping(response.code))
  }

  return data({ success: true })
}
