import { RequestConfig } from "./types.server";

type ResponseWithCookies = {
    headers?: { 'set-cookie'?: string | string[]; [key: string]: unknown }
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
    const setCookie = response.headers?.['set-cookie']
    if (!setCookie) return []
    if (Array.isArray(setCookie)) return setCookie as string[]
    return typeof setCookie === 'string' ? [setCookie] : []
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
