# Local Development Tool

A comprehensive CLI/TUI tool for managing the Interledger local development environment.

## Features

- **Rafiki Management**: Setup and seed Rafiki backend with currency assets and liquidity
- **Email Verification**: Manually verify user email addresses for local development
- **TOTP Code Generation**: Generate TOTP codes for two-factor authentication
- **Interactive TUI**: Select assets interactively with keyboard navigation
- **CLI Mode**: Automate operations with command-line flags

## Installation

```bash
cd local/scripts
make install
make build
```

## Usage

### Rafiki Setup

**Interactive mode** (TUI for asset selection):
```bash
./local-dev-tool rafiki
```

Navigate with arrow keys, space to toggle, `ctrl+s` to confirm.

**Skip UI** (create all assets):
```bash
./local-dev-tool rafiki --skip-ui
```

**Select specific assets**:
```bash
./local-dev-tool rafiki --assets USD,EUR,GBP,ZAR
```

### Email Verification

Manually verify a user's email address by directly updating the Kratos database:
```bash
./local-dev-tool verify alice@example.com
```

This command:
- Directly updates the Kratos database to mark the email as verified
- Sets `verified = true` and `status = 'completed'` in `identity_verifiable_addresses`
- Bypasses the email verification flow (useful for local development)

**Example output:**
```bash
./local-dev-tool verify alice@example.com
✅ Email verified successfully: alice@example.com

The user can now log in without email verification.
```

**Note**: This is intended for local development only. In production, users should complete the standard email verification flow.

### TOTP Code Generation

**Generate a TOTP code for login**:
```bash
./local-dev-tool totp user@example.com
```

This command:
1. **For users with TOTP configured in Kratos**: Retrieves the TOTP secret from Kratos and generates a valid 6-digit code
2. **For users without TOTP**: Prompts you to set up TOTP in the wallet application first, then enter the secret

**Example with existing TOTP** (already configured in Kratos):
```bash
./local-dev-tool totp dabatla@gmail.com
549662
```

**Example with new TOTP setup** (first time):
```bash
./local-dev-tool totp alice@example.com

⚠️  TOTP is not configured for this user.

📱 To set up TOTP:
   1. Open the wallet application and log in
   2. Go to Security Settings
   3. Enable Two-Factor Authentication (TOTP)
   4. Copy the secret shown on screen

🔐 Enter the TOTP secret from the wallet application:
Secret: JBSWY3DPEHPK3PXP

✅ Secret stored! Generated code:
   138759
```

**Subsequent runs** (secret is now stored locally):
```bash
./local-dev-tool totp alice@example.com
138759
```

**Important Notes**:
- **TOTP must be set up through the wallet application first** - the CLI cannot activate TOTP in Kratos directly
- This tool serves as a convenient code generator for users who have already configured TOTP
- TOTP codes change every 30 seconds according to RFC 6238 standard
- User-entered secrets are stored locally in `/tmp/local-dev-tool-totp-secrets.json` for convenience
- The CLI checks Kratos first, then falls back to local storage if the secret was entered manually

## Quick Reference

### Make Targets

```bash
make build          # Build the application
make test           # Run all tests
make coverage       # Generate coverage report
make coverage-html  # Open HTML coverage report
make clean          # Remove build artifacts
make run-rafiki     # Run Rafiki setup (interactive)
make all            # Install, format, test, and build
```

## What It Does

**Rafiki Setup:**
1. Creates selected currency assets in Rafiki
2. Deposits 100,000 units of liquidity per asset
3. Uses HMAC-signed GraphQL requests

## Supported Assets

USD, EUR, GBP, ZAR, MXN, SGD, CAD, EGG, PEB, PKR (all with scale 2)

## Documentation

For complete documentation and troubleshooting, see:
- [README.md](./README.md) - Full usage guide
- [../docs/rafiki-seeding.md](../docs/rafiki-seeding.md) - Rafiki setup details

## Migration from rafiki-setup.go

**Old way:**
```bash
go run rafiki-setup.go
```

**New way:**
```bash
./local-dev-tool rafiki --skip-ui
```

### Environment Variables

Configure via `local/.env` or environment variables:

```bash
GRAPHQL_ENDPOINT=http://localhost:3001/graphql
ADMIN_API_SECRET=your_signature_secret
ADMIN_SIGNATURE_VERSION=1
```

### First Run

The script will automatically download Go dependencies on first run:

```bash
go: downloading github.com/google/uuid v1.6.0
go: downloading github.com/joho/godotenv v1.5.1
```

This is normal and only happens once per machine.
