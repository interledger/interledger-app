/**
 * The standard error shape sent from the BFF to the frontend client.
 * All errors leaving the BFF (as data responses) must conform to this interface.
 */
export interface BFFApiError {
  /** Human-readable message suitable for display in a snackbar or error banner. */
  message: string
  /** Field-level validation errors. Key = form input name, value = error message.
   *  Optional — only present for validation failures (e.g. 400 with field errors). */
  formErrors?: Record<string, string>
  /** HTTP status code (e.g. 400, 404, 500). */
  status: number
}

/** Type guard: checks if a value is a BFFApiError. */
export function isBFFApiError(value: unknown): value is BFFApiError {
  return (
    typeof value === 'object' &&
    value !== null &&
    'message' in value &&
    'status' in value
  )
}
