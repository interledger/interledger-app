# Mock PTI Service Plan

## Goal
Create a new mock provider service at `go/mock/mockpti` that can stand in for PTI/Fiant during local development and automated tests, following the same quality and workflow patterns used by `go/mock/mockgatehub` and `go/mock/mockxago`.

This document is intentionally split into:
1. technical expectations of PTI behavior (source-of-truth requirements)
2. implementation plan for the mock service

## Document Index
- Roadmap: `go/mock/mockpti/roadmap.md`
- Phase 1: `go/mock/mockpti/phase-1-foundations-signup.md`
- Phase 2: `go/mock/mockpti/phase-2-wallets-payment-info.md`
- Phase 3: `go/mock/mockpti/phase-3-transactions.md`
- Phase 4: `go/mock/mockpti/phase-4-webhook-jobs.md`
- Phase 5: `go/mock/mockpti/phase-5-local-integration-sdk.md`

## Formal Documentation and How to Fetch It

Primary source (confirmed):
- Fiant/PTI API reference portal:
  - `https://developers.platform.fiant.io/reference/getusertoken`
  - `https://developers.platform.fiant.io/reference`
- Guides root:
  - `https://developers.platform.fiant.io/docs`

What this portal contains:
- Operation-level docs with method + path (for example `POST /v1/auth/jwt`).
- Header and body parameter descriptions.
- Response status codes and examples.
- A browsable nav for Users, Wallets, Payment Information, Transaction Assessment, Transactions, and Sandbox actions.

Expected API base from docs:
- `https://api.staging.fiant.io/v1`

### OpenAPI/Swagger Availability (Obvious URL Probe Results)
Probe date: 2026-03-18

Portal-hosted spec candidates:
- `https://developers.platform.fiant.io/openapi.json` -> `404` (HTML)
- `https://developers.platform.fiant.io/swagger.json` -> `404` (HTML)
- `https://developers.platform.fiant.io/v3/api-docs` -> `404` (HTML)

API-hosted spec candidates:
- `https://api.staging.fiant.io/v1/openapi.json` -> `404` (JSON)
- `https://api.staging.fiant.io/v1/swagger.json` -> `404` (JSON)
- `https://api.staging.fiant.io/openapi.json` -> `403` (JSON)
- `https://api.staging.fiant.io/swagger.json` -> `403` (JSON)
- `https://api.staging.fiant.io/v3/api-docs` -> `403` (JSON)

Conclusion:
- No publicly retrievable OpenAPI/Swagger document was found at obvious endpoints.
- Formal docs are available as a hosted API reference site (human-readable), not as a clearly exposed public spec URL.

### Practical Retrieval Workflow
1. Start from `https://developers.platform.fiant.io/reference`.
2. Open the specific operation pages relevant to our integration (`getusertoken`, `addauser`, `createwallet`, `deposit`, `withdrawal`, `transfer`, `gettransaction`, `providefeedback`, etc.).
3. Capture method/path/header requirements into this plan and into mock service contracts.
4. For machine-readable contracts (if needed for codegen/mock validation), request OpenAPI export from Fiant support/partner contact.

### If the Team Needs a Machine-Readable Spec
Request from Fiant:
- OpenAPI 3.x JSON/YAML for the exact environment/tenant we use.
- Webhook payload schema (including signing/encryption model and event/resource/status enums).
- Any tenant-specific auth/signature profile that differs from public reference pages.

## Sources Used
- Backend PTI integration:
  - `go/backend/providers/pti/external/*.go`
  - `go/backend/providers/pti/ops/*.go`
  - `go/backend/grpc/pti.go`
  - `go/backend/grpc/kyc.go`
  - `go/backend/main.go`
  - `go/backend/payments/ops/workflows.go`
  - `go/backend/payments/ops/activities.go`
