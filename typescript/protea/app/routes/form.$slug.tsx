import type { ActionArgs, LoaderArgs } from '@remix-run/node'

import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData, useParams } from '@remix-run/react'
import clsx from 'clsx'
import { useCallback, useEffect } from 'react'
import { toRemixMeta } from 'react-datocms'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Layouts } from '~/components'
import { FormSectionRecordComponent } from '~/components/Content'
import type { FormSectionRecord } from '~/generated/dato-cms-graphql'
import { connectClient } from '~/lib/connect.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { hasUserSession } from '~/lib/kratos.server'
import { getCurrentFormPage } from '~/lib/marketing.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useFormStore } from '~/lib/useFormStore'

export async function loader({ request, params }: LoaderArgs) {
  const { form } = await getCurrentFormPage({
    filter: { slug: { eq: params.slug } }
  })

  if (!form) throw json(null, { status: 404, statusText: 'Not found' })

  if (form.requireAuth && !hasUserSession(request))
    throw redirect(`/login?returnTo=/form/${params.slug}`)

  return jsonWithCSRF(request, { form })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: { title: (match) => match.data.form.title, back: 'form' }
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.form.seoMeta),
    'twitter:url': 'https://fynbos.app/form/' + params.slug,
    'og:url': 'https://fynbos.app/form/' + params.slug
  }
}

export default function Page() {
  const { form, csrfToken } = useLoaderData<typeof loader>()
  const [step, stepForward, reset] = useFormStore((state) => [
    state.step,
    state.stepForward,
    state.reset
  ])
  const { slug } = useParams()

  useEffect(() => {
    // This ensures that the state is only cleared when this route is unmounted.
    return () => {
      reset()
    }
  }, [reset])

  const _onClick = useCallback<{
    (): void
  }>(() => {
    stepForward()
  }, [stepForward])

  return (
    <>
      <Form
        id={`dynamic-${slug}`}
        action={route('/form/:slug', { slug: slug as string })}
        method='post'
        className='hidden'
      />
      <input
        form={`dynamic-${slug}`}
        defaultValue={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form={`dynamic-${slug}`}
        defaultValue={form.returnTo || '/'}
        name='returnTo'
        type='hidden'
      />
      {form.sections.map((section, index, allSections) => (
        <div
          key={section.id}
          className={clsx(index !== step ? 'hidden' : 'contents')}
        >
          <FormSectionRecordComponent content={section as FormSectionRecord} />
          {allSections.length - 1 == index && (
            <Button form={`dynamic-${slug}`} type='submit'>
              {section.buttonText}
            </Button>
          )}
          {allSections.length - 1 > index && (
            <Button type='button' onClick={_onClick}>
              {section.buttonText}
            </Button>
          )}
        </div>
      ))}
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const returnTo = form.get('returnTo') as string

  let data: Record<string, string> = {}
  form.forEach((value, key) => {
    if (typeof value === 'string' && key !== 'csrfToken' && key !== 'returnTo')
      data[key] = value
  })

  const response = connectClient.createDynamicForm(request, {
    formId: params.slug as string,
    data: JSON.stringify(data)
  })

  if (isConnectError(response)) throw response.errorResponse

  if (returnTo !== '/')
    return redirect(route('/thank-you/:slug', { slug: returnTo }))

  return redirectWithSnackbar(request, route('/'), {
    message: 'Form submitted successfully'
  })
}
