import { json, redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import type { Params } from 'react-router-dom'
import { v4 } from 'uuid'
import { getSession, commitSession } from '~/sessions'

export enum flowType {
  Pay = 'pay',
  Signup = 'signup',
  LinkAccount = 'link-account'
}

type Flow = {
  id: string
  data?: any
  startRoute: string
  defaultExitTo: string
}

type Flows = {
  [flowType.Pay]: Flow | null
  [flowType.Signup]: Flow | null
  [flowType.LinkAccount]: Flow | null
}

/**
 * Sets up the flow pseudo-stack:
 * 1. Allows stacking flows of different types on top of each other.
 * 2. Flows of the same type will only be added if a previous flow is complete.
 *
 * @param request Request received in a loader function.
 * @param type The flow type used to identify the flow.
 * @param params Params received in a loader function.
 * @returns returns the current flow (Top of stack), or throws redirects to the newly created flow.
 */
export async function requireFlow(
  request: Request,
  type: flowType,
  params?: Params
): Promise<Headers> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  let flows: Flows = userSettings.get('flows') || {
    pay: null,
    signup: null,
    linkAccount: null
  }

  let currentFlow = flows[type]
  const headers = new Headers()

  if (currentFlow != null) {
    if (
      typeof params?.flowId !== 'undefined' &&
      params?.flowId != currentFlow.id
    ) {
      flows[type] = null
      userSettings.set('flows', flows)
      headers.append('Set-Cookie', await commitSession(userSettings))
      throw redirect(currentFlow.startRoute, {
        headers
      })
    }
  } else {
    currentFlow = flowTemplate(v4(), type)
    flows[type] = currentFlow

    userSettings.set('flows', flows)
    headers.append('Set-Cookie', await commitSession(userSettings))

    const url = new URL(request.url)
    if (url.pathname != currentFlow.startRoute)
      throw redirect(currentFlow.startRoute, {
        headers
      })
  }
  return headers
}

export async function getCurrentFlow(
  request: Request,
  type: flowType
): Promise<Flow> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flows = userSettings.get('flows') || {
    pay: null,
    signup: null,
    linkAccount: null
  }
  let currentFlow = flows[type]

  if (currentFlow == null) {
    throw json(
      {
        title: 'No flows found.'
      },
      { status: 400, statusText: 'Bad Request' }
    )
  }
  return currentFlow
}

/**
 * Allows adding/updating data to a flows data.
 * @param request
 * @param type The flow type used to identify the flow.
 * @param data A data object to me merged into the flow data
 * @returns Headers
 */
export async function updateFlow(
  request: Request,
  type: flowType,
  data?: any
): Promise<Headers> {
  const headers = new Headers()
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flows = userSettings.get('flows')
  const currentFlow = flows[type]

  if (currentFlow) {
    if (data) {
      Object.assign(currentFlow.data, data)
    }
    flows[type] = currentFlow
    userSettings.set('flows', flows)

    headers.append('Set-Cookie', await commitSession(userSettings))

    return headers
  }
  throw json(
    {
      title: "Can't update/complete a flow that doesn't exist"
    },
    { status: 400, statusText: 'Bad Request' }
  )
}

/**
 * Allows removing a flow from the stack and returning to the next flow in the stack,
 * Or returns headers allowing the consumer to redirect out of the flow.
 * @param request Request
 * @param type The flow type used to identify the flow.
 * @returns Headers
 */
export async function exitFlow(
  request: Request,
  type: flowType
): Promise<Headers> {
  const headers = new Headers()
  const userSettings = await getSession(request.headers.get('Cookie'))
  const flows: Flows = userSettings.get('flows')

  flows[type] = null

  userSettings.set('flows', flows)
  headers.append('Set-Cookie', await commitSession(userSettings))

  return headers
}

const flowTemplate = (id: string, type: flowType): Flow => {
  switch (type) {
    case flowType.Pay:
      return {
        id: id,
        startRoute: route('/pay'),
        data: {},
        defaultExitTo: '/'
      }
    case flowType.LinkAccount:
      return {
        id,
        startRoute: route('/'),
        data: {},
        defaultExitTo: route('/settings/linked-accounts')
      }
    case flowType.Signup:
      return {
        id,
        startRoute: route('/signup'),
        data: {},
        defaultExitTo: route('/')
      }
    default:
      throw json(
        {
          title: "You specified a flow type that doesn't exist"
        },
        { status: 400, statusText: 'Bad Request' }
      )
  }
}