- Frontend PTI flows:
  - `typescript/protea/app/routes/personal-details.tsx`
  - `typescript/protea/app/routes/connect_.card.tsx`
  - `typescript/protea/app/routes/api_.pti_.token.ts`
  - `typescript/protea/app/routes/deposit_.$paymentId.tsx`
  - `typescript/protea/app/routes/withdraw_.$paymentId.tsx`
  - `typescript/protea/app/lib/usePTISdk.ts`
  - `typescript/protea/app/global.d.ts`
- Existing mock patterns:
  - `go/mock/mockgatehub/*`
  - `go/mock/mockxago/*`
- Attached docs:
  - `documentation/docs/kyc-guide.md`
  - `documentation/docs/signup-guide.md`
  - `documentation/docs/provider-payments-reference.md`

## 1) Technical Expectations of PTI Service (Before Implementation)

### 1.1 Integration Shape
The current PTI integration has three channels:
1. Backend to PTI REST API (signed requests)
2. Frontend to PTI browser SDK/forms (script + postMessage)
3. PTI to backend webhook callbacks (`/webhooks/pti`)

### 1.2 Authentication and Request Contract (Backend to PTI)
The backend PTI external client signs every outgoing request and sets PTI headers.

Expected headers (from `go/backend/providers/pti/external/headers.go`):
- `x-pti-client-id`
- `x-pti-signature`
- `x-pti-request-id` (for request-scoped operations)
- `x-pti-scenario-id` (for scenario-based flows)
- `x-pti-disable-webhook` (for selected transfer paths)
- `x-pti-session-id`
- `Date`

Expected env dependencies:
- `PTI_BASE_URL`
- `PTI_CLIENT_ID`
- `PTI_JWK` (private signing key)
- `PTI_PUBLIC_KEY_JWK` (for webhook verification/decryption path)

### 1.3 Backend PTI API Calls (What the mock must emulate)
From `go/backend/providers/pti/external/*` the backend currently calls these PTI endpoints.

User and KYC domain:
- `POST /users` (create user)
- `GET /users/{id}` (get user)
- `PATCH /users` (merge user info)
- `PUT /users` (update user)
- `POST /users/assessments` (start user assessment)
- `GET /users/{userId}/assessments` (get latest assessment)
- `GET /users/{userId}/payment-information/{paymentInformationId}`
- `POST /users/{userId}/payment-information` (create bank account payment method)

Wallet domain:
- `POST /users/{userId}/wallets`
- `GET /users/{userId}/wallets/{walletId}`
- `GET /users/{userId}/wallets`

Transactions domain:
- `POST /transactions/assessments` (start transfer assessment)
- `GET /transactions/assessments/{requestId}`
- `POST /transactions/transfers` (PTI transfer)
- `GET /transactions/{requestId}`
- `POST /transactions/deposits`
- `POST /transactions/withdrawals`
- `POST /transactions/{requestId}/updates` (feedback/status updates back to PTI)

Auth/token domain:
- `POST /auth/jwt`

### 1.4 Frontend PTI Behavior (iframe vs SDK)
Question: does frontend expect PTI to serve iframe workflows like GateHub?

Answer from code: PTI is SDK/form based, not GateHub-style iframe URL embedding.

Observed behavior:
- GateHub path loads an iframe URL from backend (`gatehubWidget.widgetUrl`).
- PTI path loads JS SDK (for example `https://sdk.platform.fiant.io/0.0.23/index.js`) and uses `window.PTI.init(...)` + `window.PTI.form(...)` with a local DOM parent element.

Frontend PTI touchpoints:
- KYC page (`personal-details.tsx`):
  - `PTI.form({ type: 'KYC', requestId, userId, scenarioId, parentElement, ... })`
  - waits for `postMessage` event `UserAssessmentCompleted`
- Card connect (`connect_.card.tsx`):
  - `PTI.form({ type: 'ADD_CC', ... })`
  - waits for `postMessage` event `AddCreditCardCompleted` and reads `createdId` token
