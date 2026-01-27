# Xago Mock Implementation Plan

## Overview

This document describes the specification for building a Xago mock service for local development and testing. Xago is a financial infrastructure provider that enables multi-currency payments, KYC verification, and balance account management. The Interledger Wallet integrates with Xago to provide South African (ZAR) and USD balance accounts.

**Note**: An existing `mockbos` implementation exists at `/go/mockbos`, but this document specifies a fresh, complete implementation with updated patterns and full feature coverage.

## Architecture Context

### Role in Wallet Ecosystem

Xago integration in the Interledger Wallet serves the following functions:

1. **Sub-Account Creation**: Creates individual sub-accounts within Xago that correspond to wallet users
2. **KYC Verification**: Links Persona KYC results with Xago for regulatory compliance
3. **Balance Accounts**: Manages multi-currency balance accounts (ZAR and USD)
4. **Fund Deposits**: Provides deposit addresses and references for receiving fiat deposits
5. **Beneficiary Management**: Manages bank accounts as withdrawal beneficiaries
6. **Transactions**: Enables transfers between Xago accounts and external bank accounts

### When Xago is Invoked

Xago operations are triggered at specific points in the wallet lifecycle:

| Trigger | Flow | Xago Action |
|---------|------|------------|
| User completes KYC (ZA only) | KYC status → Approved for ZA | Create balance account, create sub-account |
| User enables Xago wallet features | Wallet initialization | Create balance account if not already created |
| User deposits funds | Deposit flow | Receive webhook notification, credit balance |
| User sends funds | Transaction initiation | Create transfer, debit balance, invoke beneficiary |
| User adds bank account | Beneficiary setup | Create beneficiary record |

## Data Model

### 1. Sub-Account

A Xago sub-account represents a user's identity and regulatory profile within Xago.

**Database Table**: `xago_sub_accounts`

