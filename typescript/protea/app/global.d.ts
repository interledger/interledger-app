/**
 * Type declarations for runtime-injected window globals.
 * These are third-party SDKs loaded via script tags, not npm packages.
 */
interface Window {
  /** Server-injected environment variables (see root.tsx) */
  ENV?: {
    sentryDsn?: string
    sentryRelease?: string
  }
  /** Cloudflare Turnstile CAPTCHA widget API */
  turnstile?: {
    reset: () => void
    render: (element: HTMLElement, options: Record<string, unknown>) => void
  }
  /** PTI (Pearsurge) KYC SDK */
  PTI?: {
    init: (options: Record<string, unknown>) => void
    form: (options: Record<string, unknown>) => void
  }
  /** Persona identity verification SDK */
  Persona?: {
    Client: new (options: {
      inquiryId?: string
      sessionToken?: string
      onReady?: () => void
      onComplete?: (data: { inquiryId: string; status: string; fields: Record<string, unknown> }) => void
      onCancel?: (data: { inquiryId: string; sessionToken: string }) => void
      onError?: (error: unknown) => void
    }) => { open: () => void }
  }
}