- Payment/deposit/withdraw confirm pages call `usePTISdk(sessionId, clientId)` for PTI SDK init.
- Token generation path goes through app backend endpoint `/api/pti/token`, which proxies to backend gRPC `CreatePtiToken` and then PTI `POST /auth/jwt`.

Implication for mockpti:
- Mock service does not need to serve full hosted iframe pages like GateHub.
- It should support token/JWT endpoint and PTI API endpoints used by backend.
- Optional enhancement: local SDK stub assets for deterministic browser E2E, but not required for backend-only tests.

### 1.5 PTI Callbacks Expected by Backend
Question: what callbacks are expected from PTI, likely `/callbacks/pti`?

Answer from code:
- Primary provider callback endpoint is `POST /webhooks/pti` in backend (`go/backend/main.go`).
- There is no active `"/callbacks/pti"` route in backend source.
- PTI webhook handler expects encrypted (JWE) plus signed (JWS) payload processing and validates `clientId`.

Webhook resource types handled:
- `USER`
- `USER_ASSESSMENT` or `KYC`
- `TRANSACTION_STATUS`
- `TRANSACTION_ASSESSMENT` (currently TODO path)

Transaction status values used in logic:
- success/final: `SETTLED`
- failure/error: `REFUSED`, `ERROR`, `CANCELED`
- return path: `RETURNED`
- intermediate/no-op: `PENDING`, `PROCESSING`, `CLEARING_FUNDS`

Note on docs mismatch:
- Some documentation text references callback-style redirect flows; code currently relies on `/webhooks/pti` and frontend postMessage events from PTI SDK, not a backend `/callbacks/pti` handler.

### 1.6 Signup/KYC/Onboarding/Offboarding Expectations
Signup and provisioning:
- PTI user creation via Temporal workflow `CreateUserWorkflow`.
- User is persisted in `pti_users` (`external_id`, wallet mapping, assessment status).

KYC/onboarding:
- US wallets get PTI widget data from `GetKYCProviderWidget`.
- Frontend starts PTI KYC form using SDK.
- Backend sets KYC pending on user action, and final KYC status is updated from PTI webhook assessment status.
- Assessment accepted path drives `kyc.StatusLevel2` plus profile enrichment.

Onboarding money (deposit):
- User creates internal payment in app.
- For US path, frontend confirms and backend triggers `PtiCreateDeposit`.
- Backend starts PTI deposit workflow and creates PTI transaction (`POST /transactions/deposits`).
- Completion depends on PTI transaction status/webhooks (`SETTLED` -> settle workflow).

Offboarding money (withdraw):
- User creates internal payment in app.
- For US path, frontend confirms and backend calls `CreatePTIWithdrawal`.
- Backend creates PTI withdrawal (`POST /transactions/withdrawals`), reserves/finalizes ledger states, and reconciles via PTI status events.

## 2) Implementation Plan for `go/mock/mockpti`

### 2.1 Design Principles
- Match mockgatehub/mockxago patterns:
  - executable in `cmd/mockpti`
  - BDD features under `features/`
  - e2e harness under `testenv/`
  - Makefile with `lint`, `unit-test`, `e2e-test`, `test`, `build`, `clean`
- Use strict test-driven development (TDD): for every behavior, write failing unit tests first, then implement minimal code, then refactor aggressively before moving to the next behavior.
- Persistence must be abstracted behind a single store interface (same architectural intent as `mockxago`) with swappable implementations:
  - memory-backed store (default for unit tests)
  - redis-backed store (used in e2e)
- Keep v1 deterministic and stateful, but with persistence behavior driven through the shared store interface from day one (no direct in-memory-only shortcuts in business logic).
- All outgoing webhook HTTP calls must be executed via a persisted jobs mechanism (same pattern as `mockxago`):
  - enqueue webhook jobs in the store
  - process jobs asynchronously via worker(s)
  - persist job state transitions (queued, processing, delivered, failed/retryable)