```sql
CREATE TABLE xago_sub_accounts (
    id UUID PRIMARY KEY,                          -- Xago-assigned ID
    wallet_id UUID NOT NULL UNIQUE,               -- Link to wallet
    account_id UUID NOT NULL UNIQUE,              -- Xago account ID
    deposit_address VARCHAR(255),                 -- Crypto deposit address
    deposit_tag INT,                              -- Crypto deposit tag (for XRP)
    deposit_reference VARCHAR(255),               -- Fiat deposit reference
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Creation Trigger**: When KYC is approved for a South African user (wallet.country = 'ZA')

**Key Fields**:
- `account_id`: Unique identifier from Xago for the sub-account
- `deposit_reference`: Used in bank deposit descriptions to route funds to correct wallet
- `deposit_address` / `deposit_tag`: For cryptocurrency deposits (future feature)

### 2. Balance Account (Linked Account)

A balance account holds currency balances (ZAR or USD) for a user's wallet.

**Database Table**: `linked_accounts` (with provider='xago', type='balance')

```sql
-- Relevant columns for Xago balance accounts
CREATE TABLE linked_accounts (
    id UUID PRIMARY KEY,                          -- Auto-generated
    wallet_id UUID NOT NULL,                      -- User's wallet
    provider VARCHAR(50) NOT NULL,                -- 'xago'
    provider_id VARCHAR(255) NOT NULL,            -- Deterministic: 'xago_{currency}_{wallet_id}'
    type VARCHAR(50) NOT NULL,                    -- 'balance' (vs 'bank_account')
    name VARCHAR(255),                            -- Display name
    nickname VARCHAR(255),                        -- User-given nickname
    state VARCHAR(50) NOT NULL,                   -- 'verified'
    can_send BOOLEAN DEFAULT true,                -- Can initiate transfers
    can_receive BOOLEAN DEFAULT true,             -- Can receive deposits
    send_currency CURRENCY,                       -- ZAR or USD
    receive_currency CURRENCY,                    -- ZAR or USD
    send_country COUNTRY,                         -- ZA or US
    receive_country COUNTRY,                      -- ZA or US
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Key Properties**:
- Created via `CreateBalanceAccountWorkflow`
- Linked to Pacioli ledger for balance tracking
- Immutable once created (identified by deterministic `provider_id`)

### 3. Beneficiary

Bank accounts that receive funds from Xago transfers.

**Database Table**: `xago_beneficiaries`

```sql
CREATE TABLE xago_beneficiaries (
    id UUID PRIMARY KEY,                          -- Xago beneficiary ID
    wallet_id UUID NOT NULL,                      -- User's wallet
    provider_id VARCHAR(255),                     -- Xago reference
    bank_name VARCHAR(255),                       -- Bank name
    account_number VARCHAR(255),                  -- Account number
    account_name VARCHAR(255),                    -- Account holder name
    branch_code VARCHAR(10),                      -- Bank branch code
    currency CURRENCY,                            -- ZAR or USD
    country COUNTRY,                              -- Country of bank
    status VARCHAR(50),                           -- 'pending', 'approved', 'rejected'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4. Xago Access Token

Tokens for API authentication.

**Database Table**: `xago_access_token`

```sql
CREATE TABLE xago_access_token (
    id UUID PRIMARY KEY,                          -- Single record (immutable)
    token VARCHAR(1000) NOT NULL,                 -- Bearer token
    expires_at TIMESTAMP NOT NULL,                -- Expiration time
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

The Xago mock service should implement the following endpoints. All requests require Bearer token authentication (except login).

### Authentication

#### 1. POST `/xago/v1/login`

Authenticate and obtain a Bearer token.

**Request**:
```json
{
  "policyId": "5e2585a474b0e90012ce8ff1",  // Staging policy
  "fields": [
    {
      "fieldName": "apiPublicKey",
      "fieldValue": "{{ XAGO_API_PUBLIC_KEY }}"
    },
    {
      "fieldName": "apiSecretKey",
      "fieldValue": "{{ XAGO_API_SECRET }}"
    }
  ]
}
```

**Response**:
```json
{
  "tokenValue": "eyJhbGc..."
}
```

**Behavior**:
- Validate the provided API keys against configured environment variables
- Generate and store a JWT or opaque token with 55-minute expiration
- Return the token in JSON response
- Subsequent requests should use: `Authorization: Bearer {tokenValue}`

**Status Codes**: 
- `200 OK` - Token generated
- `401 Unauthorized` - Invalid credentials
- `400 Bad Request` - Missing required fields

---

### Sub-Account Management

#### 2. POST `/xago/v1/company/accounts`

Create a sub-account for a user.

**Request**:
```json
{
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "mobileNumber": "+27123456789",
  "identityType": "individual",
  "idNumber": "9001011234567",
  "physicalAddress": "123 Main St, Cape Town, SA",
  "thirdPartyVerificationUrl": "https://app.withpersona.com/dashboard/inquiries/{inquiry_id}"
}
```

**Response**:
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440001",
  "depositAddress": "rN7n7otQDd6FczFgLdmLN6xYLqDMNxqWPh",
  "depositTag": 12345,
  "bankDepositDetails": {
    "ZAR": [
      {
        "bankName": "FNB",
        "accountName": "Xago Holdings",
        "accountNumber": "62057334567",
        "branchCode": "250145",
        "IBAN": "ZA...",
        "swiftBIC": "FIRSZA22"
      }
    ],
    "USD": [
      {
        "bankName": "Citibank",
        "accountName": "Xago Inc",
        "accountNumber": "0123456789",
        "branchCode": "021",
        "IBAN": "US...",
        "swiftBIC": "CITIUS33"
      }
    ]
  },
  "beneficiaries": [
    {
      "beneficiaryId": "ben_001",
      "beneficiaryType": "rollup",
      "currencyId": "ZAR",
      "depositReference": "USER_12345_ZAR",
      "accountNumber": "62057334567",
      "bankName": "FNB",
      "accountName": "Xago Holdings"
    },
    {
      "beneficiaryId": "ben_002",
      "beneficiaryType": "rollup",
      "currencyId": "USD",
      "depositReference": "USER_12345_USD",
      "accountNumber": "0123456789",
      "bankName": "Citibank",
      "accountName": "Xago Inc"
    }
  ]
}
```

**Behavior**:
- Generate a unique `accountId` (UUID)
- Generate fake but realistic `depositAddress` and `depositTag` (for crypto)
- Return hardcoded bank account details for ZAR and USD
- Create beneficiaries with a `depositReference` that the wallet uses to identify incoming deposits
- Store the sub-account in database linked to the wallet
- Return `200 OK`

**Status Codes**: 
- `200 OK` - Sub-account created
- `401 Unauthorized` - Invalid token
- `400 Bad Request` - Missing required fields

---

#### 3. PUT `/xago/v1/company/accounts/{accountId}`

Update a sub-account (e.g., when KYC inquiry link changes).

**Request**:
```json
{
  "thirdPartyVerificationUrl": "https://app.withpersona.com/dashboard/inquiries/{new_inquiry_id}",
  "idNumber": "9001011234567",
  "physicalAddress": "123 Updated St, Cape Town, SA"
}
```

**Response**: 
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440001",
  "status": "updated"
}
```

**Behavior**:
- Update the sub-account record with new KYC/address information
- Return updated confirmation
- This is called when Persona KYC completes with updated information

---

#### 4. GET `/xago/v1/currencies`

List available bank accounts and currencies.

**Response**:
```json
[
  {
    "currencyId": "ZAR",
    "currencyName": "South African Rand",
    "bankName": "FNB",
    "accountName": "Xago Holdings",
    "accountNumber": "62057334567",
    "branchCode": "250145",
    "swiftBIC": "FIRSZA22"
  },
  {
    "currencyId": "USD",
    "currencyName": "US Dollar",
    "bankName": "Citibank",
    "accountName": "Xago Inc",
    "accountNumber": "0123456789",
    "branchCode": "021",
    "swiftBIC": "CITIUS33"
  }
]
```

**Behavior**:
- Return hardcoded list of available currencies and corresponding bank accounts
- No authentication required (or optional)
- Used during wallet initialization to show available deposit options

---

### Beneficiary Management

#### 5. POST `/xago/v1/beneficiaries`

Add a bank account as a withdrawal beneficiary.

**Request**:
```json
{
  "name": "My Bank Account",
  "scope": "external",
  "currencyCode": "ZAR",
  "accountNumber": "1234567890",
  "branchCode": "250155",
  "bankName": "ABSA",
  "accountName": "John Doe",
  "reference": "My ABSA Account",
  "kycRequest": {
    "isOwn": true,
    "existingIdentityId": "550e8400-e29b-41d4-a716-446655440001"
  }
}
```

**Response**:
```json
{
  "beneficiaries": [
    {
      "uuid": "ben_ext_001",
      "name": "My Bank Account",
      "scope": "external",
      "currencyCode": "ZAR",
      "accountNumber": "1234567890",
      "branchCode": "250155",
      "bankName": "ABSA",
      "accountName": "John Doe",
      "reference": "My ABSA Account",
      "status": "pending"
    }
  ]
}
```

**Behavior**:
- Generate a unique `uuid` for the beneficiary
- Store in database with `status: "pending"`
- KYC verification happens asynchronously (in mock, auto-approve after short delay)
- Return the created beneficiary
- Store association between wallet and beneficiary

---

#### 6. GET `/xago/v1/beneficiaries?limit={limit}&page={page}`

List beneficiaries for the authenticated user.

**Response**:
```json
{
  "meta": {
    "limit": 10,
    "pageNumber": 1,
    "numberOfPages": 1
  },
  "values": [
    {
      "uuid": "ben_ext_001",
      "name": "My Bank Account",
      "scope": "external",
      "currencyCode": "ZAR",
      "accountNumber": "1234567890",
      "reference": "My ABSA Account",
      "status": "approved"
    }
  ]
}
```

**Behavior**:
- Return all beneficiaries created by the authenticated user
- Support pagination
- Filter by wallet if needed

---

### Balance Management

#### 7. GET `/xago/v1/accounts/{accountId}/balance`

Get the current balance for a sub-account.

**Response**:
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440001",
  "balances": [
    {
      "currencyCode": "ZAR",
      "available": 5000.00,
      "reserved": 500.00,
      "total": 5500.00
    },
    {
      "currencyCode": "USD",
      "available": 1000.00,
      "reserved": 0.00,
      "total": 1000.00
    }
  ]
}
```

**Behavior**:
- Query balance from internal ledger (Pacioli)
- Return both available and reserved balances
- Used for UI display and transaction validation

---

### Transactions

#### 8. POST `/xago/v1/transfers`

Create a transfer from a Xago account to an external beneficiary.

**Request**:
```json
{
  "amount": 1000.50,
  "currencyCode": "ZAR",
  "beneficiaryId": "ben_ext_001",
  "reference": "Invoice #123",
  "idempotencyKey": "550e8400-e29b-41d4-a716-446655440002"
}
```

**Response**:
```
"550e8400-e29b-41d4-a716-446655440003"
```

(Returns transaction ID as plain string)

**Behavior**:
- Validate beneficiary exists and belongs to same wallet
- Debit the balance account by the amount
- Create transaction record with `status: "pending"`
- Return transaction ID (UUID)
- Idempotency: Same `idempotencyKey` should return same transaction ID
- In mock, automatically transition to `status: "completed"` after 2-3 seconds
- Do NOT send webhook in mock for outgoing transfers (wallet handles accounting)

**Status Codes**:
- `200 OK` - Transaction created
- `422 Unprocessable Entity` - Duplicate transaction (same idempotencyKey) - should return existing ID
- `401 Unauthorized` - Invalid token
- `400 Bad Request` - Invalid account, insufficient balance, etc.

---

#### 9. GET `/xago/v1/transactions?limit={limit}&page={page}`

List transactions for the authenticated account.

**Response**:
```json
{
  "meta": {
    "limit": 10,
    "pageNumber": 1,
    "numberOfPages": 1
  },
  "data": [
    {
      "transactionId": "550e8400-e29b-41d4-a716-446655440003",
      "status": "completed",
      "amount": 1000.50,
      "currencyCode": "ZAR",
      "beneficiaryId": "ben_ext_001",
      "reference": "Invoice #123",
      "createdAt": "2026-01-20T10:30:00Z",
      "settledAt": "2026-01-20T10:32:00Z"
    }
  ]
}
```

**Behavior**:
- Return paginated list of transactions for the account
- Filter by account if needed

---

#### 10. GET `/xago/v1/transactions/{transactionId}`

Get details of a single transaction.

**Response**:
```json
{
  "transactionId": "550e8400-e29b-41d4-a716-446655440003",
  "status": "completed",
  "amount": 1000.50,
  "currencyCode": "ZAR",
  "beneficiaryId": "ben_ext_001",
  "reference": "Invoice #123",
  "createdAt": "2026-01-20T10:30:00Z",
  "settledAt": "2026-01-20T10:32:00Z"
}
```

---

### Deposits (Webhook Triggered)

#### 11. POST `/xago/v1/company/accounts/testdeposit`

Simulate receiving a deposit (test endpoint).

**Request**:
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440001",
  "amount": 5000.00,
  "currencyCode": "ZAR",
  "depositReference": "USER_12345_ZAR"
}
```

