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
  PaymentMethod = 'payment-method'
}

type ProgressStep = {
  name: string
  route: string
}

type Flow = {
  id: string
  type: flowType
  name: string
  data?: any
  stepIndex: number
  steps: ProgressStep[]
  complete: boolean
  exitTo: string
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
      throw redirect(currentFlow.steps[currentFlow.stepIndex].route, {
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
        throw json({
          title: 'Your flow is invalid',
          body: 'Expected specific flow type, but found something else.'
        })
      flows.push(currentFlow)
    }

    userSettings.set('flows', flows)
    throw redirect(currentFlow.steps[currentFlow.stepIndex].route, {
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
 * Allows adding data to a flows data and incrementing the stepIndex.
 * Can be used to move backwards within a flow.
 * @param request
 * @param data A data object to me merged into the flow data
 * @param newIndex Optional index to route to
 * @returns Flow
 */
export async function stepFlow(
  request: Request,
  data?: any,
  newIndex?: number
): Promise<Flow | undefined> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const currentFlow = flows.pop()
  if (currentFlow) {
    if (data) {
      currentFlow.data = Object.assign(currentFlow.data, data)
      currentFlow.stepIndex++
    } else if (newIndex == -1) {
      currentFlow.stepIndex--
    } else if (newIndex) {
      currentFlow.stepIndex = newIndex
    }
    flows.push(currentFlow)
    userSettings.set('flows', flows)
    throw redirect(currentFlow.steps[currentFlow.stepIndex].route, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }
  return currentFlow
}

/**
 * Allows adding data to a flows data without incrementing the stepIndex.
 * @param request
 * @param data A data object to me merged into the flow data
 * @returns Flow
 */
export async function updateFlowData(
  request: Request,
  data?: any
): Promise<Response | undefined> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const currentFlow = flows.pop()
  if (currentFlow) {
    if (data) {
      currentFlow.data = Object.assign(currentFlow.data, data)
    }
    flows.push(currentFlow)
    userSettings.set('flows', flows)
    return json(data, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }
  return currentFlow
}

/**
 * Allows setting a flow as complete
 * @param request Request
 * @returns Flow
 */
export async function completeFlow(
  request: Request
): Promise<Flow | undefined> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const latestFlow = flows.pop()
  if (latestFlow) {
    latestFlow.complete = true
    flows.push(latestFlow)
    userSettings.set('flows', flows)
    throw redirect(`/confirmation/${latestFlow.id}/${latestFlow.type}`, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }
  return latestFlow
}

/**
 * Allows removing a flow from the stack and returning to the next flow in the stack,
 * Or returns to the flows exitTo route.
 * @param request Request
 * @returns Response
 */
export async function exitFlow(request: Request): Promise<Response> {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const flows: Flow[] = userSettings.get('flows') || []
  const lastFlow = flows.pop()
  const latestFlow = flows.at(-1)
  userSettings.set('flows', flows)

  if (latestFlow) {
    return redirect(latestFlow.steps[latestFlow.stepIndex].route, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  } else if (lastFlow) {
    return redirect(lastFlow?.exitTo, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }
  throw json({
    title: 'Your flow has expired',
    body: 'Your flow timed out, please restart it to continue.',
    action: { text: 'Restart flow' }
  })
}

const flowTemplate = (id: string, type: string): Flow | undefined => {
  switch (type) {
    case flowType.Deposit:
      return {
        id: id,
        type: flowType.Deposit,
        name: 'Deposit',
        stepIndex: 0,
        data: {},
        steps: [
          {
            name: 'From',
            route: route('/flows/:flowId/deposit/payment-method', {
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
        exitTo: '/home'
      }
    case flowType.Withdraw:
      return {
        id: id,
        type: flowType.Withdraw,
        name: 'Withdraw',
        stepIndex: 0,
        data: {},
        steps: [
          {
            name: 'To',
            route: route('/flows/:flowId/withdraw/payment-method', {
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
        exitTo: '/home'
      }
    case flowType.Send:
      return {
        id: id,
        type: flowType.Send,
        name: 'Send',
        stepIndex: 0,
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
        exitTo: '/home'
      }
    case flowType.PaymentMethod:
      return {
        id,
        type: flowType.PaymentMethod,
        name: 'Payment method',
        stepIndex: 0,
        data: {},
        steps: [
          {
            name: 'Type',
            route: route('/flows/:flowId/payment-method/type', { flowId: id })
          },
          {
            name: 'Details',
            route: route('/flows/:flowId/payment-method/details', {
              flowId: id
            })
          },
          {
            name: 'Review',
            route: route('/flows/:flowId/payment-method/review', { flowId: id })
          }
        ],
        complete: false,
        exitTo: '/settings/payment-methods'
      }
  }
}
