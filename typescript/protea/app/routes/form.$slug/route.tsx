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
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { hasUserSession } from '~/lib/kratos.server'
import { getCurrentFormPage } from '~/lib/marketing.server'
import { grpcClient, httpMapping, isGrpcError } from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useFormStore } from '~/lib/useFormStore'

export async function loader({ request, params }: LoaderArgs) {
  const { form } = await getCurrentFormPage({
    filter: { slug: { eq: params.slug } }
  })

  if (!form) throw json(null, { status: 404, statusText: 'Not found' })

  if (form.requireAuth && !hasUserSession(request)) throw redirect('/login')

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
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()

  await validateCSRFToken(request, form)

  let data: Record<string, string> = {}
  form.forEach((value, key) => {
    if (typeof value === 'string' && key !== 'csrfToken') data[key] = value
  })

  const response = grpcClient.createDynamicForm(
    {
      formId: params.slug as string,
      data: JSON.stringify(data)
    },
    {
      meta: {
        cookies: cookie || ''
      }
    }
  )
  if (isGrpcError(response)) throw json({}, httpMapping(response.code))

  return redirectWithSnackbar(request, route('/'), {
    message: 'Form submitted successfully'
  })
}
