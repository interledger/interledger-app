# Error Handling Guide

> **Developer reference for backend and frontend error handling.** This page documents the standardized gRPC error contract introduced by commit `95d6a42ba6b09aef2aa290a3ce80bc2729563623` and the conventions all new work in this repository should follow.

**Related documents:**

- [Logging Reference](logging-reference.md) — How errors should be logged
- [Payment Troubleshooting](payment-troubleshooting-guide.md) — Operational debugging workflow
- [Environment Variables](env-variables.md) — Runtime configuration and request tracing context

**Quick Navigation:**

- **What is the standard error contract?** → See [Canonical gRPC Contract](#canonical-grpc-contract)
- **How do I return a validation error?** → See [Validation Errors](#validation-errors)
- **How do I add a new domain error?** → See [Adding New Error Mappings](#adding-new-error-mappings)
- **What does the frontend consume?** → See [Frontend Expectations](#frontend-expectations)

---

## Purpose

All new work in `go/backend/grpc` must return errors in a predictable, machine-readable format.

The standard introduced by the reviewed commit has three goals:

1. Every gRPC error must carry a structured `AppError` detail.
2. Validation and precondition failures must expose actionable field or subject information.
3. Request IDs must be attached to errors so support and developers can correlate frontend failures with backend logs.

This guide applies to `go/backend/grpc` and to frontend consumers in Protea that read gRPC errors.

This guide does **not** redefine the older `go/backend/api/apperrors` HTTP admin flow. That package now shares the same central `errcodes`, but the commit's main standardization target was `backend/grpc`.

## Error Categories And Builders

Every error returned by `go/backend/grpc` should end up as a gRPC status with:

- a gRPC status code such as `InvalidArgument`, `FailedPrecondition`, `NotFound`, or `Internal`
- one `backend.v1.AppError` detail
- optional standard Google error details such as `google.rpc.BadRequest` or `google.rpc.PreconditionFailure`
- a populated `reqId` on `AppError` when request ID context is available

The protobuf contract is:

```proto
message AppError {
  string errorCode = 1;
  string message = 2;
  repeated AppErrorField fields = 3;
  string reqId = 4;
}

message AppErrorField {
  string field = 1;
  string error = 2;
}
```

The authoritative app error codes live in `go/backend/errcodes/errcodes.go`.

## How The Backend Produces Errors

### 1. Handler and service code returns an error

Most RPC handlers should either:

- return `toGRPCError(err)` for domain/service errors
- return a helper such as `NewValidationError(...)` or `PaymentPreconditionError(...)` when the handler itself is deciding the user-facing contract

The normal pattern is:

```go
result, err := someService.DoThing(ctx, args)
if err != nil {
    return nil, toGRPCError(err)
}
```

### 2. `toGRPCError()` maps domain errors into the public contract

`toGRPCError()` is the central translation layer.

It does the following:

- logs the original error at `info` in non-test environments
- converts validator errors into structured validation responses
- maps known sentinel errors through the `errorStatus` map
- supports wrapped errors via `errors.Is(...)`
- captures unexpected errors in Sentry and returns a generic internal error

If you add a new domain error that should have a stable public contract, it should usually be mapped here.

### 3. `withAppError()` guarantees an `AppError`

The unary interceptor in `middleware_apperror.go` runs after the handler returns.

Its job is to ensure that:

- every gRPC error contains exactly one `AppError`
- the `reqId` on `AppError` is populated from request context
- existing `AppError` details are preserved rather than overwritten
- even raw, unwrapped errors become `codes.Internal` responses with an `AppError`

This means developers should still return properly structured errors themselves. The interceptor is a safety net, not the primary place where business semantics should be defined.

## Canonical gRPC Contract

### Validation Errors

Use `codes.InvalidArgument` when the client can fix the request and retry.

Validation responses should include:

- `google.rpc.BadRequest` field violations
- `AppError.errorCode = VALIDATION` unless a more specific app code is required
- `AppError.fields[]` mirroring the field violations

Use these helpers:

- `NewValidationError(field, description)` for a single field violation
- `ValidationError(err, descriptionFn)` when converting `validator.ValidationErrors`

Typical examples:

- malformed phone number
- invalid email
- missing required amount
- malformed UUID

### Precondition Errors

Use `codes.FailedPrecondition` when the request is structurally valid but the system is not in a state that allows the action.

Examples:

- a card is blocked
- a payment requires OTP before confirmation
- KYC resubmission is required

Use helpers like:

- `CardPreconditionError(...)`
- `PaymentPreconditionError(...)`

These responses should include `google.rpc.PreconditionFailure` details.

### Not Found, Conflict, Unauthorized, Internal

Known domain sentinels are mapped in `errorStatus` using `newError(...)` and related builders.

Examples already standardized include:

- `USER_NO_USER_FOUND`
- `WALLETS_NO_WALLET_FOUND`
- `WALLETS_DUPLICATE_WALLET`
- `LINKEDACC_NOT_FOUND`
- `PAYMENTS_REQUIRED_ACTIONS`

Unexpected errors must fall back to:

- `codes.Internal`
- `AppError.errorCode = INTERNAL`
- a generic public message such as `Internal server error`

Do not leak provider payloads, SQL errors, or internal stack details to clients.

## Validation Errors

### Field naming rules

Field names are part of the frontend contract.

Use one canonical field name for one user input concept. Do not vary names by route, provider, or legacy implementation.

Examples:

- use `otp` for OTP code input
- use `phone` for phone-number input
- use `amount` for amount input

Avoid mixed or legacy variants for the same concept, such as `OTP`, `Code`, or `To`, unless you are explicitly documenting a compatibility bridge during migration.

### Message rules

Messages should be:

- user-actionable
- stable enough for tests
- free of internal implementation details

Good:

- `Phone number is invalid.`
- `Could not validate OTP`
- `Amount is required`

Bad:

- raw Twilio API text
- raw SQL constraint names
- internal service errors

## Adding New Error Mappings

When adding a new backend error path, use this checklist:

1. Define or reuse a sentinel/domain error in the owning package.
2. Decide the correct public gRPC status code.
3. Add or reuse a stable `errcodes.AppErrorCode`.
4. Map the error in `go/backend/grpc/errors.go` if it should be standardized across handlers.
5. Attach standard details such as `BadRequest` or `PreconditionFailure` when relevant.
6. Add tests that assert the exact status code and attached details.

If the handler is constructing the error directly, prefer existing helpers over ad-hoc `status.New(...).WithDetails(...)` code.

### Preferred patterns

Preferred:

```go
if err != nil {
    return nil, toGRPCError(err)
}
```

Preferred:

```go
if invalidInput {
    return nil, NewValidationError("phone", "Phone number is invalid.")
}
```

Avoid:

```go
return nil, status.Error(codes.Internal, err.Error())
```

Avoid:

```go
// route-local custom shapes that bypass the central contract
return nil, someOneOffStatusWithDifferentFieldNames()
```

## Frontend Expectations

Protea reads structured gRPC errors from the custom Connect client in `typescript/protea/app/lib/grpc.server.ts` and `typescript/protea/app/lib/error.server.ts`.

Frontend consumers should rely on:

- gRPC status code
- `google.rpc.BadRequest` field violations
- `AppError.errorCode`

Frontend code should **not** depend on:

- provider-specific raw messages
- inconsistent field names across routes
- undocumented detail ordering

When the backend changes an error contract, the corresponding Protea error mappers must be updated in the same stream of work.

## Request IDs And Debugging

The request ID middleware and `withAppError()` work together so that backend errors include `AppError.reqId`.

Use this flow when debugging:

1. Capture the request ID from the frontend-visible error.
2. Search backend logs for the same request ID.
3. Inspect the original error logged by `toGRPCError()` or the handler.
4. Verify the mapped status code and detail payload in tests when changing behavior.

This is the intended bridge between support workflows, logs, and frontend-visible failures.

## Testing Requirements

Any new standardized error path should have focused tests.

At minimum, assert:

- the gRPC status code
- the `AppError.errorCode`
- the `AppError.fields` content when applicable
- `BadRequest` or `PreconditionFailure` detail contents when applicable
- request ID propagation when testing interceptor behavior

Useful existing test patterns:

- `go/backend/grpc/errors_test.go`
- `go/backend/grpc/middleware_apperror_test.go`
- `go/backend/grpc/status_helpers_test.go`

## Scope Boundary

This standard currently governs `go/backend/grpc`.

The older HTTP/admin path in `go/backend/api/apperrors` now shares the central `errcodes` package, but it is still a separate response format and should not be conflated with the gRPC detail contract.

If future work standardizes the admin surface as well, update this guide explicitly rather than assuming the contracts are already identical.

## Developer Rules

All developers working on this project should follow these rules for new backend error handling work:

1. Do not return raw internal errors directly from gRPC handlers.
2. Prefer `toGRPCError()` for service/domain failures.
3. Use canonical field names consistently across all handlers.
4. Use stable `errcodes.AppErrorCode` values for machine-readable behavior.
5. Treat the `withAppError()` interceptor as a guarantee layer, not the place where business semantics are defined.
6. Add focused tests for every new standardized error path.
7. Update Protea consumers when backend contracts change.

If a change does not satisfy these rules, it is not aligned with the repository's current error-handling standard.