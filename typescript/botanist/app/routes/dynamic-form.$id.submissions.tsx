import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  isRouteErrorResponse,
  NavLink,
  Outlet,
  useLoaderData,
  useLocation,
  useRouteError
} from '@remix-run/react'
import { ListFormSubmissions } from '~/lib/wallet.server'
import type { FC } from 'react'
import { route } from 'routes-gen'
import clsx from 'clsx'
import { Error } from '~/components'

export async function loader({ request, params }: LoaderArgs) {
  const submissions = await ListFormSubmissions(request, params.id as string)

  return json({
    submissions
  })
}

interface Submission {
  id: string
  formId: string
  date: string
}

const ListItem: FC<Submission> = ({ id, formId, date }) => {
  return (
    <NavLink
      preventScrollReset={true}
      prefetch='none'
      className='flex'
      to={route('/dynamic-form/:id/submissions/:submissionId', {
        id: formId,
        submissionId: id
      })}
    >
      {({ isActive }) => (
        <li
          className={`flex w-full flex-col items-center space-y-2 rounded-lg p-3 hover:bg-slate-50 ${
            isActive ? 'bg-container-hover' : 'hover:bg-container'
          }`}
        >
          <div className='flex w-full items-center justify-between'>
            <span className='text-medium'>Submission ID: {id}</span>
          </div>
          <div className='flex w-full items-center justify-between'>
            <span className='text-xs text-medium'>{date}</span>
          </div>
        </li>
      )}
    </NavLink>
  )
}

export default function Page() {
  const { submissions } = useLoaderData<typeof loader>()
  let location = useLocation()
  return (
    <>
      <div
        className={clsx(
          location.pathname.endsWith('submissions') ? 'flex' : 'hidden lg:flex',
          'col-span-full h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-4'
        )}
      >
        {submissions.map((submission) => (
          <ListItem key={submission.id} {...submission} />
        ))}
      </div>
      <Outlet />
    </>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <div className='sticky top-4 col-span-full flex h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4 lg:col-span-6'>
        <h3 className='text-5xl font-medium text-error'>{error.status}</h3>
        <h3 className='text-lg font-medium text-medium'>{error.statusText}</h3>
        <h3 className='text-lg leading-6 text-strong'>
          {JSON.stringify(error.data)}
        </h3>
      </div>
    )
  }
  return <Error data={{ title: 'error.data.message' }} />
}
