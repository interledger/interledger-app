import type { RequestConfig } from "./types.server";

type ResponseWithCookies = {
    headers?: {
        'set-cookie'?: string | string[]
        [key: string]: unknown
    }
}

/**
 * Helper to extract cookie header from Request for SDK calls
 */
export function getCookie(request: Request): string {
    return request.headers.get('cookie') ?? ''
}


/**
 * Helper to extract set-cookie headers from Axios response as an array
 */
export function extractSetCookieHeaders(
    response: ResponseWithCookies
): string[] {
    const headers = response.headers
    if (!headers) return []

    // axios ≥1.14 wraps headers in AxiosHeaders where bracket access is unreliable;
    // use .get() when present, fall back to bracket notation for plain objects.
    const getMethod = headers['get']
    const raw: unknown = typeof getMethod === 'function'
        ? getMethod.call(headers, 'set-cookie')
        : headers['set-cookie']
    if (!raw) return []
    if (Array.isArray(raw)) return raw as string[]
    return typeof raw === 'string' ? [raw] : []
}


/**
 * Helper to build response headers with set-cookie from Kratos response
 */
export function buildHeadersWithCookies(
    response: ResponseWithCookies
): Headers {
    const headers = new Headers()
    const cookies = extractSetCookieHeaders(response)

    for (const cookie of cookies) {
        headers.append('Set-Cookie', cookie)
    }

    return headers
}


/**
 * Creates axios request config with cookie header for authenticated requests
 */

export function withCookie(cookie: string): RequestConfig {
    return {
        headers: {
            Cookie: cookie
        }
    }
}


/**
 * Check if a message indicates a session already exists
 */
export function isSessionAlreadyExistsMessage(msg: string): boolean {
    return (
        msg.includes('refresh=true') && msg.includes('valid session was detected')
    )
}
