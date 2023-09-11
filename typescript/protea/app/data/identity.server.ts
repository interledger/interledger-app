import { json } from '@remix-run/node'
import type { Identity } from '~/generated/connect/backend/v1/backend_pb'
import { connectClient } from '~/lib/connect.server'
import { isConnectError } from '~/lib/error.server'

export async function getIdentity(
  request: Request,
  id: string
): Promise<Identity> {
  const response = await connectClient.getIdentity(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  // TODO: remove, this should be handled by the backend
  if (!response.identity) {
    throw json({}, { status: 404, statusText: 'Not found' })
  }

  return response.identity
}

export async function getIdentityBySignatureHash(
  request: Request,
  signatureHash: string
): Promise<Identity> {
  const response = await connectClient.getIdentityBySignatureHash(request, {
    signatureHash
  })

  if (isConnectError(response)) throw response.errorResponse

  // TODO: remove, this should be handled by the backend
  if (!response.identity) {
    throw json({}, { status: 404, statusText: 'Not found' })
  } else if (response.identity.state !== 'verified') {
    throw json({}, { status: 404, statusText: 'Not found' })
  }

  return response.identity
}

type Identities = Record<string, Identity[]>

const identitiesReducer = (acc: Identities, identity: Identity) => {
  if (!acc[identity.platform]) {
    acc[identity.platform] = []
  }
  acc[identity.platform].push(identity)
  return acc
}

export async function getIdentities(request: Request): Promise<Identities> {
  const response = await connectClient.listIdentities(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response.identities.reduce(identitiesReducer, {} as Identities)
}

export async function getPublicIdentities(
  request: Request,
  walletId: string
): Promise<Identities> {
  const response = await connectClient.listPublicIdentities(request, {
    walletId
  })

  if (isConnectError(response)) throw response.errorResponse

  return response.identities.reduce(identitiesReducer, {} as Identities)
}
