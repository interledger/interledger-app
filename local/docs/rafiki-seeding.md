# Rafiki Seeding Script

## Overview

The `rafiki-setup.go` script is an automated configuration tool that initializes a local Rafiki v1.2 instance with the assets and liquidity required for wallet operations. This script eliminates the need for manual configuration through the Rafiki Admin UI or direct GraphQL API calls.

## Purpose

Rafiki requires several resources to be configured before it can facilitate Interledger payments:

1. **Assets**: Currency definitions (USD, EUR, GBP, etc.) with their respective scales
2. **Asset Liquidity**: Funds deposited into each asset's liquidity account to enable transactions

Without these configured, the wallet backend cannot create payment accounts, process deposits, or initiate outgoing payments.

## Motivation

### Why This Script Exists

**Manual Setup Pain Points:**
- Configuring 10+ currencies manually through the admin UI is time-consuming
- GraphQL mutations require HMAC signature generation and proper header formatting
- Each asset needs liquidity deposited before it can be used
- Developers need to repeat this process every time they reset their local environment

**Script Benefits:**
- One-command setup: `go run rafiki-setup.go`
- Idempotent: Safe to run multiple times (checks for existing assets)
- Consistent: Ensures all environments have the same asset configuration
- Fast: Completes in seconds vs. minutes of manual work
- Type-safe: Go's strong typing prevents common configuration errors
History

This script was originally written in JavaScript for Rafiki v2, which introduced multi-tenant support. It has since been:

- **Rewritten in Go**: To match the repository's preferred language and provide better type safety
- **Adapted for v1.2**: Removed multi-tenant operations (v1.2 uses single-tenant architecture)
- **Removed tenant operations**: v1.2 uses a single-tenant architecture
- **Simplified authentication**: No `tenant-id` header required (defaults to single tenant)
- **Preserved core logic**: Asset creation and liquidity deposit mutations remain compatible

## How It Works

### 1. Environment Configuration

Thego
// Priority order:
1. Environment variables (highest priority)
2. local/.env file (loaded via godotenv)ghest priority)
2. local/.env file
3. Hardcoded defaults (lowest priority)
```

**Key Environment Variables:**
```bash
GRAPHQL_ENDPOINT=http://localhost:3001/graphql  # Rafiki Admin API endpoint
ADMIN_API_SECRET=your_signature_secret          # HMAC signing secret
ADMIN_SIGNATURE_VERSION=1                       # Signature algorithm version
```

### 2. HMAC Request Signing

All requests to the Rafiki Admin API must be signed using HMAC SHA-256:
go
// Signature format: t=<timestamp>, v<version>=<digest>
payload := fmt.Sprintf("%d.%s", timestamp, canonicalizedRequest)
h := hmac.New(sha256.New, []byte(adminAPISecret))
h.Write([]byte(payload))
digest := hex.EncodeToString(h.Sum(nil))
signature := fmt.Sprintf("t=%d, v%s=%s", timestamp, version, digest
const signature = HMAC-SHA256(payload, ADMIN_API_SECRET)
```

**Example Request Headers:**
```http
POST /graphql
Content-Type: application/json
signature: t=1706123456789, v1=a3f2b1c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0
```

The canonicalized request ensures consistent JSON formatting (sorted keys, no whitespace variations) to prevent signature mismatches.

### 3. Asset Configuration

The script ensures the following assets exist:

| Currency Code | Scale | Purpose                          |
|---------------|-------|----------------------------------|
| USD           | 2     | US Dollar (cents)                |
| EUR           | 2     | Euro (cents)                     |
| GBP           | 2     | British Pound (pence)            |
| ZAR           | 2     | South African Rand (cents)       |
| MXN           | 2     | Mexican Peso (centavos)          |
| SGD           | 2     | Singapore Dollar (cents)         |
| CAD           | 2     | Canadian Dollar (cents)          |
| EGG           | 2     | Test currency (EGG coins)        |
| PEB           | 2     | Test currency (PEB pebbles)      |
| PKR           | 2     | Pakistani Rupee (paisa)          |

**Scale**: The number of decimal places for fractional units (scale=2 means 1.00 USD = 100 cents).

