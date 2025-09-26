/**
 * The pattern in error.server.ts: The error wrapper function accepts a mapping parameter.
 * This allows mapping the *Frontend defined fields* to the *Backend defined fields*.
 * 
 * Ex: Frontend defines a field called "otp" and Backend defines a field called "Code".
 */

export const TwillioErrorMapper = {
    otp: "Code",
    phone: "To",
    verificationSid: "VerificationSid",
    amount: "Amount",
    payee: "Payee"
}
export type TwillioError = typeof TwillioErrorMapper