- Prioritize endpoints used by live backend code paths.

### 2.2 Proposed Directory Layout
- `go/mock/mockpti/cmd/mockpti/main.go`
- `go/mock/mockpti/internal/config/`
- `go/mock/mockpti/internal/http/`
- `go/mock/mockpti/internal/domain/` (users, wallets, payment_information, transactions)
- `go/mock/mockpti/internal/auth/` (header validation/signature mode)
- `go/mock/mockpti/internal/webhooks/`
- `go/mock/mockpti/features/*.feature`
- `go/mock/mockpti/testenv/*`
- `go/mock/mockpti/README.md`
- `go/mock/mockpti/Makefile`
- `go/mock/mockpti/.golangci.yml`

### 2.3 Phased Delivery

Phase 0: Scaffold and quality baseline
- Initialize module structure.
- Add `Makefile` and `.golangci.yml` aligned to existing mocks.
- Add health endpoint (`GET /health`).
- Establish the persistence seam immediately:
  - define shared store interface
  - implement memory and redis store packages with identical contracts
  - wire service dependencies through the interface only.

Phase 1: Core API for signup + KYC
- Implement:
  - `POST /users`
  - `GET /users/{id}`
  - `POST /users/assessments`
  - `GET /users/{id}/assessments`
  - `POST /auth/jwt`
- Add validation errors with realistic status codes.
- Deliver each endpoint using TDD loops (test first -> implementation -> refactor).

Phase 2: Wallet and payment information
- Implement:
  - `POST /users/{id}/wallets`
  - `GET /users/{id}/wallets`
  - `GET /users/{id}/wallets/{walletId}`
  - `POST /users/{id}/payment-information`
  - `GET /users/{id}/payment-information/{id}`
- Continue TDD-first delivery for each behavior.

Phase 3: Transaction flows
- Implement:
  - `POST /transactions/deposits`
  - `POST /transactions/withdrawals`
  - `POST /transactions/transfers`
  - `GET /transactions/{requestId}`
  - `POST /transactions/{requestId}/updates`
- Add deterministic status progression (PENDING -> SETTLED or configured failure).
- Continue TDD-first delivery for each behavior.

Phase 4: Webhook emission to backend
- Add async webhook sender to configured backend URL (default `/webhooks/pti`) using persisted jobs.
- Emit resource types:
  - `USER_ASSESSMENT` for KYC completion
  - `TRANSACTION_STATUS` for deposit/withdraw/transfer status changes
- Include mode switch:
  - `plain` webhook payload mode for local testing
  - `signed/encrypted` compatibility mode (later enhancement)
- Persist and test job lifecycle semantics (enqueue, execute, retry/backoff, terminal failure) through the shared store interface.

Phase 5: SDK testing support (optional but recommended)
- Provide optional lightweight PTI SDK stub endpoints for E2E (`sdkUrl`, `formsUrl`) so local browser tests are stable without external PTI assets.
- Emit browser `postMessage` events expected by Protea (`UserAssessmentCompleted`, `AddCreditCardCompleted`).

### 2.4 Backend + Protea + Local Environment Changes Needed for Easy Switching
Goal: make PTI switching in local feel like Gatehub/Xago switching today (compose-managed mock service + env defaults), without editing application code each time.

Current state (investigated):
- Gatehub/Xago are first-class local services in compose (`local/mockgatehub.yaml`, `local/mockxago.yaml`) and are included by default from `local/docker-compose.yaml`.
- Backend env defaults in `local/wallet.yaml` point to these local mock services (`http://mockgatehub:8080`, `http://mockxago:8080/v1`) and webhook secrets are shared via `BACKEND_*` env vars.
- PTI is different today:
  - No `mockpti` compose service exists yet.
  - `PTI_ENABLED` defaults to `false` in local.
  - `PTI_BASE_URL` default points to remote staging (`https://api.staging.fiant.io/v1/`).
  - PTI SDK URLs in backend are hardcoded by environment in `go/backend/providers/pti/ops/ops.go`.
  - Protea has mixed PTI SDK configuration: KYC/Card pages use widget-provided SDK URL, but `usePTISdk` is hardcoded to `https://sdk.pearsurge.io/0.0.18/index.js`.

