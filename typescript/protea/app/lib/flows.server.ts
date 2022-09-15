import { json, redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import type { Params } from 'react-router-dom'
import { v4 } from 'uuid'
import { getSession, commitSession } from '~/sessions'
import invariant from 'ts-invariant'

export enum flowType {
  Deposit = 'deposit',
  Withdraw = 'withdraw',
  Send = 'send',
  Signup = 'signup',
  LinkedAccount = 'linked-account',
  UnitOnboarding = 'unit-onboarding'
}

type ProgressStep = {
  name: string
  route: string
}

type Flow = {
  id: string
  type: flowType
  data?: any
  currentRoute: string
  steps: ProgressStep[]
  complete: boolean
  defaultExitTo: string
}

/**
 * requireFlow should only be called in the layout route of the flow - $flowId.tsx
 * DANGER: Don't call this anywhere else, you will run into race conditions.
 *
 * If you want to fetch a specific flow, use getFlow with the flowId instead.
 *
 * Sets up the flow pseudo-stack:
 * 1. Allows stacking flows of different types on top of each other.
 * 2. Flows of the same type will only be added if a previous flow is complete.
 *
 * @param request Request received in a loader function.
 * @param params Params received in a loader function.
 * @returns returns the current flow (Top of stack), or throws redirects to the newly created flow.
 */
export async function requireFlow(
  request: Request,
  params: Params
): Promise<Flow> {
  const userSettings = await getSession(request.headers.get('Cookie'))
  const flowId = params.flowId

  invariant(flowId, 'Expected flowId param, but found none.')

  const url = new URL(request.url)
  const pathname = url.pathname
  const type = pathname.split('/')[3]
  let flows: Flow[] = userSettings.get('flows') || []

  let currentFlow = flows.find((flow) => !flow.complete && flow.id == flowId)

  if (currentFlow) {
    const del = flows.splice(flows.findIndex((flow) => flow.id == flowId) + 1)

    if (del.length > 0) {
      // del ensures unnecessary/erraneous flows are pruned from the flow stack.
      // We must ensure that flows is set by redirecting.
      userSettings.set('flows', flows)
      throw redirect(currentFlow.currentRoute, {
        headers: {
          'Set-Cookie': await commitSession(userSettings)
        }
      })
    }

    return currentFlow
  } else {
    currentFlow = flows.find((flow) => !flow.complete && flow.type === type)

    if (currentFlow) {
      flows.splice(flows.findIndex((flow) => flow.type == type) + 1)
    } else {
      currentFlow = flowTemplate(v4(), type)
      if (!currentFlow)
        throw json(
          {
            title: 'Expected specific flow type, but found something else.'
          },
          { status: 400, statusText: 'Bad Request' }
        )
      flows.push(currentFlow)
    }

    userSettings.set('flows', flows)
    throw redirect(currentFlow.currentRoute, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }
}

export async function getCurrentFlow(
  request: Request,
  params: Params
): Promise<Flow | undefined> {
  const userSettings = await getSession(request.headers.get('Cookie'))
  invariant(params.flowId, 'Expected flowId param, but found none.')
  const flows: Flow[] = userSettings.get('flows') || []
  return flows.find((flow) => flow.id == params.flowId)
}

/**
 * Allows adding/updating data to a flows data.
 * @param request
 * @param data A data object to me merged into the flow data
 * @returns Headers
 */
export async function updateFlow(
  request: Request,
  data?: any,
  complete?: boolean
): Promise<Headers> {
  const headers = new Headers()
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const currentFlow = flows.pop()
  if (currentFlow) {
    if (data) {
      Object.assign(currentFlow.data, data)
    }
    if (typeof complete != 'undefined') currentFlow.complete = complete
    flows.push(currentFlow)
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
 * @returns Headers
 */
export async function exitFlow(request: Request): Promise<Headers> {
  const headers = new Headers()
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const exitedFlow = flows.pop()
  const currentFlow = flows.at(-1)
  userSettings.set('flows', flows)

  headers.append('Set-Cookie', await commitSession(userSettings))

  if (currentFlow) {
    // If there is a flow still in the stack, redirect there.
    throw redirect(currentFlow.currentRoute, {
      headers
    })
  } else if (exitedFlow) {
    return headers
  }
  throw json(
    {
      title: 'Your flow timed out, please restart it to continue.',
      action: { text: 'Restart flow' }
    },
    { status: 504, statusText: 'Gateway Timeout' }
  )
}

const flowTemplate = (id: string, type: string): Flow | undefined => {
  switch (type) {
    case flowType.Deposit:
      return {
        id: id,
        type: flowType.Deposit,
        currentRoute: route('/flows/:flowId/deposit/linked-account', {
          flowId: id
        }),
        data: {},
        steps: [
          {
            name: 'From',
            route: route('/flows/:flowId/deposit/linked-account', {
              flowId: id
            })
          },
          {
            name: 'Amount',
            route: route('/flows/:flowId/deposit/amount', { flowId: id })
          },
          {
            name: 'Review',
            route: route('/flows/:flowId/deposit/review', { flowId: id })
          }
        ],
        complete: false,
        defaultExitTo: '/'
      }
    case flowType.Withdraw:
      return {
        id: id,
        type: flowType.Withdraw,
        currentRoute: route('/flows/:flowId/withdraw/linked-account', {
          flowId: id
        }),
        data: {},
        steps: [
          {
            name: 'To',
            route: route('/flows/:flowId/withdraw/linked-account', {
              flowId: id
            })
          },
          {
            name: 'Amount',
            route: route('/flows/:flowId/withdraw/amount', { flowId: id })
          },
          {
            name: 'Review',
            route: route('/flows/:flowId/withdraw/review', { flowId: id })
          }
        ],
        complete: false,
        defaultExitTo: '/'
      }
    case flowType.Send:
      return {
        id: id,
        type: flowType.Send,
        currentRoute: route('/flows/:flowId/send/to', {
          flowId: id
        }),
        data: {},
        steps: [
          {
            name: 'To',
            route: route('/flows/:flowId/send/to', {
              flowId: id
            })
          },
          {
            name: 'Amount',
            route: route('/flows/:flowId/send/amount', { flowId: id })
          },
          {
            name: 'Review',
            route: route('/flows/:flowId/send/review', { flowId: id })
          }
        ],
        complete: false,
        defaultExitTo: '/'
      }
    case flowType.LinkedAccount:
      return {
        id,
        type: flowType.LinkedAccount,
        currentRoute: route('/flows/:flowId/linked-account/type', {
          flowId: id
        }),
        data: {},
        steps: [
          {
            name: 'Type',
            route: route('/flows/:flowId/linked-account/type', { flowId: id })
          },
          {
            name: 'Details',
            route: route('/flows/:flowId/linked-account/details', {
              flowId: id
            })
          },
          {
            name: 'Review',
            route: route('/flows/:flowId/linked-account/review', { flowId: id })
          }
        ],
        complete: false,
        defaultExitTo: route('/settings/linked-accounts')
      }
    case flowType.Signup:
      return {
        id,
        type: flowType.Signup,
        currentRoute: route('/flows/:flowId/signup/about', { flowId: id }),
        data: {},
        steps: [
          {
            name: 'About',
            route: route('/flows/:flowId/signup/about', { flowId: id })
          },
          {
            name: 'Phone',
            route: route('/flows/:flowId/signup/phone', {
              flowId: id
            })
          },
          {
            name: 'SMS',
            route: route('/flows/:flowId/signup/sms', { flowId: id })
          },
          {
            name: 'Password',
            route: route('/flows/:flowId/signup/password', { flowId: id })
          }
        ],
        complete: false,
        defaultExitTo: route('/')
      }
    case flowType.UnitOnboarding:
      return {
        id,
        type: flowType.UnitOnboarding,
        currentRoute: route('/flows/:flowId/unit-onboarding/address', {
          flowId: id
        }),
        data: {},
        steps: [
          {
            name: 'Address',
            route: route('/flows/:flowId/unit-onboarding/address', {
              flowId: id
            })
          },
          {
            name: 'About',
            route: route('/flows/:flowId/unit-onboarding/about', {
              flowId: id
            })
          }
        ],
        complete: false,
        defaultExitTo: route('/onboarding/unit')
      }
  }
}
