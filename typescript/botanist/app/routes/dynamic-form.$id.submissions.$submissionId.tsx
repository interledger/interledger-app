import type { LoaderArgs } from '@remix-run/node'

import { Error, GridCard, GridCardError } from '~/components'
import { json } from '@remix-run/node'
import {
  isRouteErrorResponse,
  useLoaderData,
  useRouteError
} from '@remix-run/react'
import { GetFormSubmissionDetails } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const submission = await GetFormSubmissionDetails(
    request,
    params.submissionId as string
  )

  const submissionData = JSON.parse(submission.data)

  for (const key in submissionData) {
    submissionData[slugToSentence(key)] = submissionData[key]
    delete submissionData[key]
  }

  return json({
    submissionData
  })
}

export default function Page() {
  const { submissionData } = useLoaderData<typeof loader>()

  return (
    <GridCard
      className='sticky top-4 col-span-full lg:col-span-4'
      options={submissionData}
    />
  )
}

export function ErrorBoundary() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    return (
      <GridCardError
        className='sticky top-4 col-span-full lg:col-span-4'
        error={error}
      />
    )
  }
  return <Error data={{ title: 'error.data.message' }} />
}

function slugToSentence(slug: string) {
  return slug
    .split('-')
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join(' ')
}
