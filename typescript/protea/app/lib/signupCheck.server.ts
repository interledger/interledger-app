import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { commitSession, getSession } from '~/session.server'
import { grpcClient } from '~/lib/proto.server'

export const IS_SIGNUP_GATED = process.env.GATE_SIGNUP == 'true' || false

export async function canSignup(request: Request): Promise<void> {
  if (!IS_SIGNUP_GATED) {
    return
  }
  const url = new URL(request.url)
  // Allow if loop11
  if (url.searchParams.get('l11_tracking') && url.searchParams.get('l11_uid')) {
    return
  }

  const userSettings = await getSession(request.headers.get('Cookie'))

  let waitlistSignupId = userSettings.get('waitlistSignupId')
  if (waitlistSignupId) {
    // Check against service
    const canSignupCall = await grpcClient.canSignup({
      id: waitlistSignupId
    })

    if (canSignupCall.response.canSignup) {
      return
    }
  }

  //Check for QP
  waitlistSignupId = url.searchParams.get('waitlistSignupId')
  if (waitlistSignupId) {
    // Check
    const canSignupCall = await grpcClient.canSignup({
      id: waitlistSignupId
    })

    if (!canSignupCall.response.canSignup) {
      throw redirect(route('/'))
    }

    //if exists && valid redirect back to /signup with cookie
    userSettings.set('waitlistSignupId', waitlistSignupId)
    throw redirect('/signup' + url.search, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }

  throw redirect(route('/'))
}

export async function setWaitlistSignupComplete(
  request: Request,
  userId: string
): Promise<void> {
  if (!IS_SIGNUP_GATED) {
    return
  }
  const userSettings = await getSession(request.headers.get('Cookie'))

  const waitlistSignupId = userSettings.get('waitlistSignupId')
  if (waitlistSignupId) {
    await grpcClient.setSignupComplete({
      userId: userId,
      id: waitlistSignupId
    })
  }
}
