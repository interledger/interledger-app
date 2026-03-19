import { href, redirect } from 'react-router'
import { redirectWithSnackbar } from '../snackbar.server'
import { BffApiError, isClientError } from './bff-error'

/**
 * Higher-order function to intercept a client method call and log errors.
 * `request` is captured from the outer `createClient(request)` closure —
 * no need to pass it into every method, same pattern as grpc.server.ts.
 */
function withInterceptor<T extends (...args: any[]) => ReturnType<T> | typeof BffApiError>(
    name: string,
    request: Request,
    fn: T
): T {
    return ((...args: Parameters<T>) => {
        let result;
        try {
            result = fn(...args)
        } catch (err) {
            const error = err as any
            console.error(`🆘🆘🆘[Interceptor] method "${name}" received error:`, error)

            if (error.cause == 401) {
                throw redirect(href('/settings'))
            }

            if (error.cause == 500) {
                throw error
            }

            console.error(`🆘🆘🆘[Interceptor] method "${name}" returning client error:`)
            return BffApiError(error.message)
        }

        return result
    }) as T
}

/**
 * Factory that creates a dummyClient bound to the current request.
 * Mirrors the `createPromiseClient(service, transport)` pattern in grpc.server.ts —
 * `request` is injected once and shared across all methods via closure.
 */
export function createDummyClient(request: Request) {
    const withErrorInterceptor = <T extends (...args: any[]) => any>(name: string, fn: T) =>
        withInterceptor(name, request, fn)

    return {
        getTransactions: withErrorInterceptor('getTransactions', (withError = false) => {
            if (withError) {
                throw new Error('Couldnt get transactions', { cause: 500 })
            }

            return [
                { id: 1, details: 'transaction 1' },
                { id: 2, details: 'transaction 2' }
            ]
        }),

        submitSuccesful: withErrorInterceptor('submitTransaction1', (data: any) => {
            return { message: 'Transaction 1 submitted successfully' }
        }),

        submit401: withErrorInterceptor('submit401', (data: any) => {
            throw new Error('400: Transaction 2 failed: Invalid amount', { cause: 401 })
        }),

        submit403: withErrorInterceptor('submit403', (data: any) => {
            throw new Error('403: Transaction 3 failed: Invalid amount', { cause: 403 })
        }),

        isError: isClientError
    }
}
