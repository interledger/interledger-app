import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { grpcClient } from '~/lib/proto.server'
import { commitSession, getSession } from '~/session.server'

export const IS_SIGNUP_GATED = process.env.GATE_SIGNUP == 'true' || false

export async function canSignup(request: Request): Promise<void> {
  if (!IS_SIGNUP_GATED) {
    return
  }
  const url = new URL(request.url)
  const userSettings = await getSession(request.headers.get('Cookie'))

  if (userSettings.get('canSignup')) {
    return
  }

  // Allow if loop11 first time and set the session to canSignup
  if (url.searchParams.get('l11_tracking') && url.searchParams.get('l11_uid')) {
    if (userSettings.get('canSignup')) {
      return
    }

    userSettings.set('canSignup', true)

    throw redirect('/signup' + url.search, {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    })
  }

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

  throw redirect(route('/waitlist'))
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
