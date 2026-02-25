# E2E Testing Guide for AI Agents

## Running Tests

Run from `e2e/`. Select scenarios by tag:

```bash
go test -v -timeout 5m  -args -tags @signup
go test -v -timeout 10m -args -tags "@kyc&&@xago"
go test -v -timeout 10m -args -tags "@deposit&&@gatehub"
go test -v -timeout 10m -args -tags "@deposit&&@xago"
go test -v -timeout 10m -args -tags "@p2p-payment&&@gatehub"
go test -v -timeout 10m -args -tags "@p2p-payment&&@xago"
go test -v -timeout 10m -args -tags "@withdrawal&&@fees"
```

Available tags: `@signup`, `@kyc`, `@deposit`, `@p2p-payment`, `@withdrawal` combined with `@gatehub`, `@xago`, `@fees`, `@germany`, `@skip`.

**Do not suppress test output** — visibility is critical for debugging.

## Environment

- Start with `make all-nowatch` from `local/` (not `make all`).
- Mock services use public URLs: `https://mockgatehub.interledger.test`, `https://mockxago.interledger.test`.
- Feature backgrounds must include the relevant `mockgatehub is running at` / `mockxago is running at` steps.
- Investigate Temporal jobs: `docker compose exec temporal temporal workflow list`

## Key Concepts

- **Users are unique per test run** — database cleanup has limited value.
- **Phone numbers**: E.164 format required (`+49987654321`). Tests generate randomized numbers. Failures usually mean duplicate or wrong format in Kratos.
- **KYC by country**: Germany → MockGatehub iframe; South Africa → MockXago Persona-like iframe (`/kyc/iframe`, step: `I fill and submit the mockxago KYC iframe`).
- **Deposits**: GateHub uses an iframe widget; Xago uses EFT simulation (`I simulate a "<amount>" "<currency>" EFT payment to Xago`).
- **Withdrawals**: `/withdraw` loads a MockGatehub iframe (GateHub only; Xago withdrawal not yet supported, tagged `@skip`).
- **Wallet stability**: `waitForStableWalletCount` in `db_helpers.go` polls the DB to avoid race conditions during wallet creation.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Kratos `tel` validation errors | `make reset` in `local/` to clear state |
| Random failures after many test runs (200+ Kratos identities) | `make reset` in `local/` |
| `@kyc` stuck on `/wallet-address` ("not in Reserved state") | Check backend logs: `docker compose logs backend \| grep -i "wallet\|error"` |
| Balance not appearing after deposit | Verify webhook delivery in mock service logs; check Temporal workflow completion |

## Adding Tests

1. Create or update a `.feature` file in `features/`.
2. Implement step definitions in the appropriate Go file (see existing patterns).
3. Reuse steps where possible — see `STEP_REFERENCE.md`.
4. Tag scenarios for selective execution.

## Maintain

It is the job of the agent to add, update or remove relevant information to this file.