import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router';
import logger, { addRequestId, addCorrelationId } from './logger.server'
import {
  initializeRequestContext,
  extractOrGenerateRequestId
} from './requestContext.server'

/**
 * Wrapper for loader/action functions that initializes request context and logging
 * Automatically captures requestId from x-request-id header or generates one
 * Also handles correlation ID and initializes the request context for structured logging
 *
 * Usage:
 * export const loader = withLoggingContext(async ({ request, params, context }) => {
 *   logger.info({}, 'Processing user request')
 *   // ... rest of loader
 * })
 *
 * export const action = withLoggingContext(async ({ request, params, context }) => {
 *   logger.info({}, 'Processing form submission')
 *   // ... rest of action
 * })
 */
export function withLoggingContext<T extends any[] | Record<string, any>>(
  handler: (args: LoaderFunctionArgs | ActionFunctionArgs) => Promise<T | Response>
) {
  return async (args: LoaderFunctionArgs | ActionFunctionArgs) => {
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
