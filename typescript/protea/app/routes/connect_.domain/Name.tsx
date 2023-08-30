import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, CardContent, TextField } from '~/components'
import type { action, loader } from './route'

export function Name() {
  const actionData = useActionData<typeof action>()
  const { csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='connect-domain'
        action={route('/connect/domain')}
        method='post'
        className='hidden'
      />
      <input
        form='connect-domain'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p className='text-medium'>
            Please provide the domain name to connect.
          </p>
        </CardContent>
        <TextField
          id='domainName'
          form='connect-domain'
          label='Domain name'
          name='domainName'
          placeholder='example.com'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.domainName) || undefined}
          aria-describedby={
            actionData?.errors?.domainName ? 'domainName-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.domainName}
        />
      </Card>

      <Button form='connect-domain' type='submit'>
        Continue
      </Button>
    </>
  )
}
