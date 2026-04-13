import logger from './lib/logger.server'

export function envVarValidation() {
    try {
        const missing: string[] = []

        // Validation for variables required by Interledger Pay
        if (envBool("OP_INTPAY_ENABLED")) {
            requireEnv("OP_INTPAY_KEY_ID", missing)
            requireEnv("OP_INTPAY_PRIVATE_KEY", missing)
            requireEnv("OP_INTPAY_WALLET_ADDRESS", missing)
            requireEnv("OP_INTPAY_REDIRECT_URL", missing)
        }

        if (missing.length > 0) {
            throw new Error(
                `Missing required environment variables: ${missing.join(", ")}`
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
    return process.env[name]?.toLowerCase() === "true"
}