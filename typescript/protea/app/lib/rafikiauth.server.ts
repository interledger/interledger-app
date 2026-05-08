import { data } from 'react-router';
import { envValue } from '~/env.server';

const RAFIKI_AUTH_ENDPOINT =
  envValue("RAFIKI_AUTH_ENDPOINT")
const RAFIKI_AUTH_SECRET = envValue("RAFIKI_AUTH_SECRET")

export type GrantDetails = {
  access: Access[]
}

export type Access = {
  type: string
  actions: AccessAction[]
  identifier?: string
  limits?: Limits
}

export type Amount = {
  value: string
  assetScale: number
  assetCode: string
}

export type Limits = {
  receiver?: string
  debitAmount?: Amount
  receiveAmount?: Amount
  interval?: string
}

type AccessAction =
  | 'create'
  | 'complete'
  | 'read'
  | 'read-all'
  | 'list'
  | 'list-all'

export async function getInteraction(
  interactionId: string,
  nonce: string
): Promise<Access[]> {
  let rpc = await fetch(
    `${RAFIKI_AUTH_ENDPOINT}/grant/${interactionId}/${nonce}`,
    {
      headers: { 'x-idp-secret': RAFIKI_AUTH_SECRET }
    }
  )
  if (rpc.status > 300) {
    throw data({}, rpc.status)
  }
  let body = (await rpc.json()) as GrantDetails

  return body.access
}

export async function consent(
  interactionId: string,
  nonce: string,
  userDecision: 'accept' | 'reject'
): Promise<void> {
  let rpc = await fetch(
    `${RAFIKI_AUTH_ENDPOINT}/grant/${interactionId}/${nonce}/${userDecision}`,
    {
      body: JSON.stringify({}),
      method: 'POST',
      headers: { 'x-idp-secret': RAFIKI_AUTH_SECRET }
    }
  )

  if (rpc.status != 202) {
    throw data({}, rpc.status)
  }
}
