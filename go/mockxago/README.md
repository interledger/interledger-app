# MockXago

MockXago is a lightweight mock of the Xago API used by the Interledger Wallet for local development and tests. It implements the minimal endpoints and webhook behavior that the wallet backend depends on.

## Quick Start

```bash
# From this directory
export XAGO_API_PUBLIC_KEY=test-public-key
export XAGO_API_SECRET=test-secret
export XAGO_MOCK_PORT=8080
export WEBHOOK_URL=http://localhost:3000/xago/webhooks/event
export WEBHOOK_SECRET=local-webhook-secret

# Optional: enable test-only endpoints
export XAGO_MOCK_TEST_MODE=true

go run ./cmd/mockxago
```

Health check:

```bash
curl -s http://localhost:8080/health
```

## Configuration

- `XAGO_MOCK_PORT`: HTTP port (default `8080`)
- `XAGO_API_PUBLIC_KEY`: expected login public key (default `test-public-key`)
- `XAGO_API_SECRET`: expected login secret key (default `test-secret`)
- `XAGO_MOCK_TEST_MODE`: enables `/v1/test/*` endpoints when set to `true`
- `WEBHOOK_URL`: where MockXago sends wallet-facing webhooks (KYC + deposits)
- `WEBHOOK_SECRET`: HMAC secret for `X-Signature` header
- `PERSONA_WEBHOOK_URL`: optional Persona-style webhook URL (default `http://backend:8080/webhooks/persona`)
- `PERSONA_WEBHOOK_TOKEN`: HMAC secret for Persona webhook signatures

## Authentication

Most `/v1` endpoints require a bearer token:

```
Authorization: Bearer <tokenValue>
```

Get a token with `POST /v1/login`.

## API

### POST /v1/login

Request:

```json
{
  "policyId": "5e2585a474b0e90012ce8ff1",
  "fields": [
    {"fieldName": "apiPublicKey", "fieldValue": "test-public-key"},
    {"fieldName": "apiSecretKey", "fieldValue": "test-secret"}
  ]
}
```

Response:

```json
{
  "tokenValue": "..."
}
```

### GET /v1/currencies

Returns bank details for ZAR and USD.

### POST /v1/company/accounts

Create a sub-account. Required fields:

- `firstName`, `lastName`, `email`, `mobileNumber`, `identityType`, `idNumber`, `physicalAddress`, `thirdPartyVerificationUrl`
- Optional `walletId` (if omitted, MockXago generates one)

Response includes `accountId`, `bankDepositDetails`, and deposit references.

### PUT /v1/company/accounts/{accountId}

Update KYC-related fields for a sub-account.

### GET /v1/company/accounts?walletId={walletId}

Fetch a sub-account by wallet id.

### GET /v1/accounts/{accountId}/balance

Returns balances for ZAR and USD.

### GET /v1/company/transactions
### GET /v1/company/transactions/{id}

Lists or fetches deposit transactions. These endpoints require authentication.

### KYC iframe (wallet flow)

- `GET /kyc/iframe?token={token}&user_id={walletId}`
- `POST /kyc/submit` (form submission)

### Persona-compatible KYC endpoints

- `POST /v1/inquiries`
- `GET /v1/inquiries/{inquiryId}`
- `GET /v1/inquiries/{inquiryId}/iframe`
- `POST /v1/inquiries/{inquiryId}/submit`

### Test-only endpoints (require `XAGO_MOCK_TEST_MODE=true`)

All are authenticated with bearer token:

- `POST /v1/test/balances/set`
- `POST /v1/test/balances/deposit`
- `POST /v1/test/balances/transfer`
- `POST /v1/test/transactions`

## Webhooks

MockXago sends two wallet-facing webhook types to `WEBHOOK_URL`:

### KYC webhook

Triggered after KYC iframe or Persona submit. Payload:

```json
{
  "event_type": "id.verification.accepted",
  "wallet_id": "<walletId>",
  "timestamp": "2026-01-20T10:00:00Z",
  "data": {"message": "Persona KYC verification accepted"}
}
```

Authentication:

- Header: `X-Signature`
- Value: hex-encoded HMAC-SHA256 of the raw JSON body using `WEBHOOK_SECRET`

### Deposit webhook

Triggered by `POST /v1/test/balances/deposit` (test mode). Payload:

```json
{
  "accountId": "xago-acct-<walletIdPrefix>",
  "amount": 5000,
  "currencyCode": "ZAR",
  "transactionId": "<transactionId>",
  "code": 104,
  "status": "settled",
  "createdAt": "2026-01-20T10:00:00Z",
  "settledAt": "2026-01-20T10:00:00Z",
  "isDuplicate": false,
  "isRequested": false,
  "isRequestMatched": false
}
```

Authentication:

- Header: `X-Signature`
- Value: hex-encoded HMAC-SHA256 of the raw JSON body using `WEBHOOK_SECRET`

### Withdrawal webhooks

MockXago does not emit withdrawal webhooks. Transfers in this mock are handled via test endpoints that adjust balances only.

## How the real flow maps to the mock

1. The wallet retrieves Xago banking details using `GET /v1/currencies`.
2. The user EFTs funds to the bank account with the provided reference.
3. Xago confirms the deposit and sends a webhook to the wallet backend.
4. In the mock, you simulate that confirmation by calling `POST /v1/test/balances/deposit`, which both credits the balance and emits the deposit webhook.
