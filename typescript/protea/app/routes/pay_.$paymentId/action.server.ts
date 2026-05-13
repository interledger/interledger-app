import { Code } from '@bufbuild/connect'
import type { ActionFunctionArgs } from 'react-router';
import { data } from 'react-router';
import { href } from 'react-router'
import { stringToBigInt } from '~/lib/amount'
import {  validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getClientIP } from '~/lib/ip.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function confirmPaymentAction({
  request,
  params
}: ActionFunctionArgs) {
  const form = await request.formData()
  const serviceAgreement = form.get('serviceAgreement') as string
  // const otp = String(form.get('otp') || '')

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    otp: '',
    serviceAgreement: ''
  }

  if (serviceAgreement == null) {
    errors.serviceAgreement = 'You are required to authorize to continue.'
    return error(request, { errors })
  }

  const clientIpAddress = getClientIP(request)

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    ipAddress: clientIpAddress
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  response = await grpc.confirmPayment(request, {
    id: params.paymentId
  })
  if (isConnectError(response)) {
    return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, href('/'), {
    message: 'Payment created successfully.',
    icon: 'close'
  })
}

export async function updatePaymentAction({
  request,
  params
}: ActionFunctionArgs) {
  const form = await request.formData()
  const send = String(form.get('send') || '')
  const receive = String(form.get('receive') || '')
  const note = String(form.get('note') || '')
  const accountId = String(form.get('accountId') || '')
  const sendCurrency = String(form.get('sendCurrency') || '')
  const receiveCurrency = String(form.get('receiveCurrency') || '')
  const intent = form.get('intent') as string

  const sendToSubmit = stringToBigInt(send)
  const receiveToSubmit = stringToBigInt(receive)

  const errors = {
    form: '',
    amount: '',
    linkedAccount: '',
    note: ''
  }

  if (intent == 'submit' && sendToSubmit == 0n) {
    errors.amount = 'Amount is required.'
    return error(request, { errors, payment: null, intent: '' })
  }

  const clientIpAddress = getClientIP(request)

  let senderAmount, receiverAmount
  if (send != '') {
    senderAmount = {
      amount: sendToSubmit,
      assetScale: 2,
      asset: sendCurrency
    }
  }
  if (receive != '') {
    receiverAmount = {
      amount: receiveToSubmit,
      assetScale: 2,
      asset: receiveCurrency
    }
  }

  let response = await grpc.updatePayment(request, {
    id: params.paymentId,
    note,
    senderAccount: accountId,
    senderAmount,
    receiverAmount,
    ipAddress: clientIpAddress
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors, payment: null, intent: '' })
    }
    if (
      response.code == Code.FailedPrecondition &&
      response.violations.findIndex(
        (violation) =>
          violation.type === 'Payment' &&
          violation.subject === 'insufficientFunds'
      ) > -1
    ) {
      return response.error({
        errors: { ...errors, amount: 'You have insufficient funds available.' },
        payment: null,
        intent: ''
      })
    }
    return response.error(
      { errors, payment: null, intent: '' },
      {},
      { action: 'Contact support' }
    )
  }

  return data({
    payment: response,
    intent,
    errors
  })
}
