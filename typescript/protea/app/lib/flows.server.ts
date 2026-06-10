import { data, href, redirect } from 'react-router'
import { v4 } from 'uuid'
import { commitSession, getSession } from '~/session.server'

export enum flowType {
  LinkCardAccount = 'linkCard',
  LinkBankAccount = 'linkBank',
  PersonalDetails = 'personalDetails',
  PasswordChallenge = 'passwordChallenge',
  TopUp = 'topUp',
  Withdraw = 'withdraw'
}

type Flow = {
  data?: any
  startRoute: string
  returnTo: string
}

type Flows = {
  [K in flowType]: Flow | null
}

/**
 * Sets up the flow object.
 * Will throw a redirect to the newly created flow, if one of the specified type doesn't exist yet.
 *
 * @param request Request received in a loader function.
 * @param type The flow type used to identify the flow.
 * @param template A dynamic runtime flow template to use to override the templates defined here.
 * @returns returns the current flow of type flowType.
 */
export async function requireFlow(
  request: Request,
  type: flowType,
  template?: Flow
): Promise<Flow> {
  const session = await getSession(request.headers.get('Cookie'))

  const flows: Flows = session.get('flows') ?? {
    pay: null,
    linkCard: null,
    linkBank: null,
    personalDetails: null,
    passwordChallenge: null
  }

  let currentFlow = flows[type]

  if (currentFlow == null) {
    currentFlow = template ?? flowTemplate(type)
    flows[type] = currentFlow

    session.set('flows', flows)
    // Ensure qp are preserved for loop11
    const url = new URL(request.url)
    throw redirect(currentFlow.startRoute + url.search, {
      headers: {
        'Set-Cookie': await commitSession(session)
      }
    })
  }

  return currentFlow
}

/**
 * Allows updating a flow's data.
 * @param request
 * @param type The flow type used to identify the flow.
 * @param data A data object to me merged into the flow data.
 * @returns string The set cookie header to commit the session.
 */
export async function updateFlow(
  request: Request,
  type: flowType,
  data?: any
): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))

  const flows: Flows = session.get('flows')
  const currentFlow = flows[type]

  if (currentFlow) {
    if (data) {
      Object.assign(currentFlow.data, data)
    }
    flows[type] = currentFlow
    session.set('flows', flows)

    return commitSession(session)
  }
  throw data(
    {
      title: "Can't update a flow that doesn't exist."
    },
    { status: 400, statusText: 'Bad Request' }
  )
}

/**
 * Allows removing a flow from the stack and returning to the next flow in the stack,
 * Or returns headers allowing the consumer to redirect out of the flow.
 * @param request Request
 * @param type The flow type used to identify the flow.
 * @returns string The set cookie header to commit the session
 */
export async function exitFlow(
  request: Request,
  type: flowType
): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))
  const flows: Flows = session.get('flows')

  flows[type] = null
  session.set('flows', flows)

  return commitSession(session)
}

const flowTemplate = (type: flowType): Flow => {
  switch (type) {
    case flowType.LinkCardAccount:
      return {
        startRoute: href('/connect/card'),
        data: {},
        returnTo: href('/accounts')
      }
    case flowType.PersonalDetails:
      return {
        startRoute: href('/personal-details'),
        data: {
          idempotencyKey: v4()
        },
        returnTo: href('/')
      }
    case flowType.PasswordChallenge:
      return {
        startRoute: href('/login/challenge'),
        data: {},
        returnTo: href('/settings/password')
      }
    default:
      throw data(
        {
          title: "You specified a flow type that doesn't exist"
        },
        { status: 400, statusText: 'Bad Request' }
      )
  }
}
