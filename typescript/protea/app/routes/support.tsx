import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  TextArea,
  WalletGrid,
  WalletShapes
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error } from '~/lib/error.server'
import { getUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import { StatusError, grpcClient, isGrpcError } from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  return jsonWithCSRF(request, {
    traits: session.identity.traits
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Support',
      actions: <WalletShapes />
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Support'
  }
}

export default function Page() {
  const { traits, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <WalletGrid>
      <Form
        id='support-form'
        action={route('/support')}
        method='post'
        className='hidden'
      />
      <input
        form='support-form'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        defaultValue={traits.firstName}
        name='firstName'
        form='support-form'
        type='hidden'
      />
      <input
        defaultValue={traits.lastName}
        name='lastName'
        form='support-form'
        type='hidden'
      />
      <input
        defaultValue={traits.email}
        name='email'
        form='support-form'
        type='hidden'
      />
      <GridColumn className='col-span-full lg:col-span-6'>
        <Card>
          <CardHeader>
            <CardTitle>Contact support</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              Please share all relevant details so we can assist you effectively
              and efficiently.
            </p>
          </CardContent>
          <TextArea
            id='description'
            form='support-form'
            label='Your message'
            name='description'
            className='mt-2'
            aria-invalid={Boolean(actionData?.errors.description) || undefined}
            aria-describedby={
              actionData?.errors.description ? 'description-error' : undefined
            }
            required
            errorMessage={actionData?.errors.description}
          />
        </Card>
        <Button form='support-form' type='submit'>
          Submit
        </Button>
        <Card>
          <CardHeader>
            <CardTitle>Support details</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              Our telephone support lines are open Monday to Friday between 9am
              and 5pm EST.
            </p>
            <div className='mt-4 flex items-center space-x-2 text-medium'>
              <Icon>call</Icon>
              <AnchorRouter
                to='tel:+1 (856) 249-3067'
                className='text-sm text-primary'
              >
                +1 (856) 249-3067
              </AnchorRouter>
            </div>
            <div className='mt-4 flex items-center space-x-2 text-medium'>
              <Icon>mail</Icon>
              <AnchorRouter
                to='mailto:support@fynbos.app'
                className='text-sm text-primary'
              >
                support@fynbos.app
              </AnchorRouter>
            </div>
          </CardContent>
        </Card>
      </GridColumn>
    </WalletGrid>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'FirstName' | 'Email' | 'LastName' | 'Description'

function mapper(
  field: fieldErrorsMap
): 'firstName' | 'email' | 'lastName' | 'description' | null {
  switch (field) {
    case 'Email':
      return 'email'
    case 'FirstName':
      return 'firstName'
    case 'LastName':
      return 'lastName'
    case 'Description':
      return 'description'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const email = form.get('email') as string
  const description = form.get('description') as string

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: '',
    firstName: '',
    lastName: '',
    email: '',
    description: ''
  }

  let response = await grpcClient
    .createSupportTicket(
      {
        description,
        firstName,
        lastName,
        email
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
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return error(request, { errors: { ...fieldErrors } })
    } else
      return error(
        request,
        { errors: { ...fieldErrors } },
        { action: 'Contact support' }
      )
  }

  return redirectWithSnackbar(request, route('/'), {
    message: 'Support ticket created.',
    icon: 'close'
  })
}