Required changes (file-by-file):

1. Local compose wiring (same pattern as Gatehub/Xago)
- Add `local/mockpti.yaml` with service `mockpti`:
  - Build from `go/mock/mockpti/Dockerfile`.
  - Expose port `8080`.
  - Add Traefik host rule for `mockpti.interledger.test`.
  - Configure webhook target to backend: `http://backend:8080/webhooks/pti`.
  - Pass mockpti runtime vars (for example webhook delay/mode).
- Include `mockpti.yaml` in `local/docker-compose.yaml`.
- Add backend dependency on `mockpti` in `local/wallet.yaml` (`depends_on` block), matching existing mock dependencies.

2. Local env defaults (switch by env var, not code edits)
- In `local/wallet.yaml`:
  - Keep `PTI_ENABLED=${BACKEND_PTI_ENABLED:-false}` if you want explicit opt-in, or change default to `true` once mockpti is stable.
  - Change local default PTI base URL to mock service:
    - `PTI_BASE_URL=${BACKEND_PTI_BASE_URL:-http://mockpti:8080}`
  - Keep `PTI_CLIENT_ID`, `PTI_JWK`, `PTI_PUBLIC_KEY_JWK` from local defaults so backend still starts when PTI is enabled.
- In `local/example.env`:
  - Add/align PTI override block with explicit mock defaults and commented real-PTI values, mirroring Gatehub conventions.

3. Hosts and certificates for browser-facing PTI assets
Needed if mockpti serves SDK/forms URLs to browser via Traefik domain:
- Add `mockpti.interledger.test` to `HOSTS` in `local/Makefile`.
- Add `mockpti.interledger.test` SAN entry in `local/config/san.cnf`.
- Add URL row to `local/README.md` table (like MockGateHub/MockXago).

4. Backend PTI widget configurability (remove hardcoded SDK/forms URLs)
- In `go/backend/providers/pti/ops/ops.go`, replace hardcoded `sdkUrl/formsUrl` selection with env-overridable values, for example:
  - `PTI_SDK_URL`
  - `PTI_FORMS_URL`
  - fall back to existing staging/prod defaults when vars are unset.
- Keep `GenerateTokenPath` as backend URL (`/api/pti/token`) so browser token requests continue to proxy through Protea -> backend gRPC.

5. Backend startup and webhook compatibility
- No new backend route is required: backend already handles `POST /webhooks/pti` when `PTI_ENABLED=true`.
- Ensure mockpti webhook payload format has a mode the backend can consume during local runs:
  - preferred initial mode: plain JSON webhook payload compatible with current PTI webhook handler expectations.
  - optional later mode: signed/encrypted parity mode.

6. Protea PTI SDK consistency for switching
- Remove hardcoded PTI SDK URL in `typescript/protea/app/lib/usePTISdk.ts`.
- Make `usePTISdk` consume SDK URL from server data (or from `ptiWidget.sdkUrl`) so all PTI flows use one source of truth from backend.
- In `deposit_.$paymentId.tsx` and `withdraw_.$paymentId.tsx`:
  - include PTI widget metadata from backend (at minimum `sdkUrl`; optionally `formsUrl/generateTokenPath` if needed by SDK init path).
  - pass SDK URL into `usePTISdk`.
- Keep `/api/pti/token` proxy route as-is; it already supports token generation through backend (`CreatePtiToken`).

7. Optional hardening change in backend PTI client initialization
- Today `pti_client.New` always parses `PTI_JWK` and will fail startup if malformed, even if PTI is not used.
- Optional improvement: initialize real PTI client only when `PTI_ENABLED=true`, and inject a no-op/mock client otherwise.
- This is not strictly required for local switching if defaults remain valid, but improves robustness.

