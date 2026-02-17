# Local Development Tool

CLI/TUI helper for local Interledger development. It manages Rafiki seeding, Kratos email verification, and TOTP codes.

## Install

```bash
cd local/scripts
make install
make build
```

## Commands

**Rafiki setup**

```bash
./local-dev-tool rafiki              # TUI select assets
./local-dev-tool rafiki --skip-ui    # create all assets for CI use
./local-dev-tool rafiki --assets USD,EUR,GBP
```

**Email verification (Kratos DB)**
Performs a database call to mark the user as verified.

```bash
./local-dev-tool verify alice@example.com
```

**TOTP code generation**

```bash
./local-dev-tool totp --secret XXXXXXXXXXXXXXX user@example.com
./local-dev-tool totp user@example.com
```

Notes:
- TOTP must be enabled in the wallet app first.
- From the second run onward, the CLI fetches the TOTP config from Kratos admin when available; otherwise it falls back to the locally stored secret.
- Local secrets are stored in `/tmp/local-dev-tool-totp-secrets.json`.

## Make Targets

```bash
make build          # Build the application
make test           # Run all tests
make coverage       # Generate coverage report
make clean          # Remove build artifacts
make all            # Install, test, and build
```

## Environment Variables

You can override some of the detault assumptions this tool make about the environment.

```bash
GRAPHQL_ENDPOINT=http://localhost:3001/graphql
ADMIN_API_SECRET=your_signature_secret
ADMIN_SIGNATURE_VERSION=1
```
