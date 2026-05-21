import logger from './lib/logger.server'

const knownEnvKeysRequired: string[] = [
    "COOKIE_SECRETS",
    "BACKEND_GRPC_URL",
    "BACKEND_HTTP_URL",
    "NODE_ENV",
    "REDIS_URL",
    "KRATOS_URL",
    "RAFIKI_AUTH_ENDPOINT",
    "RAFIKI_AUTH_SECRET",
    "PUBLIC_OP_AUTH_HOST",
    "PTI_FORMS_URL",
    "PTI_SDK_URL",
    "PERSONA_SDK_URL",
    "MOCKXAGO_ENDPOINT",
    "TARGET_HOST",
    "SUPPORT_EMAIL",
    "PAYMENT_POINTER_BASE",
    "PTI_CLIENT_ID"
]

const knownEnvKeysOptional: string[] = [
    "SENTRY_DSN",
    "SENTRY_RELEASE",
    "SENTRY_ENV_LABEL",
    "FYNBOS_ENV",
    "LOG_LEVEL",
    "LOG_PRETTY",
    "PUSHER_APP_KEY",
    "PUSHER_APP_CLUSTER",
    "DEFAULT_RATE_LIMIT_REQUESTS",
    "DEFAULT_RATE_LIMIT_TIME",
    "GOOGLE_MAPS_API_KEY"
]

const knownEnvKeysEnabled = {
    "OP_INTPAY_ENABLED": [
        "OP_INTPAY_KEY_ID",
        "OP_INTPAY_PRIVATE_KEY",
        "OP_INTPAY_WALLET_ADDRESS",
        "OP_INTPAY_REDIRECT_URL",
        "OP_INTPAY_HOST"],
}

export function envVarValidation() {
    try {
        const missingRequired: string[] = []
        const missingOptional: string[] = []

        for (const envKey of knownEnvKeysOptional) {
            requireEnv(envKey, missingOptional)
        }

        for (const [flagEnabled, requiredVars] of Object.entries(knownEnvKeysEnabled)) {

            // Validation for variables required by Interledger Pay
            if (envBool(flagEnabled)) {
                for (const envKey of requiredVars) {
                    requireEnv(envKey, missingRequired)
                }
            }
        }

        for (const envKey of knownEnvKeysRequired) {
            requireEnv(envKey, missingRequired)
        }

        if (missingOptional.length > 0) {
            console.warn(
                `WARNING: Missing optional environment variables: ${missingOptional.join(", ")}`
            )
        }

        if (missingRequired.length > 0) {
            throw new Error(
                `Missing required environment variables: ${missingRequired.join(", ")}`
            )
        }
    } catch (err) {
        logger.error(err)
        process.exit(1)
    }
}

// Checks variable and records it if missing
function requireEnv(name: string, missing: string[]): string | undefined {
    const value = process.env[name]

    if (!value) {
        missing.push(name)
        return undefined
    }
    return value
}

export function envBool(name: string): boolean {
    const value = process.env[name]?.trim().toLowerCase()
    if (value === undefined) return false

    if (value !== "true" && value !== "false") {
        throw new Error(`Invalid boolean env var ${name}: ${value}`)
    }

    return value === "true"
}

export const envValue = (key: string) => {
    const allKeys = [
        ...knownEnvKeysRequired,
        ...knownEnvKeysOptional,
        ...Object.values(knownEnvKeysEnabled).flat()
    ]

    if (!allKeys.includes(key)) {
        throw new Error(`Unknown env variable: ${key}`)
    }

    const value = process.env[key]

    if (knownEnvKeysRequired.includes(key) && !value) {
        throw new Error(`Env variable value is required: ${key}`)
    }

    if (!value) {
        return ""
    }
    return value
}