Recommended local switch UX (target behavior):
- Default local profile: Gatehub + Xago + PTI all backed by local mocks.
- To test against real PTI temporarily: override `BACKEND_PTI_BASE_URL`, `BACKEND_PTI_CLIENT_ID`, keys, and optional SDK/forms URLs in `.env`.
- No frontend/backend code changes required between mock and real PTI.

## 3) Testing and Linting Plan (aligned with mockgatehub/mockxago)

### 3.0 Delivery Method: Mandatory TDD
- For each behavior in each phase, implementation order is mandatory:
  1. write failing unit test(s)
  2. implement the minimal code to pass
  3. refactor aggressively while keeping tests green
- Do not merge behavior-first code without prior failing tests.

### 3.1 Make targets
Planned targets:
- `lint`: `gofmt -l .`, `go vet ./...`, `golangci-lint run ./...`
- `unit-test`: `go test -count=1 -v ./cmd/... ./internal/...`
- `e2e-test`: `go test -count=1 -v -tags e2e ./testenv`
- `test`: `lint unit-test e2e-test`
- `build`: `go build -o mockpti ./cmd/mockpti`
- `clean`: remove binary/artifacts

### 3.2 BDD strategy
- Use godog feature files in `features/` (created below).
- Testenv should start `mockpti` and (optionally) a webhook sink.
- Validate both API responses and side effects (state transitions + webhook events).

### 3.3 Store and Queue Test Matrix
- Unit tests:
  - run against memory-backed store only
  - verify domain behavior, validation, and job enqueue semantics without redis dependency
- E2E tests:
  - run with redis-backed store enabled
  - verify persistence across process boundaries and webhook job processing behavior end-to-end
- Webhook queue tests must assert persisted job state transitions and retries, not only successful HTTP delivery.

## 4) Open Questions and Team Follow-up Checklist

### 4.1 Auth Contract: Client-ID Only vs Signed Requests
Context:
- Public docs pages consistently show `x-pti-client-id`.
- Our backend implementation expects/signs additional headers (`x-pti-signature`, `Date`, and flow headers such as request/scenario/session IDs).

Question:
- Is our tenant on an extended signed-auth profile beyond the public docs baseline?

Why this matters:
- It determines whether mockpti must enforce signature validation in v1 or can run with client-id-only checks.

Follow-up ask to team/PTI:
- "Please confirm the authoritative auth/header contract for our tenant (required headers, signing algorithm, canonical string rules, clock skew tolerance)."

### 4.2 Webhook Schema and Security Mode
Context:
- Backend currently handles PTI webhooks at `/webhooks/pti` and expects signed/encrypted processing paths.
- We did not find a clear webhook reference page at obvious docs URLs.

Question:
- What is the canonical webhook schema and required crypto envelope for our integration?

Why this matters:
- Determines mock webhook payload format, status/resource enums, and verification behavior.

Follow-up ask to team/PTI:
- "Please provide webhook docs/spec for PTI (resource types, status enums, required fields, JWE/JWS expectations, retry policy, and signature verification inputs)."

### 4.3 Transaction Status Vocabulary Alignment
Context:
- Public docs (`providefeedback`) show statuses including `CANCELLED`, `REJECTED`, `AWAITING_PAYMENT`, `PENDING_SETTLEMENT`, `RETURNED`.
- Our plan/code paths currently rely on statuses like `SETTLED`, `REFUSED`, `ERROR`, `CANCELED`, `RETURNED`, `PENDING`, `PROCESSING`, `CLEARING_FUNDS`.

Question:
- Which status set is authoritative for API responses and webhook events in our tenant?

Why this matters:
- Directly impacts payment workflow transitions and mock scenario fidelity.

Follow-up ask to team/PTI:
- "Please provide the exact status mapping for API + webhooks (including spelling variants such as `CANCELLED` vs `CANCELED`, and whether `REFUSED` or `REJECTED` is used)."

