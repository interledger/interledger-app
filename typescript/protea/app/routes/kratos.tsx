import type { MetaFunction } from '@remix-run/node'
import { Link } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, CardHeader, CardTitle, Layouts } from '~/components'
import { mergeMeta } from '~/lib/meta'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Kratos Flow Testing'
  }
])

export default function KratosIndex() {
  const flows = [
    { name: 'Login Flow', path: '/kratos/login', description: 'Test login flow with SDK' },
    { name: 'Registration Flow', path: '/kratos/register', description: 'Test registration flow with SDK' },
    { name: 'Logout Flow', path: '/kratos/logout', description: 'Test logout flow with SDK' },
    { name: 'Verification Flow', path: '/kratos/verify', description: 'Test verification flow with SDK' }
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kratos Flow Testing</CardTitle>
      </CardHeader>
      <CardContent>
        <p className='mb-6 text-sm text-medium'>
          Test Kratos flows using the official SDK before implementing them in the application.
        </p>
        <ul className='space-y-4'>
          {flows.map((flow) => (
            <li key={flow.path}>
              <Link
                to={flow.path}
                className='block rounded-lg border border-stroke p-4 transition-colors hover:bg-mercury'
              >
                <h3 className='font-medium text-primary'>{flow.name}</h3>
                <p className='text-sm text-medium'>{flow.description}</p>
              </Link>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}
