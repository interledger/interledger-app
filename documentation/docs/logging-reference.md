# Logging Reference

> **Logging standards for deployed services.** Consistent log formatting, levels, and guidelines for all components running on sandbox or production clusters.

**Related documents:**

- [Terminology](terminology.md) — Core vocabulary and provider translation
- [Payment Troubleshooting](payment-troubleshooting-guide.md) — Using logs to debug payment issues
- [Signup Flow](signup-guide.md) — What can and cannot be logged during registration
- [Environment Variables](env-variables.md) — `LOG_LEVEL`, OpenTelemetry, and other runtime logging config

**Quick Navigation:**

- **Which log level should I use?** → See [Log Levels](#log-levels)
- **What format should logs be in?** → See [General Rules](#general-rules)
- **What extra fields should I add?** → See [Additional Fields](#additional-fields)

---

## General Rules

- **Application logs ≠ Audit logs.** Application logs are retained 7–14 days depending on configuration; audit logs are kept indefinitely.
- All persistent components deployed to the cluster must support a `LOG_LEVEL` environment variable.
- All logs must be rendered as **JSON objects, one object per line**.
- Do not render large objects — serialisation has a performance cost.
- Do not embed conditionals inside logging statements — this makes code flaky.
- Take great care when de-referencing objects inside a logging statement — this can also make code flaky.
- If the configured `LOG_LEVEL` is not a valid option, treat it as a fatal configuration error: log a fatal-level message indicating the invalid value and terminate during startup.

## Log Levels

### `fatal`

- **Output:** `stderr` only
- **Behaviour:** Must always cause the application to exit as safely as it can. The whole process must stop.
- **Examples:**
    - Database credentials are incorrect — the service cannot function
    - Security has been compromised — the service refuses to continue
    - Obvious configuration problem — the service cannot start

### `error`

- **Output:** `stderr` only
- **Purpose:** Notify administrators of an issue that needs immediate attention.
- **Examples:**
    - Tried to trigger a transaction against provider X for the 10th time and it still failed — someone needs to investigate

### `warning`

- **Output:** `stdout` only
- **Purpose:** Notify maintainers of a non-critical but problematic event. Support and AI will periodically review warnings to decide if escalation is needed. **They are not to be ignored.**
- **When to use:**
    - "There was some issue, but the rest of the system is probably fine so it can be checked out later"
- **When NOT to use:**
    - User typed the wrong password — normal application flow, not a warning
    - Insufficient balance for a transaction — normal flow, may justify `info` at most
    - Wanting logs to "stand out" — use the correct level, not a higher one
- **Examples:**
    - Transaction against provider X timed out, retrying in 30 seconds — might be a network hiccup

### `info`

- **Output:** `stdout` only
- **Purpose:** Help administrators understand the state of the application and track noteworthy events.
- **When to use:** Noteworthy but non-problematic events.
- **When NOT to use:** Events that occur very frequently.
- **Examples:**
    - Printing useful statistics
    - Indicating that a report is being generated (higher-than-normal processing)
    - `{"level":"info","ts":1769499135.3677459,"caller":"ops/workflows.go:219","msg":"No rafiki payment pointer found for wallet","walletID":"2e4ea62d-3e5f-4266-8e03-374cce423584"}`

### `debug`

- **Output:** `stdout` only
- **When to use:** Troubleshooting a specific service in some environment.
- **Examples:**
    - What IP address is a call coming from?
    - Print out an object to understand why a test is failing
- **Special note:** Debug logs should generally be stripped out when merging into main. In some cases developers need debug logs in a target environment — this is acceptable, but be aware of the performance cost of rendering log strings even when they are not actually printed.

## Additional Fields

| Field | Required | Description |
|-------|----------|-------------|
| `ts` | Yes | Timestamp. For Go, follow Zap's unix-format convention. |
| `caller` | Optional | Source location. Provided automatically by Zap in Go. Other services may omit this. |
| `requestId` | Optional | Attach a request ID for easier troubleshooting. |
| `correlationId` | Optional | Correlate async log entries together. |