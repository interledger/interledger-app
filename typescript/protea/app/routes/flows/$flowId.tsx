import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  Form,
  Outlet,
  useCatch,
  useLoaderData,
  useLocation
} from '@remix-run/react'
import type { FC } from 'react'
import React from 'react'
import { Icon, Error } from '~/components'
import { exitFlow, requireFlow, stepFlow } from '~/lib/flows.server'
import { requireUserSession } from '~/lib/kratos.server'

/**
 * Steps:
 * Initialise flow if there isn't one, much like how kratos does it. redirects to '/flows/:flowId/deposit/payment-method'
 * At every step, should save data and update index.
 */

export const loader: LoaderFunction = async ({ request, params }) => {
  await requireUserSession(request)
  const flow = await requireFlow(request, params)
  return json({
    flow
  })
}

export default function FlowLayout() {
  const { flow } = useLoaderData()

  return (
    <div className='w-full'>
      <Form
        id='flow-control'
        className='hidden'
        action={`/flows/${flow.id}`}
        method='post'
      />
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-between bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <div className='flex items-center'>
          {flow.stepIndex > 0 && (
            <button
              form='flow-control'
              name='route'
              value={parseInt(flow.stepIndex) - 1}
            >
              <div className='-ml-3 p-3 text-medium'>
                <Icon>arrow_back</Icon>
              </div>
            </button>
          )}
          <div className='flex items-center justify-start font-display text-2xl font-medium'>
            {flow.name}
          </div>
        </div>
        <button form='flow-control' name='route' value='exit'>
          <div className='-mr-3 p-3 text-medium'>
            <Icon>close</Icon>
          </div>
        </button>
      </header>
      {/* Body */}

      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <ProgressBar stepIndex={flow.stepIndex} steps={flow.steps} />
        <Outlet />
      </div>
    </div>
  )
}

type ProgressStep = {
  name: string
  route: string
}

type ProgressBarProps = {
  steps: ProgressStep[]
  stepIndex: number
}

const ProgressBar: FC<ProgressBarProps> = ({ steps, stepIndex }) => {
  return (
    <div className='col-span-full mb-4 flex space-x-2 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
      {steps.map((step, index) => (
        <div key={step.route} className='flex w-full flex-col space-y-2'>
          <div className='flex w-full justify-center font-display text-sm font-medium'>
            {step.name}
          </div>
          <div className='h-2 w-full rounded-full bg-container'>
            <div
              className={`h-2 rounded-full bg-container-primary transition-all duration-700 ${
                stepIndex >= index ? 'w-full' : 'w-0'
              }`}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

export function CatchBoundary() {
  const caught = useCatch()
  const location = useLocation()
  debugger
  if (caught.data.action) {
    caught.data.action.route = location.pathname.replace(
      /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/,
      'init'
    )
  }

  return <Error data={caught.data} />
}

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const routeTo = form.get('route')

  if (routeTo == 'exit') {
    return exitFlow(request)
  }
  await stepFlow(request, null, -1)
}
