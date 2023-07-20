import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
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
  WalletGrid
} from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  return json({ traits: session.identity.traits })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Support'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Support'
  }
}

export default function Page() {
  const { traits } = useLoaderData<typeof loader>()
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
      <GridColumn className='col-span-full lg:col-span-7'>
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
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  return redirect('/', {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Support ticket created.',
        icon: 'close'
      })
    }
  })
}
