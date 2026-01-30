import pino, { Logger as PinoLogger, LoggerOptions } from 'pino'
import { getRequestId, getCorrelationId } from './requestContext.server'

declare global {
  var logger: PinoLogger
}

// Valid log levels that match the logging policy
const VALID_LOG_LEVELS = ['fatal', 'error', 'warn', 'info', 'debug', 'trace']

// Get LOG_LEVEL from environment, default to 'info'
function getLogLevel(): string {
  const logLevel = process.env.LOG_LEVEL || 'info'

  // Validate LOG_LEVEL - exit if invalid
  if (!VALID_LOG_LEVELS.includes(logLevel)) {
    // Use stderr for fatal errors
    process.stderr.write(
      JSON.stringify({
        level: 'fatal',
        ts: Date.now() / 1000,
        caller: 'logger.server.ts:24',
        msg: 'Invalid LOG_LEVEL configuration',
        error: `LOG_LEVEL must be one of: ${VALID_LOG_LEVELS.join(', ')}`,
        providedValue: logLevel,
      }) + '\n'
    )
    process.exit(1)
  }

  return logLevel
}

// Pino configuration following the logging policy
function getPinoConfig(): LoggerOptions {
  const logLevel = getLogLevel()
  const isDevelopment = process.env.NODE_ENV === 'development'

  const config: LoggerOptions = {
    level: logLevel,
    timestamp: pino.stdTimeFunctions.isoTime, // ISO format timestamp
    // Use different transports for different log levels as per policy
    transport:
      isDevelopment && process.env.LOG_PRETTY !== 'false'
        ? {
            target: 'pino-pretty',
            options: {
              colorize: true,
              singleLine: false,
              translateTime: 'SYS:standard',
              ignore: 'pid,hostname',
            },
          }
        : undefined,
    formatters: {
      level: (label) => {
        return { level: label }
      },
    },
    hooks: {
      logMethod: (args, method, level) => {
        // Route different levels to appropriate streams
        // fatal and error -> stderr
        // warning, info, debug -> stdout
        if (level <= 40) {
          // 40 is 'error' level in pino (50 is 'fatal')
          return method.apply(process.stderr, args)
        }
        return method.apply(process.stdout, args)
      },
    },
    base: {
      // Remove the default pid and hostname for cleaner logs
      pid: undefined,
      hostname: undefined,
    },
  }

  return config
}

// Create or retrieve the global logger instance
export function getLogger(): PinoLogger {
  if (global.logger) {
    return global.logger
  }

  const config = getPinoConfig()
  const logger = pino(config)

  // Store in global for reuse
  global.logger = logger

  return logger
}

// Create a child logger with additional context (useful for request-scoped logging)
export function createChildLogger(
  parentLogger: PinoLogger,
  context: Record<string, any>
): PinoLogger {
  return parentLogger.child(context)
}

// Default export for convenience
export default getLogger()

/**
 * Helper to add requestId to log context
 * Usage: logger.info({ ...addRequestId(requestId) }, 'message')
 */
export function addRequestId(requestId?: string) {
  const id = requestId || getRequestId()
  return id && id !== 'unknown' ? { requestId: id } : {}
}

/**
 * Helper to add correlation ID to log context
 * Usage: logger.info({ ...addCorrelationId(correlationId) }, 'message')
 */
export function addCorrelationId(correlationId?: string) {
  const id = correlationId || getCorrelationId()
  return id ? { correlationId: id } : {}
}
