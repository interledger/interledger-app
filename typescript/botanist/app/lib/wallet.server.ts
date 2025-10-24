import { json } from '@remix-run/node'
import { DateTime } from 'luxon'
import type {
  Features,
  GatehubUser,
  GetTransactionDetailsResponse,
  LinkedAccount,
  LinkedAccountReview,
  LinkedAccountReviews,
  ListAuditResponse,
  ListCountriesResponse,
  ListWalletsResponse,
  PaginationRequest,
  User,
  WalletDetails
} from '~/generated/protobuf-ts/backend/admin/v1/backend'
import type { Transfer } from '~/generated/protobuf-ts/backend/v1/backend'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import type { LinkedAccountReviewState } from '~/lib/types'

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE
export const KRATOS_ADMIN_URL =
  process.env.KRATOS_ADMIN_URL || 'http://kratos:4434'

export const formatAmount = (amount: number, asset: string): string => {
  if (typeof amount == 'undefined') return '$ 0.00'
  return `${asset} ${(amount / 100).toFixed(2)}`
}

export async function ListWallets(
  request: Request,
  page: PaginationRequest
): Promise<ListWalletsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listWallets(page, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export interface Wallet {
  users: User[]
  walletID: string
  firstName: string
  lastName: string
  countryCode: string
  gender: string
  dateOfBirth?: string
  address: string
  kycStatus: string
}

export async function GetWalletDetails(
  request: Request,
  walletID: string
): Promise<Wallet> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return {
    ...response.response,
    gender:
      response.response.gender == 0
        ? 'Unknown'
        : response.response.gender == 1
        ? 'Male'
        : response.response.gender == 2
        ? 'Female'
        : 'Other',
    dateOfBirth: DateTime.fromSeconds(
      parseInt(response.response.dateOfBirth?.seconds ?? '')
    ).toLocaleString(DateTime.DATETIME_FULL)
  }
}