### 4.4 SDK/Forms URL Source of Truth
Context:
- KYC/card flows consume backend-provided PTI widget URLs.
- Some frontend code still contains hardcoded SDK URL(s).

Question:
- Should all PTI SDK/form endpoints be controlled by backend/env (single source of truth)?

Why this matters:
- Needed for reliable local mock switching and environment consistency.

Follow-up ask to team:
- "Confirm we should fully remove hardcoded PTI SDK URLs from Protea and drive all PTI SDK/form config via backend widget/env."

### 4.5 Local Mock Webhook Mode for Milestone 1
Context:
- v1 mock proposal includes a plain webhook mode first, with signed/encrypted mode as enhancement.

Question:
- Is plain mode acceptable for first milestone, or must parity crypto be included immediately?

Why this matters:
- Affects initial implementation complexity and delivery timeline.

Follow-up ask to team:
- "Approve whether Milestone 1 can ship with plain webhook mode for local tests, with crypto mode in Milestone 2."

### 4.6 Transaction Progression Timing and Determinism
Context:
- Real PTI flows are asynchronous.
- Mock can be immediate or delayed/configurable.

Question:
- What timing behavior is most useful for local dev and CI stability?

Why this matters:
- Influences flaky tests vs realistic behavior tradeoff.

Follow-up ask to team:
- "Choose default progression strategy: immediate transitions, fixed delay, or configurable delay per scenario."

### 4.7 Scenario Control API for Tests
Context:
- Planned optional feature: force next transaction outcome (for example `REFUSED`).

Question:
- Should we expose a test-control endpoint in mockpti for deterministic failure testing?

Why this matters:
- Enables robust negative-path E2E coverage without brittle timing hacks.

Follow-up ask to team:
- "Approve a test-only scenario control endpoint and define allowed controls (next status, delay, webhook drop/retry simulation)."

### 4.8 Need for Official OpenAPI Export
Context:
- No obvious public OpenAPI/Swagger endpoint was available in probes.

Question:
- Do we need an official machine-readable spec for long-term maintenance and contract testing?

Why this matters:
- Reduces drift between mock behavior and provider contract.

Follow-up ask to team/PTI:
- "Request OpenAPI 3.x export for our tenant/environment, or confirm that the hosted reference is the only supported source."

## 5) Acceptance Criteria for First Milestone
- Service starts and passes lint/unit/e2e.
- Development follows mandatory TDD loops (test-first, then implementation, then refactor) for all delivered behaviors.
- Persistence is exclusively accessed through one shared interface with both memory and redis implementations.
- Unit tests use only memory-backed persistence; mockpti e2e uses redis-backed persistence.
- Outgoing webhook HTTP calls are executed through a persisted jobs queue (not direct fire-and-forget HTTP calls).
- Local environment can run PTI mock through compose (same operational model as `mockgatehub`/`mockxago`):
  - `mockpti` service is included in local compose.
  - backend defaults use `PTI_BASE_URL=http://mockpti:8080` when PTI is enabled.
  - browser-accessible PTI mock host is available (`mockpti.interledger.test`) when SDK/forms are served locally.
- Backend can point `PTI_BASE_URL` to mockpti and complete:
  - PTI user creation path
  - PTI KYC widget token path (`/auth/jwt` via `/api/pti/token` in app)
  - PTI bank account link path
  - PTI deposit/withdraw creation path
- Protea PTI flows do not require hardcoded external SDK URLs to run locally.
- Backend receives and processes transaction and assessment webhooks from mockpti through `/webhooks/pti`.

---

## Proposed Feature Inventory
Feature files are created in `go/mock/mockpti/features` for review:
- `service_health.feature`
- `user_and_kyc.feature`
- `wallet_and_payment_information.feature`
- `token_generation.feature`
- `transactions.feature`
- `webhooks.feature`
