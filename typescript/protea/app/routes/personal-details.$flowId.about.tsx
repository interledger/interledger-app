import type { FinishedUnaryCall } from '@protobuf-ts/runtime-rpc'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { PhoneAutocompleteOptions } from '~/components'
import { Button, Select, Shape, TextField } from '~/components'
import {
  flowType,
  getCurrentFlow,
  requireFlow,
  updateFlow
} from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import type { SetSignupUserDataResponse } from '~/generated/protobuf-ts/backend/v1/backend'
import { requireUserSession } from '~/lib/kratos.server'
import { useCallback, useState } from 'react'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderArgs) {
  const session = await requireUserSession(request)
  // await requireFlow(request, flowType.PersonalDetails, params, {
  //   id: params.flowId as string,
  //   startRoute: route('/personal-details/:flowId/about', {
  //     flowId: params.flowId as string
  //   }),
  //   data: {
  //     firstName: '',
  //     lastName: '',
  //     dateOfBirth: '',
  //     gender: '0'
  //   },
  //   defaultExitTo: route('/settings/linked-accounts')
  // })
  const flow = await getCurrentFlow(request, flowType.PersonalDetails)

  const maxDate = `${DateTime.now().toFormat('yyyy-MM-dd')}`

  // TODO: Pull whatever details we can from users current KYC data. (Should be set before flow data)
  // TODO: Can redirect past this step if the KYC data is already stored.

  return json({
    traits: session.identity.traits,
    flow,
    type: params.type,
    maxDate,
    genders: [
      // 0 Unknown, 1 Male, 2 Female, 3 Other
      { id: '1', name: 'Male' },
      { id: '2', name: 'Female' },
      { id: '3', name: 'Other' }
    ]
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { traits, flow, type, maxDate, genders } =
    useLoaderData<typeof loader>()

  const [gender, setGender] = useState<{ id: string; name: string }>(
    genders.find((gender) => flow?.data.gender == gender.id) || {
      id: '0',
      name: ''
    }
  )

  const _onChangeGender = useCallback((event) => {
    setGender(event)
  }, [])

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <div className='flex justify-between'>
          <h1 className='font-display text-2xl font-medium'>
            Personal details
          </h1>
          <div className='hidden sm:flex'>
            <Shape
              width={'w-8'}
              radius={'rounded-br-full'}
              color={'bg-rose-300'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-lime-500'}
            />
          </div>
        </div>
        <p className='mt-6 text-medium'>
          Please provide your personal details.
        </p>

        <Form
          id='personal-details-about'
          action={`/personal-details/${flow.id}/about`}
          method='post'
          className='hidden'
        />

        <TextField
          id='firstName'
          form='personal-details-about'
          label='First name'
          labelSuffix='1'
          name='firstName'
          defaultValue={traits.firstName || flow?.data.firstName}
          type='text'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors.firstName) || undefined}
          aria-describedby={
            actionData?.errors.firstName ? 'firstName-error' : undefined
          }
          required
          errorMessage={actionData?.errors.firstName}
        />

        <TextField
          id='lastName'
          form='personal-details-about'
          label='Last name'
          labelSuffix='1'
          name='lastName'
          defaultValue={traits.lastName || flow?.data.lastName}
          type='text'
          className='mt-1'
          aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
          aria-describedby={
            actionData?.errors.lastName ? 'lastName-error' : undefined
          }
          required
          errorMessage={actionData?.errors.lastName}
        />

        <TextField
          id='dateOfBirth'
          form='personal-details-about'
          label='Birth date'
          name='dateOfBirth'
          max={maxDate}
          defaultValue={flow?.data.dateOfBirth}
          type='date'
          className='mt-1'
          aria-invalid={Boolean(actionData?.errors.dateOfBirth) || undefined}
          aria-describedby={
            actionData?.errors.dateOfBirth ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors.dateOfBirth}
        />

        <Select
          id='gender'
          label='Gender'
          value={gender}
          options={genders}
          onChange={_onChangeGender}
          aria-invalid={Boolean(actionData?.errors.gender) || undefined}
          aria-describedby={
            actionData?.errors.gender ? 'gender-error' : undefined
          }
          errorMessage={actionData?.errors.gender}
        />
        <input
          form='personal-details-about'
          value={gender.id}
          name='gender'
          type='hidden'
        />

        <Button className='mt-12' form='personal-details-about' type='submit'>
          Continue
        </Button>
      </div>
      <div className='flex w-full flex-col p-4'>
        <span className='text-xs text-medium'>
          <sup>1</sup> First and last name as they appear on your government
          issued ID document.
        </span>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'FirstName' | 'LastName' | 'gender' | 'dateOfBirth'

function mapper(
  field: fieldErrorsMap
): 'firstName' | 'lastName' | 'gender' | 'dateOfBirth' | null {
  switch (field) {
    case 'FirstName':
      return 'firstName'
    case 'LastName':
      return 'lastName'
    case 'gender':
      return 'gender'
    case 'dateOfBirth':
      return 'dateOfBirth'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const dateOfBirth = form.get('dateOfBirth') as string
  const gender = form.get('gender') as string

  const fieldErrors = {
    firstName: '',
    lastName: '',
    dateOfBirth: '',
    gender: ''
  }

  const ipAddress = request.headers.get('x-forwarded-for') as string

  const seconds = `${DateTime.fromFormat(
    dateOfBirth,
    'yyyy-MM-dd'
  ).toSeconds()}`

  const timestamp = {
    seconds,
    nanos: 0
  }

  let response = await grpcClient
    .updateIndividualKYC(
      {
        firstName,
        lastName,
        gender: parseInt(gender),
        dateOfBirth: timestamp,
        ipAddress
      },
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
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

  const headers = await updateFlow(request, flowType.PersonalDetails, {
    firstName,
    lastName,
    gender,
    dateOfBirth,
    address: {}
  })

  return redirect(
    route('/personal-details/:flowId/address', {
      flowId: params.flowId as string
    }),
    { headers }
  )
}
