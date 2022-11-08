import {httpMapping, isGrpcError, openPaymentsClient, StatusError} from "~/lib/proto.server";
import {json} from "@remix-run/node";
import {PaymentPointer} from "~/generated/protobuf-ts/backend/v1/backend";

export const PAYMENT_POINTER_BASE = process.env.PAYMENT_POINTER_BASE

export async function getWalletPaymentPointer(request: Request): Promise<PaymentPointer> {
  const cookie = String(request.headers.get('cookie'))
  let response = await openPaymentsClient
    .listWalletPaymentPointers(
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

  return response.response.pointers[0]
}