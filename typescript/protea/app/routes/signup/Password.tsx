import { useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  AnchorRouter,
  Button,
  Card,
  CardContent,
  Checkbox,
  Router,
  TextField
} from '~/components'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useSignupStore } from '~/lib/useSignupStore'
import type { loader, passwordAction } from './route'

export function Password() {
  const passwordFetcher = useFetcher<typeof passwordAction>()
  const errors = passwordFetcher.data?.errors
  const { kratosFlowId, kratosCsrfToken, csrfToken } =
    useLoaderData<typeof loader>()

  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])
  const [serviceAgreement, setServiceAgreement] = useState(false)

  useEffect(() => {
    if (errors?.form) {
      pushSnackbar({
        id: errors?.form,
        message: errors?.form,
        icon: 'close',
        canShow: true
      })
    }

    if (errors?.phone) {
      pushSnackbar({
        id: errors?.phone,
        message: errors?.phone,
        icon: 'close',
        canShow: true
      })
    }
  }, [errors, pushSnackbar])

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
            onChange={(e) => setServiceAgreement(e.target.checked)}
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
            I agree to the Interledger Wallet&nbsp;
            <Router className='text-primary' to='/legal/privacy-policy'>
              Privacy Policy
            </Router>
            ,&nbsp;
            <Router className='text-primary' to='/legal/terms-of-service'>
              Terms of Use
            </Router>
            , &nbsp;
            <Router className='text-primary' to='/legal/us/e-sign-agreement'>
              E-sign Agreement
            </Router>
            , &nbsp;
            <AnchorRouter to='https://astrafi.com/terms'>
              Astra terms
            </AnchorRouter>
            , and &nbsp;
            <AnchorRouter to='https://astrafi.com/privacy'>
              Astra privacy policy
            </AnchorRouter>
            .
          </Checkbox>
        </CardContent>
      </Card>
      <Button
        disabled={!serviceAgreement}
        form='signup-password'
        name='formName'
        value='password'
        type='submit'
      >
        Confirm
      </Button>
    </>
  )
}