**Response**:
```json
{
  "transactionId": "550e8400-e29b-41d4-a716-446655440004",
  "status": "pending"
}
```

**Behavior**:
- Create a deposit transaction in database
- Simulate incoming webhook after 1-2 seconds
- Webhook should POST to the configured `WEBHOOK_URL` with the webhook payload (see section below)
- In mock, automatically transition to `status: "completed"` and emit webhook

---

#### 12. GET `/xago/v1/company/transactions`

List deposits for the company account.

**Response**:
```json
{
  "meta": {
    "limit": 10,
    "pageNumber": 1,
    "numberOfPages": 1
  },
  "data": [
    {
      "transactionId": "550e8400-e29b-41d4-a716-446655440004",
      "status": "completed",
      "amount": 5000.00,
      "currencyCode": "ZAR",
      "accountId": "550e8400-e29b-41d4-a716-446655440001",
      "depositReference": "USER_12345_ZAR",
      "createdAt": "2026-01-20T10:00:00Z",
      "settledAt": "2026-01-20T10:05:00Z",
      "code": 104
    }
  ]
}
```

---

## Webhook Events

When deposits complete, Xago sends a webhook to the wallet's registered webhook URL.

### Deposit Webhook

**Endpoint**: POST to `WEBHOOK_URL` (configured in environment)

