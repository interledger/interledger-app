import type { PartialMessage } from '@bufbuild/protobuf'
import { href, redirect } from 'react-router'
import type {
  Features,
  GetPublicWalletDetailsResponse,
  KYCStatusResponse,
  ListTransactionsResponse,
  PaginationRequest,
  PublicWalletInfo,
  WalletInfo
} from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export async function getKycStatus(
  request: Request
): Promise<KYCStatusResponse> {
  const response = await grpc.kYCStatus(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getFeatures(request: Request): Promise<Features> {
  const response = await grpc.listFeatures(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getWalletId(request: Request): Promise<string> {
  let response = await grpc.getCurrentWallet(request, {})

  if (isConnectError(response)) throw response.errorResponse

  return response.id
}

export async function getPublicWalletDetails(
  request: Request,
  id: string
): Promise<GetPublicWalletDetailsResponse> {
  let response = await grpc.getPublicWalletDetails(request, {
    id
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getPublicWalletInfo(
  request: Request,
  walletAddress: string
): Promise<PublicWalletInfo> {
  let response = await grpc.getPublicWalletInfo(request, {
    walletAddress
  })

  if (isConnectError(response)) throw response.errorResponse

  return response
}

export async function getWalletInfo(request: Request): Promise<WalletInfo> {
  let response = await grpc.getWalletInfo(request, {})

  if (isConnectError(response)) {
    throw response.errorResponse
  } else if (!response.hasWalletAddress) {
    throw redirect(href('/wallet-address'))
  }

  return response
}

export type Transaction = {
  id: string
  type: string
  icon: string
  title: string
  total: string
  time: string
  date: string
}

export async function getTransactionsWithPending(
  request: Request,
  input: PartialMessage<PaginationRequest>
): Promise<ListTransactionsResponse> {
  const response = await grpc.listTransactionsWithPending(request, input)

  if (isConnectError(response)) throw response.errorResponse
  return response
}
