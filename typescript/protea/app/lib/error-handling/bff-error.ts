// Error to be used inside Bff loaders/actions
export const BffApiError = (message: string, status = 500) => {
    return { message, status }
}

/**
type Fields = {
  email: string
  password: string
}

type LoginError = UserFacingError<Fields>
// => { errors: Record<'email' | 'password', string> }
 */
export type UserFacingError<T extends Record<string, string> = any> = {
    errors: Partial<Record<keyof T, string>>
    status: number
}

// Utility function to be used in Bff to check for errors in client responses
export const isClientError = (data: any): data is typeof BffApiError => {
    return data && typeof data === 'object' && 'status' in data && 'message' in data
}

// type Client = string
// type ErrorMappingFn = {
//     toUserFacingError: (data: any) => UserFacingError
// }

// export const ErrorMapper: Record<Client, ErrorMappingFn> = {
//     dummyClient: {
//         toUserFacingError: (data: ClientErrorType) => ({
//             errors: {
//                 transaction: data.message,
//                 status: data.status
//             }
//         })
//     }
//     // kratosClient
//     // grpcClient
// }