**Request Headers**:
```
x-gatehub-app-id: xago-mock
x-gatehub-timestamp: 1674076800
x-gatehub-signature: <HMAC-SHA256 signature>
Content-Type: application/json
```

**Request Body**:
```json
{
  "accountId": "550e8400-e29b-41d4-a716-446655440001",
  "amount": 5000.00,
  "currencyCode": "ZAR",
  "transactionId": "550e8400-e29b-41d4-a716-446655440004",
  "transactionReference": "USER_12345_ZAR",
  "status": "completed",
  "code": 104,
  "createdAt": "2026-01-20T10:00:00Z",
  "settledAt": "2026-01-20T10:05:00Z",
  "isRequested": false,
  "isDuplicate": false,
  "requestData": {
    "amount": 5000.00,
    "currencyCode": "ZAR",
    "customRequestId": ""
  }
}
```

**Signature Calculation**:
```
signature = HMAC-SHA256(
  timestamp + method + path + body,
  secret_key
)
```

**Behavior**:
- Wallet receives webhook
- Looks up the `transactionReference` to find the sub-account
- Finds the corresponding balance linked account
- Creates transaction record with `type: deposit`, `state: completed`
- Credits the balance account
- Updates Pacioli ledger

---

## Workflows & Processes

### 1. User Onboarding (ZA)

