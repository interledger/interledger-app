import { data } from 'react-router'
import type { BFFApiError } from '~/lib/bff-error'
import { ErrorDescriptions } from '~/lib/error.constants'
import logger from '~/lib/logger.server'

// ─── Kratos error IDs ─────────────────────────────────────────────────────────

enum KratosErrorId {
  ErrorValidation = 4000000,
  ErrorValidationGeneric,
  ErrorValidationRequired,
  ErrorValidationMinLength,
  ErrorValidationInvalidFormat,
  ErrorValidationPasswordPolicyViolation,
  ErrorValidationInvalidCredentials,
  ErrorValidationDuplicateCredentials,
  ErrorValidationTOTPVerifierWrong,
  ErrorValidationIdentifierMissing,
  ErrorValidationAddressNotVerified,
  ErrorValidationNoTOTPDevice,
  ErrorValidationLookupAlreadyUsed,
  ErrorValidationNoWebAuthnDevice,
  ErrorValidationNoLookup,
  ErrorValidationSuchNoWebAuthnUser,
  ErrorValidationLookupInvalid,
  ErrorValidationLogin = 4010000,
  ErrorValidationLoginFlowExpired,
  ErrorValidationLoginNoStrategyFound,
  ErrorValidationRegistrationNoStrategyFound,
  ErrorValidationSettingsNoStrategyFound,
  ErrorValidationRecoveryNoStrategyFound,
  ErrorValidationVerificationNoStrategyFound,
  ErrorValidationRegistration = 4040000,
  ErrorValidationRegistrationFlowExpired,
  ErrorValidationSettings = 4050000,
  ErrorValidationSettingsFlowExpired,
  ErrorValidationRecovery = 4060000,
  ErrorValidationRecoveryRetrySuccess,
  ErrorValidationRecoveryStateFailure,
  ErrorValidationRecoveryMissingRecoveryToken,
  ErrorValidationRecoveryTokenInvalidOrAlreadyUsed,
  ErrorValidationRecoveryFlowExpired,
  ErrorValidationRecoveryCodeInvalidOrAlreadyUsed,
  ErrorValidationVerification = 4070000,
  ErrorValidationVerificationTokenInvalidOrAlreadyUsed,
  ErrorValidationVerificationRetrySuccess,
  ErrorValidationVerificationStateFailure,
  ErrorValidationVerificationMissingVerificationToken,
  ErrorValidationVerificationFlowExpired,
  ErrorValidationVerificationCodeInvalidOrAlreadyUsed,
  ErrorSystem = 5000000,
  ErrorSystemGeneric
}

type KratosMessage = {
  id: KratosErrorId
  text: string
}

function resolveKratosMessage(msg: KratosMessage): string {
  switch (msg.id) {
    case KratosErrorId.ErrorValidationInvalidCredentials:
      return 'The provided credentials are invalid.'
    case KratosErrorId.ErrorValidationDuplicateCredentials:
      return 'An account with the same identifier (email, phone, username, ...) exists already.'
    default:
      return msg.text
  }
}

// ─── Public API ───────────────────────────────────────────────────────────────

function parseKratosBody(data: unknown): { message: string; formErrors: Record<string, string> } {
  const formErrors: Record<string, string> = {}
  let message = ErrorDescriptions.DEFAULT

  const body = data as Record<string, any>

  // Field-level errors from ui.nodes
  if (Array.isArray(body.ui?.nodes)) {
    for (const node of body.ui.nodes) {
      if (node.messages?.length > 0) {
        const fieldName: string = node.attributes?.name
        if (!fieldName || fieldName === 'csrf_token') continue
        const errorMsg = resolveKratosMessage(node.messages[0])
        logger.warn({ field: fieldName, message: errorMsg }, 'Kratos field validation error')
        formErrors[fieldName] = errorMsg
      }
    }
  }

  // Form-level message (shown in snackbar or banner)
  if (body.ui?.messages?.length > 0) {
    message = resolveKratosMessage(body.ui.messages[0])
  } else if (Object.keys(formErrors).length > 0) {
    message = 'Please correct the highlighted fields.'
  }

  return { message, formErrors }
}

/**
 * Converts a Kratos HTTP error response into a BFFApiError.
 *
 * Parses `ui.nodes` for field-level validation errors and
 * `ui.messages` for form-level messages.
 *
 * @param response - The raw fetch Response from a Kratos API call (status >= 400).
 */
export async function fromKratosResponse(response: Response): Promise<BFFApiError> {
  try {
    const body = await response.json()
    const { message, formErrors } = parseKratosBody(body)
    return {
      message,
      formErrors: Object.keys(formErrors).length > 0 ? formErrors : undefined,
      status: response.status >= 400 ? response.status : 400,
    }
  } catch (e) {
    logger.error({ status: response.status }, 'Failed to parse Kratos error response as JSON')
    return { message: ErrorDescriptions.DEFAULT, status: response.status >= 400 ? response.status : 400 }
  }
}

/**
 * Converts already-parsed Kratos error JSON into a BFFApiError.
 * Use this when the response body has already been consumed (e.g. by a prior `.json()` call).
 *
 * @param kratosData - Already-parsed JSON body from a Kratos error response.
 * @param status - The HTTP status code from the response.
 */
export function fromKratosData(kratosData: unknown, status: number): BFFApiError {
  const { message, formErrors } = parseKratosBody(kratosData)

  return {
    message,
    formErrors: Object.keys(formErrors).length > 0 ? formErrors : undefined,
    status: status >= 400 ? status : 400,
  }
}

/**
 * Wraps a BFFApiError as a React Router data response with the correct HTTP status code.
 * Use this as the return value from a loader or action when you want to send the error
 * back to the client as inline data (i.e. without redirecting or triggering an error boundary).
 */
export function sendBffError(err: BFFApiError) {
  return data(err, { status: err.status })
}