export async function GetWalletFeatures(
  request: Request,
  walletID: string
): Promise<Features> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletFeatures(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function SetWalletFeatures(
  request: Request,
  features: Features
): Promise<Features> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .setWalletFeatures(features, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

interface Transaction {
  walletID: string
  id: string
  type: string
  status: string
  date: string
  amount: string
  source: string
  destination: string
  transfers?: Transfer[]
}

export async function GetWalletTransactions(
  request: Request,
  walletID: string
): Promise<Transaction[]> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listTransactions(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  console.log('RESPONSE', response)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response.transactions.map((transaction) => {
    return {
      walletID,
      id: transaction.id,
      status: '???',
      type: transaction.type,
      date: DateTime.fromSeconds(
        parseInt(transaction.timestamp?.seconds ?? '')
      ).toLocaleString(DateTime.DATETIME_FULL),
      amount: formatAmount(transaction.amount, transaction.asset),
      source: transaction.source,
      destination: transaction.destination
    }
  })
}

export async function GetWalletTransactionDetails(
  request: Request,
  walletID: string,
  transactionID: string
): Promise<GetTransactionDetailsResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getTransactionDetails(
      { walletID, transactionID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

/**
 * RPC endpoints required for the admin panel
 *
 * - Transactions
 *     - List
 *     - Get detailed
 * - Identities
 *     - List
 *     - Get Detailed
 * - KYC Status
 * - Linked Accounts
 *     - List
 *     - Detailed
 * - Audit
 *     - List
 */

export async function GetWalletIdentities(
  request: Request,
  walletID: string
): Promise<WalletDetails> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetWalletIdentityDetails(
  request: Request,
  walletID: string
): Promise<WalletDetails> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetWalletLinkedAccounts(
  request: Request,
  walletID: string
): Promise<WalletDetails> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetWalletLinkedAccountDetails(
  request: Request,
  walletID: string
): Promise<WalletDetails> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getWalletDetails(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetWalletAudits(
  request: Request,
  walletID: string
): Promise<ListAuditResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listAudit(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function ListLinkedAccounts(
  request: Request,
  walletID: string
): Promise<LinkedAccount[]> {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .listLinkedAccounts(
      {
        walletID
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response.accounts
}

export async function ListLinkedAccountReviews(
  request: Request,
  page: PaginationRequest
): Promise<LinkedAccountReviews> {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .listIncompleteLinkedAccountReviews(page, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function GetLinkedAccount(
  request: Request,
  linkedAccountID: string
): Promise<LinkedAccount> {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .getLinkedAccount(
      { id: linkedAccountID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function GetReview(
  request: Request,
  reviewID: string
): Promise<LinkedAccountReview> {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .getLinkedAccountReview(
      { id: reviewID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function CompleteLinkedAccountReview(
  request: Request,
  reviewID: string,
  newState: LinkedAccountReviewState,
  reason: string
) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .completeLinkedAccountReview(
      {
        id: reviewID,
        newState: newState,
        reason: 'Manually approved.'
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }
}

export async function ListFormSubmissionCounts(
  request: Request,
  page: PaginationRequest
) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .listFormSubmissionCounts(page, {
      meta: {
        cookies: cookie || ''
      }
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response.formSubmissionCounts
}

export async function ExportDynamicFormResult(
  request: Request,
  formId: string
) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = grpcClient.exportFormSubmissions(
    { formId },
    {
      meta: {
        cookies: cookie || ''
      }
    }
  )

  const chunks = []
  try {
    for await (let message of rpc.responses) {
      chunks.push(message.chunk)
    }
  } catch (err) {
    if (isGrpcError(rpc) && rpc.message !== 'EOF') {
      throw json({}, httpMapping(rpc.code))
    }
  }

  const blob = new Blob(chunks, { type: 'text/csv' })

  return blob
}

export async function ListFormSubmissions(request: Request, formId: string) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .listFormSubmissions(
      {
        formId
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response.formSubmissions.map((submission) => {
    return {
      id: submission.id,
      formId: submission.formId,
      date: DateTime.fromSeconds(
        parseInt(submission.timestamp?.seconds ?? '')
      ).toLocaleString(DateTime.DATETIME_FULL)
    }
  })
}

export async function GetFormSubmissionDetails(request: Request, id: string) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .getFormSubmissionDetails(
      {
        id
      },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function ListExternalApiCalls(
  request: Request,
  paymentId: string
) {
  const cookie = String(request.headers.get('cookie'))
  let rpc = await grpcClient
    .listExternalApiCalls(
      { paymentId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response.list
}

export async function GetPendingPayouts(request: Request) {
  let rpc = await grpcClient
    .listPaymentsAwaitingSignal(
      {},
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response.payments
}

export async function GetXagoWalletBalance(request: Request, walletId: string) {
  let rpc = await grpcClient
    .getWalletXagoBalance(
      {
        walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    if (rpc.code == Code.NOT_FOUND) {
      return null
    }
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function EnableXagoWalletBalance(
  request: Request,
  walletId: string
) {
  let rpc = await grpcClient
    .setWalletXagoBalanceEnabled(
      {
        walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function setWalletCountry(
  request: Request,
  walletId: string,
  country: string
) {
  let rpc = await grpcClient
    .setWalletCountry(
      {
        id: walletId,
        countryCode: country
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function GetPtiWalletBalance(request: Request, walletId: string) {
  let rpc = await grpcClient
    .getPTIBalance(
      {
        walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    if (rpc.code == Code.NOT_FOUND) {
      return null
    }
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function GetGatehubWalletBalance(
  request: Request,
  walletId: string
) {
  let rpc = await grpcClient
    .getGatehubBalance(
      {
        walletID: walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    if (rpc.code == Code.NOT_FOUND) {
      return null
    }
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function EnablGatehubBalance(request: Request, walletId: string) {
  let rpc = await grpcClient
    .createGatehubUser(
      {
        walletID: walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function EnablePtiWalletBalance(
  request: Request,
  walletId: string
) {
  let rpc = await grpcClient
    .enablePTIBalance(
      {
        walletId
      },
      { meta: { cookies: String(request.headers.get('cookie')) } }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function ListCountries(
  request: Request
): Promise<ListCountriesResponse> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .listCountries(
      {},
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function GetGatehubUser(
  request: Request,
  walletID: string
): Promise<GatehubUser> {
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .getGatehubUser(
      { walletID },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}

export async function CheckUserTotpEnabled(
  request: Request,
  identityId: string,
  walletId: string
): Promise<boolean> {
  const cookie = String(request.headers.get('cookie'))
  const response = await grpcClient
    .checkUserTotpEnabled(
      { identityId, walletID: walletId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response.isEnabled
}

export async function DeleteUserTotp(
  request: Request,
  identityId: string,
  walletId: string
): Promise<any> {
  console.log('Deleting TOTP enrollment for identity:', identityId)
  const cookie = String(request.headers.get('cookie'))
  let response = await grpcClient
    .delete2FATotpEnrollment(
      { identityId, walletID: walletId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return response.response
}