**Preconditions**: User completes KYC via Persona

**Workflow**:
1. Persona KYC inquiry completes with `status: approved`
2. Wallet KYC status → `StatusApproved`
3. Trigger: `SetKYCStatusWorkflow` (Temporal workflow)
4. Activity: `CreateKYCWallets` detects country is ZA
5. Activity: Check if Xago sub-account already exists (lookup by wallet_id)
6. If not exists:
   - Activity: Call Xago `CreateBalanceAccountWorkflow`
   - Activity: Create sub-account via `CreateSubAccount` activity
     - Fetch user details (name, email, phone)
     - Fetch KYC details (address, ID number)
     - Fetch approved Persona inquiry URL
     - Call Xago POST `/company/accounts`
   - Activity: Save sub-account to database
   - Activity: Create balance linked account via `CreateBalanceAccountWorkflow`
     - Create linked account record (provider=xago, type=balance, currency=ZAR)
     - Configure Pacioli ledger account for this linked account
   - Activity: Update Xago inquiry link with approved Persona URL
7. If exists:
   - Activity: Update inquiry link only
8. Workflow complete, user has active ZAR balance account

**Temporal Workflows to Implement**:
- `CreateBalanceAccountWorkflow(args CreateBalanceAccArgs) → LinkedAccount`
- `UpdateInquiryLinkWorkflow(args InquiryLinkUpdate) → error`

---

### 2. Adding Beneficiary (Withdrawal Setup)

**Preconditions**: User has active Xago balance account

**User Journey**:
1. User clicks "Add Bank Account" in wallet UI
2. User fills in bank details (account number, branch, bank name, currency)
3. Request: `POST /linkedaccounts/add-beneficiary` (wallet backend endpoint)
4. Wallet calls: Xago `CreateBeneficiary` operation
5. Temporal Workflow: `CreateBeneficiaryWorkflow(args CreateBankAccountArgs)`
   - Activity: Fetch sub-account for wallet
   - Activity: Call Xago POST `/beneficiaries` with bank details
   - Activity: Save beneficiary to local database
   - Activity: Create linked account (provider=xago, type=bank_account)
6. Workflow returns linked account
7. User sees new bank account in UI with `status: pending` → `approved` (2-3 sec)

---

### 3. Sending Money

**Preconditions**: User has balance account with funds and at least one approved beneficiary

**User Journey**:
1. User enters amount and selects destination beneficiary
2. Request: `POST /transactions` (wallet backend endpoint)
3. Wallet validates balance is sufficient
4. Wallet calls: Xago `CreateTransaction` operation
5. Operation steps:
   - Generate idempotency key (transaction ID)
   - Call Xago POST `/transfers` with beneficiary ID and amount
   - Xago deducts balance and returns transaction ID
6. Wallet:
   - Reserves balance in Pacioli
   - Creates local transaction record
   - Waits for completion webhook (or polling)
7. Mock Xago:
   - Simulates transfer processing (2-5 seconds)
   - Auto-completes transaction
   - Mock does NOT send webhook (wallet handles outgoing transfers differently)

---

### 4. Receiving Deposits

**User Journey**:
1. User views wallet balance screen
2. UI displays deposit details:
   - Bank account number: `62057334567` (FNB)
   - Branch code: `250145`
   - Account name: `Xago Holdings`
   - Reference: `USER_{walletID}_ZAR` (from sub-account creation)
3. User initiates real bank transfer to Xago account with reference
4. Bank transfer clears (IRL: 2-3 business days, mock: immediately via test endpoint)
5. Xago mock:
   - Receives deposit (via POST `/company/accounts/testdeposit` in testing)
   - Creates deposit transaction
   - Sends webhook to wallet's webhook URL after delay
6. Wallet:
   - Receives webhook on POST `/xago/webhooks/event`
   - Validates webhook signature
   - Looks up sub-account by `accountId` from webhook
   - Finds corresponding balance account (by wallet_id and currency)
   - Creates transaction with `type: deposit`, `state: completed`
   - Credits balance account via Pacioli `CreateTransfers`
   - User sees updated balance in UI

