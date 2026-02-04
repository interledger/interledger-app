import { json } from '@remix-run/node'
import type { ActionFunctionArgs, LoaderFunctionArgs } from '@remix-run/node'
import logger, { addRequestId, addCorrelationId } from './logger.server'
import {
  initializeRequestContext,
  extractOrGenerateRequestId
} from './requestContext.server'

/**
 * Wrapper for loader functions that initializes request context and logging
 * Automatically captures requestId from x-request-id header or generates one
 *
 * Usage:
 * export const loader = withRequestLogging(async ({ request, params, context }) => {
 *   logger.info({}, 'Processing user request')
 *   // ... rest of loader
 * })
 */
export function withRequestLogging<T extends any[] | Record<string, any>>(
  handler: (args: LoaderFunctionArgs) => Promise<T | Response>
) {
  return async (args: LoaderFunctionArgs) => {
    const requestId = extractOrGenerateRequestId(args.request)
    const correlationId = args.request.headers.get('x-correlation-id') || undefined

    return initializeRequestContext(
      () => handler(args),
      requestId,
      correlationId
    )
  }
}

/**
 * Wrapper for action functions that initializes request context and logging
 * Automatically captures requestId from x-request-id header or generates one
 *
 * Usage:
 * export const action = withRequestLogging(async ({ request, params, context }) => {
 *   logger.info({}, 'Processing form submission')
 *   // ... rest of action
 * })
 */
export function withActionLogging<T extends any[] | Record<string, any>>(
  handler: (args: ActionFunctionArgs) => Promise<T | Response>
) {
  return async (args: ActionFunctionArgs) => {
    const requestId = extractOrGenerateRequestId(args.request)
    const correlationId = args.request.headers.get('x-correlation-id') || undefined

    return initializeRequestContext(
      () => handler(args),
      requestId,
      correlationId
    )
  }
}

/**
 * Helper to log with request context automatically attached
 * Usage: logWithContext('info', { userId: '123' }, 'User logged in')
 */
export function logWithContext(
  level: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal',
  fields: Record<string, any> = {},
  message: string
): void {
  const context = {
    ...addRequestId(),
    ...addCorrelationId(),
    ...fields
  }

  logger[level](context, message)
}
