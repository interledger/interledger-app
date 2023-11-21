import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { redirect } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import type {
  ApplicationProps,
  TurnstileAppearance,
  TurnstileInstance
} from '~/components'
import {
  AnchorRouter,
  Icon,
  Layouts,
  OutlineButton,
  TextArea,
  TextField
} from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import Turnstile from '~/components/Turnstile'
import { getContactRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { datoMeta, mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const { contactRoute, footer } = await getContactRoute()
  const cfTurnstileSiteKey = process.env.CF_TURNSTILE_SITE_KEY || ''

  return jsonWithCSRF(request, { contactRoute, footer, cfTurnstileSiteKey })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer
  }
}

export const meta: MetaFunction<typeof loader> = mergeMeta(
  ({ data, location }) => datoMeta(data?.contactRoute?._seoMetaTags, location)
)

export default function Page() {
  const { contactRoute, csrfToken, cfTurnstileSiteKey } =
    useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const [turnstileToken, setTurnstileToken] = useState('')
  const [turnstileAppearance, setTurnstileAppearance] =
    useState<TurnstileAppearance>('interaction-only')
  let turnstileRef = useRef<TurnstileInstance>(null)

  useEffect(() => {
    if (actionData?.errors.captcha) {
      setTurnstileAppearance('always')
      setTurnstileToken('')
      turnstileRef.current?.reset()
      actionData.errors.captcha = ''
    }
  }, [actionData])

  return (
    <>
      {contactRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        >
          <div className='grid w-full grid-cols-12 gap-y-6 px-4 lg:px-0'>
            <div className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <h2 className='font-display text-2xl font-medium'>
                Send us a message
              </h2>
            </div>
            <Form
              id='contact-form'
              action={route('/contact')}
              method='post'
              className='hidden'
            />
            <input
              form='contact-form'
              value={csrfToken}
              name='csrfToken'
              type='hidden'
            />
            <input
              form='contact-form'
              value={turnstileToken}
              name='cf-turnstile-response'
              type='hidden'
            />
            <TextField
              id='firstName'
              form='contact-form'
              label='First name'
              name='firstName'
              type='text'
              className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
              aria-invalid={Boolean(actionData?.errors.firstName) || undefined}
              aria-describedby={
                actionData?.errors.firstName ? 'firstName-error' : undefined
              }
              errorMessage={actionData?.errors.firstName}
            />

            <TextField
              id='lastName'
              form='contact-form'
              label='Last name'
              name='lastName'
              type='text'
              className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
              aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
              aria-describedby={
                actionData?.errors.lastName ? 'lastName-error' : undefined
              }
              errorMessage={actionData?.errors.lastName}
            />

            <TextField
              id='email'
              form='contact-form'
              label='Email address*'
              name='email'
              type='text'
              className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
              aria-invalid={Boolean(actionData?.errors.email) || undefined}
              aria-describedby={
                actionData?.errors.email ? 'email-error' : undefined
              }
              required
              errorMessage={actionData?.errors.email}
            />

            <TextArea
              id='description'
              form='contact-form'
              label='Details / comments*'
              name='description'
              className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
              aria-invalid={
                Boolean(actionData?.errors.description) || undefined
              }
              aria-describedby={
                actionData?.errors.description ? 'description-error' : undefined
              }
              required
              errorMessage={actionData?.errors.description}
            />
            <div className='col-span-full flex justify-start sm:col-span-3 sm:col-start-2 lg:col-start-4'>
              <Turnstile
                ref={turnstileRef}
                siteKey={cfTurnstileSiteKey}
                appearance={turnstileAppearance}
                onSuccess={(token) => {
                  setTurnstileToken(token)
                }}
              />
            </div>
            <div className='col-span-full flex justify-start sm:col-span-3 sm:col-start-2 lg:col-start-4'>
              <OutlineButton
                className='h-20 px-20'
                form='contact-form'
                type='submit'
              >
                Submit
              </OutlineButton>
            </div>
            <div className='col-span-full mt-10 flex flex-col justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <h2 className='font-display font-medium text-strong'>Support</h2>
              <span className='mt-4 text-sm'>
                Our telephone support lines are open Monday to Friday between
                9am and 5pm EST.
              </span>
              <div className='mt-3 flex items-center space-x-2 text-medium'>
                <Icon>call</Icon>
                <AnchorRouter
                  to='tel:+1 (856) 249-3067'
                  className='text-sm text-primary'
                >
                  +1 (856) 249-3067
                </AnchorRouter>
              </div>
              <div className='mt-2 flex items-center space-x-2 text-medium'>
                <Icon>mail</Icon>
                <AnchorRouter
                  to='mailto:support@fynbos.app'
                  className='text-sm text-primary'
                >
                  support@fynbos.app
                </AnchorRouter>
              </div>
            </div>
          </div>
        </MarketingPageWithSections>
      ))}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const email = form.get('email') as string
  const description = form.get('description') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    firstName: '',
    lastName: '',
    description: '',
    email: '',
    captcha: ''
  }

  // cloudflare turnstile captcha validation
  const turnstileToken = form.get('cf-turnstile-response') as string
  const validateTurnstileData = new FormData()
  validateTurnstileData.append('response', turnstileToken)
  validateTurnstileData.append(
    'secret',
    process.env.CF_TURNSTILE_SECRET_KEY || ''
  )
  validateTurnstileData.append('remoteip', getClientIP(request))
  const res = await fetch(
    'https://challenges.cloudflare.com/turnstile/v0/siteverify',
    {
      method: 'POST',
      body: validateTurnstileData
    }
  )
  const body = await res.json()
  if (!body.success) {
    errors.captcha = 'There was an error validating the captcha.'
    return error(request, { errors }, { message: errors.captcha })
  }

  let response = await grpc.createSupportTicket(request, {
    description,
    firstName,
    lastName,
    email
  })

  if (isConnectError(response)) return response.error({ errors })

  return redirect('/contact/success')
}