---

## Storage Implementation

### In-Memory Storage (for local testing)

The mock should provide an in-memory implementation similar to MockGatehub:

```go
type MemoryStore struct {
    mu              sync.RWMutex
    subAccounts     map[string]*SubAccount          // By wallet_id
    beneficiaries   map[string][]*Beneficiary       // By wallet_id
    transactions    map[string]*Transaction         // By ID
    balances        map[string]map[string]float64   // [wallet_id][currency] → amount
    tokens          map[string]*AccessToken         // By token value
    accessToken     *AccessToken                    // Single shared token
}
```

### Redis Storage (for production-like testing)

For integration testing with real Redis:
- `subaccount:{wallet_id}` → JSON
- `beneficiaries:{wallet_id}` → JSON array
- `transaction:{id}` → JSON
- `balance:{wallet_id}:{currency}` → float64
- `xago:token` → JSON with expiration

---

## Configuration & Environment

**Required Environment Variables**:

```bash
# Xago API Credentials
XAGO_API_PUBLIC_KEY=your_public_key
XAGO_API_SECRET=your_secret_key

# Webhook Configuration
WEBHOOK_URL=http://localhost:8080/xago/webhooks/event
WEBHOOK_SECRET=webhook_secret_for_hmac

# Service Configuration
XAGO_MOCK_PORT=8080
XAGO_MOCK_REDIS_URL=redis://localhost:6379  # Optional, use memory if absent
XAGO_MOCK_REDIS_DB=1                        # Redis database number

# Timing Configuration (for realistic simulation)
XAGO_DEPOSIT_DELAY_MS=2000      # Delay before deposit webhook
XAGO_TRANSFER_DELAY_MS=3000     # Delay before transfer completion
```

---

## Testing Strategy

### Unit Tests

Test each handler independently with table-driven tests:

```go
func TestCreateSubAccount(t *testing.T) {
    tests := []struct {
        name       string
        req        SubAccountReq
        wantStatus int
        wantErr    string
    }{
        {
            name: "valid sub-account",
            req: SubAccountReq{
                FirstName: "John",
                LastName: "Doe",
                Email: "john@test.com",
            },
            wantStatus: 200,
        },
        {
            name: "missing first name",
            req: SubAccountReq{
                LastName: "Doe",
                Email: "john@test.com",
            },
            wantStatus: 400,
            wantErr: "firstName is required",
        },
    }
    // ...
}
```

### Integration Tests

Full workflow tests using `httptest`:

```go
func TestDepositWorkflow(t *testing.T) {
    // 1. Create sub-account
    // 2. Create balance account
    // 3. Send test deposit
    // 4. Verify balance updated
    // 5. Verify webhook received
}
```

### End-to-End Tests

Test against actual wallet backend:

```bash
# Start Xago mock
./xago-mock

# Start wallet backend with XAGO_MOCK_URL=http://localhost:8080

# Run wallet integration tests
cd /path/to/wallet && go test ./...
```

---

## Differences from Real Xago API

The mock service intentionally simplifies or differs from production Xago:

| Aspect | Real Xago | Mock Xago |
|--------|-----------|----------|
| **Authentication** | OAuth with client credentials | Simple token endpoint with basic validation |
| **KYC Verification** | Third-party verification required | Auto-approved (Persona URL stored but not validated) |
| **Beneficiary Approval** | 24-48 hour approval | Auto-approved after 2-3 seconds |
| **Deposit Processing** | 2-3 business days | Immediate via test endpoint |
| **Transfer Execution** | 1-2 business days | Immediate (2-5 second simulation) |
| **Error Handling** | Complex error codes | Simplified error responses |
| **Rate Limits** | API rate limits enforced | No rate limits |
| **Compliance** | Full regulatory checks | Minimal compliance checks |

---

## Directory Structure

