import { useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  CardContent,
  Checkbox,
  Router,
  Snackbar,
  TextField
} from '~/components'
import { useSignupStore } from '~/lib/useSignupStore'
import type { loader } from './route'

export function Password() {
  const passwordFetcher = useFetcher()
  const { kratosFlowId, kratosCsrfToken, csrfToken } =
    useLoaderData<typeof loader>()

  const [snackbarMessage, setSnackbar] = useState<any>(
    passwordFetcher.data?.errors?.form
  )
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    Boolean(passwordFetcher.data?.errors?.form) ?? false
  )

  const [firstName, lastName, email, country, id, phone] = useSignupStore(
    (state) => [
      state.firstName,
      state.lastName,
      state.email,
      state.country,
      state.id,
      state.phone
    ]
  )

  useEffect(() => {
    if (passwordFetcher.data?.errors?.form) {
      setSnackbar(passwordFetcher.data?.errors?.form)
      setShowSnackbar(true)
    }
  }, [passwordFetcher.data?.errors?.form])

  return (
    <>
      <passwordFetcher.Form
        id='signup-password'
        action={route('/signup')}
        method='post'
        className='hidden'
      />
      <input
        defaultValue={csrfToken}
        name='csrfToken'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={kratosCsrfToken}
        name='csrf_token'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={kratosFlowId}
        name='kratosFlowId'
        form='signup-password'
        type='hidden'
      />
      <input defaultValue={id} name='id' form='signup-password' type='hidden' />
      <input
        defaultValue={firstName}
        name='firstName'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={lastName}
        name='lastName'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={country?.id as string}
        name='country'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={email}
        name='email'
        form='signup-password'
        type='hidden'
      />
      <input
        defaultValue={phone}
        name='phone'
        form='signup-password'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Create a password to log in to your account securely.</p>
        </CardContent>
        <TextField
          id='password'
          label='Password'
          name='password'
          form='signup-password'
          type='password'
          className='mt-2'
          aria-invalid={
            Boolean(passwordFetcher.data?.errors?.password) || undefined
          }
          aria-describedby={
            passwordFetcher.data?.errors?.password
              ? 'password-error'
              : undefined
          }
          required
          errorMessage={passwordFetcher.data?.errors?.password}
        />
      </Card>
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='service-agreement'
            form='signup-password'
            className='flex'
            aria-invalid={
              Boolean(passwordFetcher.data?.errors?.serviceAgreement) ||
              undefined
            }
            aria-describedby={
              passwordFetcher.data?.errors?.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={passwordFetcher.data?.errors?.serviceAgreement}
          >
            I agree to the Fynbos&nbsp;
            <Router className='text-primary' to='/legal/privacy-policy'>
              Privacy Policy
            </Router>
            ,&nbsp;
            <Router className='text-primary' to='/legal/terms-of-service'>
              Terms of Use
            </Router>
            , and&nbsp;
            <Router className='text-primary' to='/legal/us/e-sign-agreement'>
              E-sign Agreement
            </Router>
            .
          </Checkbox>
        </CardContent>
      </Card>
      <Button
        form='signup-password'
        name='formName'
        value='password'
        type='submit'
      >
        Confirm
      </Button>
      <Snackbar
        message={snackbarMessage}
        icon='close'
        show={showSnackbar}
        id='error-snackbar'
        onClose={() => {
          setSnackbar('')
          setShowSnackbar(false)
        }}
      />
    </>
  )
}
