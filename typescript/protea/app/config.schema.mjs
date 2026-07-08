// Plain JS (not .ts) so this schema can be imported directly by server.js
// (a vanilla Node ESM entry point, never processed by Vite/TypeScript) as
// well as by config.server.ts. See config.server.ts for why the config load
// itself must happen in server.js rather than via a top-level await here.
import { z } from 'zod'

const OpIntPaySchema = z
  .object({
    enabled: z.boolean().default(false),
    key_id: z.string().default(''),
    private_key: z.string().default(''),
    wallet_address: z.string().default(''),
    host: z.string().default(''),
    redirect_url: z.string().default('')
  })
  .superRefine((op, ctx) => {
    if (!op.enabled) return
    const required = [
      ['key_id', op.key_id],
      ['private_key', op.private_key],
      ['wallet_address', op.wallet_address],
      ['host', op.host],
      ['redirect_url', op.redirect_url]
    ]
    for (const [field, value] of required) {
      if (!value) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: [field],
          message: `op_intpay.${field} is required when op_intpay.enabled is true`
        })
      }
    }
  })

export const ProteaConfigSchema = z.object({
  // Human-readable environment label surfaced to the client (e.g. "local", "sandbox", "production").
  environment: z.string().default(''),
  // A Kubernetes Secret key can only ever hold a string, so in-cluster this
  // arrives as a {{ secret ... }} resolved JSON-encoded string rather than a
  // native YAML array (which local/dev configs use directly).
  cookie_secrets: z.preprocess((val) => {
    if (typeof val === 'string') {
      try {
        return JSON.parse(val)
      } catch {
        return val
      }
    }
    return val
  }, z.array(z.string()).min(1)),
  target_host: z.string().min(1),
  support_email: z.string().min(1),
  payment_pointer_base: z.string().min(1),
  public_op_auth_host: z.string().min(1),
  persona_sdk_url: z.string().min(1),
  // Browser-accessible MockXago URL for the Persona iframe; only set in local/dev.
  mockxago_endpoint: z.string().default(''),
  log_level: z.string().default(''),
  log_pretty: z.boolean().default(true),
  google_maps_api_key: z.string().default(''),

  backend: z.object({
    grpc_url: z.string().min(1),
    http_url: z.string().min(1)
  }),
  redis: z.object({
    url: z.string().min(1)
  }),
  kratos: z.object({
    url: z.string().min(1)
  }),
  rafiki: z.object({
    auth_endpoint: z.string().min(1),
    auth_secret: z.string().min(1)
  }),
  pti: z.object({
    sdk_url: z.string().min(1),
    forms_url: z.string().min(1),
    client_id: z.string().min(1)
  }),
  sentry: z
    .object({
      dsn: z.string().default(''),
      env_label: z.string().default('')
    })
    .default({ dsn: '', env_label: '' }),
  pusher: z
    .object({
      app_key: z.string().default(''),
      app_cluster: z.string().default('eu')
    })
    .default({ app_key: '', app_cluster: 'eu' }),
  rate_limit: z
    .object({
      requests: z.number().default(4),
      window_seconds: z.number().default(3600)
    })
    .default({ requests: 4, window_seconds: 3600 }),
  op_intpay: OpIntPaySchema.default({
    enabled: false,
    key_id: '',
    private_key: '',
    wallet_address: '',
    host: '',
    redirect_url: ''
  })
})
