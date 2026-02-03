import { v4 as uuidv4 } from 'uuid'
import { AsyncLocalStorage } from 'async_hooks'

/**
 * RequestContext stores request-scoped data like requestId and correlationId
 * This allows us to pass these values through async operations without drilling props
 */
interface RequestContext {
  requestId: string
  correlationId?: string
}

const requestAsyncLocalStorage = new AsyncLocalStorage<RequestContext>()

/**
 * Initialize request context with requestId and optional correlationId
 * Call this in middleware or loader to set up context for the request
 */
export function initializeRequestContext(
  callback: () => Promise<any>,
  requestId?: string,
  correlationId?: string
): Promise<any> {
  const context: RequestContext = {
    requestId: requestId || uuidv4(),
    correlationId
  }

  return requestAsyncLocalStorage.run(context, callback)
}

/**
 * Get the current request context (requestId, correlationId)
 * Returns undefined if called outside of request context
 */
export function getRequestContext(): RequestContext | undefined {
  return requestAsyncLocalStorage.getStore()
}

/**
 * Get just the requestId from context
 * Falls back to 'unknown' if not in context
 */
export function getRequestId(): string {
  return getRequestContext()?.requestId || 'unknown'
}

/**
 * Get just the correlationId from context
 * Returns undefined if not set
 */
export function getCorrelationId(): string | undefined {
  return getRequestContext()?.correlationId
}

/**
 * Extract requestId from request headers or generate new UUIDv4
 */
export function extractOrGenerateRequestId(request: Request): string {
  const headerRequestId = request.headers.get('x-request-id')
  return headerRequestId || uuidv4()
}