**GraphQL Mutation:**
```graphql
mutation CreateAsset($input: CreateAssetInput!) {
  createAsset(input: $input) {
    asset {
      id
      code
      scale
    }
  }
}
```

**Variables:**
```json
{
  "input": {
    "code": "USD",
    "scale": 2
  }
}
```

### 4. Liquidity Seeding

Each asset receives an initial liquidity deposit of **100,000 units** (in major currency units):

- USD: $100,000.00 = 10,000,000 cents
- EUR: €100,000.00 = 10,000,000 cents
- etc.

**Why Liquidity Matters:**
Rafiki uses double-entry accounting with liquidity accounts. Before outgoing payments can be processed, the asset must have sufficient liquidity deposited. Think of it as the "float" or reserve that enables the system to operate.

**GraphQL Mutation:**
```graphql
mutation DepositAssetLiquidity($input: DepositAssetLiquidityInput!) {
  depositAssetLiquidity(input: $input) {
    success
  }
}
```

**Variables:**
```json
{
  "input": {
    "id": "uuid-for-transfer",
    "assetId": "asset-uuid-from-previous-step",
    "amount": "10000000",
    "idempotencyKey": "unique-uuid"
  }
}go
// For 100,000 units with scale=2:
baseAmount := int64(100000)
scale := int64(2)
multiplier := int64(1)
for i := int64(0); i < scale; i++ {
    multiplier *= 10
}
amount := baseAmount * multiplier
**Amount Calculation:**
```javascript
// For 100,000 units with scale=2:
const amount = BigInt(100000) * BigInt(10) ** BigInt(2)
// = 100000 * 100 = 10,000,000 (cents)
```

### 5. Idempotency
Go 1.25+ installed:**
   ```bash
   go version  # Should show go1.25 or higher
   ```

2. **Rafiki services running:**
   ```bash
   cd local
   docker compose --profile services up -d
   ```

3. **Environment variables (optional):**
   ```bash
   # Create local/.env if you need custom configuration
   cp .env.example .env
   # Edit .env with your values
   ```

### Running the Script

**From the scripts directory:**
```bash
cd local/scripts
go run rafiki-setup.go
```

**From the local directory:**
```bash
cd local
go run scripts/rafiki-setup.go
```

**From the project root:**
```bash
cd interledger-app
go run local/scripts/rafiki-setup.go
```

**Expected Output:**
```
Rafiki v1.2 admin endpoint: http://localhost:3001/graphql
Checking existing assets...
```

**From the project root:**
```bash
node local/rafiki-setup.js
```

**Expected Output:**

**Error: "no required module provides package"**
- Run `go mod download` in the `local/scripts` directory
- Ensure you have internet connectivity to download dependencies
```
Rafiki v1.2 admin endpoint: http://localhost:3001/graphql
Asset USD already exists
Asset EUR already exists
...
Ensuring asset liquidity...
Depositing liquidity for USD: 10000000 (scale 2)
Liquidity deposited for USD
...
✅ Rafiki configuration complete
```

### Troubleshooting

**Error: "Signature verification failed"**
- Ensure `ADMIN_API_SECRET` matches the Rafiki backend configuration
- Check that `ADMIN_SIGNATURE_VERSION` is set to `1`
- Verify the Rafiki backend is running and accessible

**Error: "Connection refused"**
- Confirm Rafiki backend is running: `docker compose ps rafiki_backend`
- Check the `GRAPHQL_ENDPOINT` URL (default: `http://localhost:3001/graphql`)
- Verify no port conflicts

**Error: "Asset already exists" (API error)**
- This is normal and expected behavior
- The script will continue with the next asset
- No action needed
 in `rafiki-setup.go`:**
   ```go
   var assetsToEnsure = []Asset{
     // Existing assets...
     {Code: "JPY", Scale: 0},  // Japanese Yen (no fractional units)
     {Code: "BTC", Scale: 8},  // Bitcoin (satoshis)
   }
   ```

2. **Run the script:**
   ```bash
   cd local/scripts
   go run rafiki-setup.gob deposits require matching asset codes (USD, EUR, etc.)
3. **Payments**: Outgoing payments require sufficient asset liquidity in Rafiki

**Workflow:**
```
User signs up → Wallet creates accounts → Rafiki wallet addresses created
                                  the `ensureLiquidity()` function:

```go
// Current: 100,000 units
baseAmount := int64(100000)

// Example: 1,000,000 units
baseAmount := int64(1000000
To support additional currencies:

1. **Update the asset list:**
   ```javascript
   const assetsToEnsure = [
     // Existing assets...
     { code: 'JPY', scale: 0 },  // Japanese Yen (no fractional units)
     { code: 'BTC', scale: 8 }   // Bitcoin (satoshis)
   ]
   ```
son
// Unsorted (will fail signature verification):
{"operationName":"CreateAsset","variables":{"input":{"code":"USD"}}}

// Canonical (sorted keys, no whitespace):
{"operationName":"CreateAsset","query":"...","variables":{"input":{"code":"USD","scale":2}}}
```

The `canonicalize()` function recursively sorts object keys to ensure

Edit the liquidity calculation in `ensureLiquidity()`:

```javascript
// Current: 100,000 units
const amount = BigInt(100000) * BigInt(10) ** BigInt(node.scale)

// Example: 1,000,000 units
const amount = BigInt(1000000) * BigInt(10) ** BigInt(node.scale)
```

**Note:** Liquidity deposits are additive. Running the script multiple times will increase the total liquidity.

## Technical Details

### Canonical JSON Formatting

HMAC signatures require deterministic JSON serialization:

```javascript
// Unsorted (will fail signature verification):
{"operationName":"CreateAsset","variables":{"input":{"code":"USD"}}}

// Canonical (sorted keys, no whitespace):
{"operationName":"CreateAsset","variables":{"input":{"code":"USD"}}}
```

The `canonicalize()` function ensures consistent formatting regardless of input order.

### Error Handling

The script uses a lenient error handling strategy:

- **Asset creation failures**: Logs warning, continues to next asset
- **Liquidity deposit failures**: Logs error, continues to next asset
- **Go Rewrite (Current)
- Rewrote from JavaScript to Go for better type safety and repository consistency
- Added proper error handling with typed structures
- Uses `github.com/google/uuid` for UUID generation
- Uses `github.com/joho/godotenv` for .env file loading
- Standalone executable with `go run` - no Node.js required

### v1.2 Compatibility
- Removed multi-tenant support (createTenant, updateTenant mutations)
- Removed `tenant-id` header from API requests
- Updated default GraphQL endpoint to port 3001 (v1.2 admin port)
- Updated script header comments and documentation

### v2 Original (JavaScript)
- Included tenant creation and update operations
- Required `OPERATOR_TENANT_ID`, `AUTH_IDENTITY_SERVER_SECRET`, `IDP_CONSENT_URL`
- Used tenant-scoped authentication
- Written in Node.js/JavaScript

---

**Last Updated:** January 26, 2026  
**Rafiki Version:** v1.2.0-beta  
**Script Location:** `local/scripts/rafiki-setup.go`  
**Language:** Go 1.25+
## References

- [Rafiki v1.2 Documentation](https://rafiki.dev/v1-beta/integration/requirements/overview)
- [Rafiki Admin API Reference](https://rafiki.dev/apis/graphql/backend/queries)
- [Interledger Accounting Concepts](https://rafiki.dev/v1-beta/overview/concepts/accounting)
- [HMAC Signature Specification](https://rafiki.dev/v1-beta/integration/requirements/webhook-events#verify-webhook-signatures)

## Changelog

### v1.2 Compatibility (Current)
- Removed multi-tenant support (createTenant, updateTenant mutations)
- Removed `tenant-id` header from API requests
- Updated default GraphQL endpoint to port 3001 (v1.2 admin port)
- Updated script header comments and documentation

### v2 Original
- Included tenant creation and update operations
- Required `OPERATOR_TENANT_ID`, `AUTH_IDENTITY_SERVER_SECRET`, `IDP_CONSENT_URL`
- Used tenant-scoped authentication

---

**Last Updated:** January 26, 2026  
**Rafiki Version:** v1.2.0-beta  
**Script Location:** `local/rafiki-setup.js`
