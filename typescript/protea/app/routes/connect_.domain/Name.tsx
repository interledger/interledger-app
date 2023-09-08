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
          id='domain'
          form='connect-domain'
          label='Domain name'
          name='domain'
          placeholder='example.com'
          type='text'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.domain) || undefined}
          aria-describedby={
            actionData?.errors?.domain ? 'domain-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.domain}
        />
      </Card>

      <Button form='connect-domain' type='submit'>
        Continue
      </Button>
    </>
  )
}
