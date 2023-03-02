import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import {
  AnchorRouter,
  Button,
  Card,
  Checkbox,
  Layouts,
  Shape
} from '~/components'
import { flowType, requireFlow } from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { getUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  let flow = await requireFlow(request, flowType.PersonalDetails)

  return json({
    flow,
    country: { id: session.identity.traits.countryCode, name: '' }
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Machnet privacy and terms'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()

  return (
    <Card>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Privacy and Terms</h1>
        <div className='hidden sm:flex'>
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            flex='flex-none'
            width={'w-8'}
            radius={'rounded-tr-full'}
            color={'bg-rose-400'}
          />
        </div>
      </div>

      <p className='mt-6 text-medium'>
        Fynbos uses Machnet to power our financial services for our customers in
        the United States.
      </p>
      <p className='mt-6 text-medium'>
        Please read and agree to their privacy policy and terms of service to
        continue.
      </p>
      <div className='mt-6'>
        <img
          className='aspect-auto h-6'
          alt='Machnet logo'
          src='https://cdn.fynbos.app/logos/machnet.png'
        />
      </div>

      <Form
        id='personal-details-machnet'
        action='/machnet-terms'
        method='post'
        className='hidden'
      />
      <Checkbox
        id='service-agreement'
        name='service-agreement'
        form='personal-details-machnet'
        className='mt-8 flex'
        aria-invalid={Boolean(actionData?.errors.serviceAgreement) || undefined}
        aria-describedby={
          actionData?.errors.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.errors.serviceAgreement}
      >
        I agree to the Machnet&nbsp;
        <AnchorRouter
          className='text-primary'
          to={
            'https://machnetservices.com/fynbos-technologies-llc-privacypolicy/'
          }
        >
          Privacy Policy
        </AnchorRouter>
        &nbsp;and&nbsp;
        <AnchorRouter
          className='text-primary'
          to={
            'https://machnetservices.com/fynbos-technologies-llc-termsofservice/'
          }
        >
          Terms of Service
        </AnchorRouter>
        .
      </Checkbox>

      <Button className='mt-12' form='personal-details-machnet' type='submit'>
        Continue
      </Button>
    </Card>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  const serviceAgreement = form.get('service-agreement') as string
  // TODO: use actual service agreement service
  const fieldErrors = {
    form: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to agree to continue.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  let res = await grpcClient
    .startMachnetKYC(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(res)) throw json({}, httpMapping(res.code))

  const flow = await requireFlow(request, flowType.PersonalDetails)
  // NOTE Temporarily not exciting this flow so that if the user needs to fix something their data will be there.
  // We should find a better way to do this.
  // await exitFlow(request, flowType.PersonalDetails)
  return redirect(flow.returnTo)
}
