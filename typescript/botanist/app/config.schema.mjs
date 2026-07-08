// Plain JS (not .ts) so this schema can be imported directly by bootstrap.mjs
// (a vanilla Node ESM script, never processed by Vite/TypeScript).
import { z } from 'zod'

export const BotanistConfigSchema = z.object({
  backend_grpc_url: z.string().min(1),
  payment_pointer_base: z.string().min(1)
})