```
xago-mock/
├── cmd/xago-mock/
│   └── main.go              # HTTP server setup, routing
├── internal/
│   ├── auth/
│   │   ├── middleware.go    # Token validation
│   │   └── signature.go     # HMAC signature generation
│   ├── handler/
│   │   ├── handler.go       # Handler struct & dependencies
│   │   ├── auth.go          # Login endpoint
│   │   ├── accounts.go      # Sub-account endpoints
│   │   ├── beneficiaries.go # Beneficiary endpoints
│   │   ├── transactions.go  # Transaction endpoints
│   │   ├── deposits.go      # Deposit endpoints
│   │   ├── *_test.go        # Unit tests for each handler
│   │   └── handlers_test.go # Integration tests
│   ├── models/
│   │   └── models.go        # All data structures
│   ├── storage/
│   │   ├── interface.go     # Storage contract
│   │   ├── memory.go        # In-memory implementation
│   │   ├── redis.go         # Redis implementation
│   │   └── *_test.go        # Storage tests
│   ├── webhook/
│   │   ├── manager.go       # Async webhook sender
│   │   └── manager_test.go
│   ├── consts/
│   │   └── consts.go        # Constants (bank details, etc.)
│   ├── utils/
│   │   └── utils.go         # Utility functions
│   └── logger/
│       └── logger.go        # Logging setup
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Local development stack
├── go.mod
├── go.sum
├── README.md               # User documentation
└── AGENTS.md               # Development guide for AI agents
```

---

## Implementation Priorities

### Phase 1: Core Endpoints (MVP)

1. **POST `/xago/v1/login`** - Token generation
2. **POST `/xago/v1/company/accounts`** - Create sub-account
3. **GET `/xago/v1/currencies`** - List currencies
4. **Storage**: In-memory implementation
5. **Tests**: Unit tests for handlers

### Phase 2: Beneficiaries & Transactions

6. **POST `/xago/v1/beneficiaries`** - Add beneficiary
7. **GET `/xago/v1/beneficiaries`** - List beneficiaries
8. **POST `/xago/v1/transfers`** - Create transfer
9. **GET `/xago/v1/transactions`** - List transactions

### Phase 3: Deposits & Webhooks

10. **POST `/xago/v1/company/accounts/testdeposit`** - Simulate deposit
11. **Webhook Manager** - Async webhook delivery
12. **GET `/xago/v1/company/transactions`** - List deposits
13. **Integration Tests**: Full workflow testing

### Phase 4: Polish & Documentation

14. **Redis Storage** - Production-like setup
15. **Docker Build** - Multi-stage Dockerfile
16. **Documentation** - API docs, setup guide
17. **Edge Cases** - Idempotency, rate limiting, error handling

---

## Key Implementation Notes

### Balance Management

Balances should be tracked in Pacioli (double-entry ledger system). Each balance account has:
- A debit account (money in Xago)
- A credit account (operations/fees)

Deposits and transfers adjust balances through journal entries.

### Webhook Idempotency

Webhooks should include transaction IDs that are idempotent. If the wallet receives the same webhook twice, it should not create duplicate transactions.

### Token Expiration

Access tokens expire after 55 minutes. The client (wallet) automatically refreshes by calling POST `/login` again. The mock should:
- Accept old tokens until they expire
- Reject expired tokens with 401
- Support force refresh via new login call

### Error Handling

Return standard HTTP status codes:
- `200 OK` - Success
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Invalid/missing token
- `422 Unprocessable Entity` - Duplicate idempotency key
- `500 Internal Server Error` - Unexpected error

---

## Similarity to MockGatehub

Like MockGatehub, the Xago mock should follow these patterns:

1. **Dual-Backend Storage**: Interface-based design allowing memory/Redis swap
2. **Middleware Authentication**: Reusable middleware for token validation
3. **Async Webhooks**: Non-blocking webhook delivery with retries
4. **Comprehensive Testing**: 80%+ coverage with table-driven tests
5. **Isolated Testenv**: Separate Docker compose for integration tests
6. **Clear Documentation**: Both user docs and agent development guide

---

## Success Criteria

The Xago mock service should:

1. ✅ Enable local wallet development without real Xago credentials
2. ✅ Provide realistic deposit/withdrawal flow simulation
3. ✅ Support both ZAR and USD balance accounts
4. ✅ Maintain data consistency across all operations
5. ✅ Support idempotent operations (safe retries)
6. ✅ Deliver webhooks reliably (with retry logic)
7. ✅ Pass all wallet integration tests
8. ✅ Achieve 80%+ code coverage
9. ✅ Support docker-compose deployment
10. ✅ Include comprehensive documentation for maintainers

---

## References

- Wallet backend: `/go/backend/providers/xago/`
- Existing mockbos: `/go/mockbos/`
- MockGatehub patterns: `/mockgatehub/`
- Wallet tests: `/go/backend/wallets/` and `/go/backend/linkedaccounts/`
- Temporal workflows: `/go/backend/temporal/worker.go`
