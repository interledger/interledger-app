# mockplaid

Local mock of [Plaid](https://plaid.com) for the Interledger wallet POC. Stands in for
both Plaid surfaces the app touches:

- **Plaid REST API** — the server-to-server endpoints the backend SDK (`plaid-go/v42`) calls.
- **Plaid Link** — the hosted `cdn.plaid.com` SDK + Link UI the browser loads.

So the full link → exchange → processor-token → Fiant flow runs **offline and deterministically**,
with no real Plaid credentials or network egress.

> Tasks / per-task notes: `documentation/poc/mockplaid/`. How-it-works + diagrams:
> `documentation/poc/plaid/mockplaid.md`.

## How it plugs in

| Layer | Real Plaid | Mock |
|---|---|---|
| Backend SDK | `sandbox.plaid.com` | `PLAID_API_URL=http://mockplaid:8080` (backend env override) |
| Browser SDK + Link UI | `cdn.plaid.com` | `cdn.plaid.com` → traefik → mockplaid (via `/etc/hosts` + self-signed cert) |

No application code branches on mock-vs-real — only the network target changes (same pattern as
mockgatehub). The frontend keeps using `react-plaid-link` unchanged; mockplaid serves a `window.Plaid`
shim at the exact CDN path.

## The two mock banks (determinism)

| Bank | `institution_id` | Account-id behaviour | Tests |
|---|---|---|---|
| **Tartan Bank** | `ins_mock_a` | **Stable** per account (`acc_mock_a_checking` / `acc_mock_a_savings`) | reconnect / duplicate detection |
| **Platypus Bank** | `ins_mock_b` | **Always-new** (`acc_mock_b_<seq>`) | multi-account / stress |

`account_id` is fixed at selection time and read back by every later call, so the guarantee holds
across exchange / accounts / processor.

## Endpoints

Plaid REST (called by the backend SDK):

| Method | Path | Purpose |
|---|---|---|
| POST | `/link/token/create` | mint link token |
| POST | `/item/public_token/exchange` | public_token → access_token + item_id |
| POST | `/item/get` | item → institution_id |
| POST | `/institutions/get_by_id` | institution name |
| POST | `/processor/token/create` | processor token (for Fiant) |
| POST | `/item/remove` | revoke item (idempotent) |
| POST | `/accounts/get` `/auth/get` `/accounts/balance/get` `/identity/get` `/transactions/sync` | product reads (static) |

Browser / mock-only:

| Method | Path | Purpose |
|---|---|---|
| GET | `/link/v2/stable/link-initialize.js` | `window.Plaid` shim |
| GET | `/link?link_token=…` | mock Link dropdown UI (2 banks × 2 accounts) |
| POST | `/link/session/select` | mock-only: UI picks bank+account → mints public_token (stand-in for Plaid's hosted-iframe token mint) |

Ops:

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | `{"status":"ok"}` |
| POST | `/test/reset` | wipe state (e2e isolation) |

## Configuration

| Var | Default | Notes |
|---|---|---|
| `MOCKPLAID_PORT` | `8080` | HTTP port |
| `LOG_LEVEL` | `info` | zap level |
| `MOCKPLAID_REDIS_URL` | _(empty)_ | set → Redis store; empty → in-memory |
| `MOCKPLAID_REDIS_DB` | `6` | Redis DB index |

## Run / test

```bash
make build        # compile
make unit-test    # unit + storage contract (memory + miniredis)
make e2e-test     # godog suite (health, link-token, determinism, exchange)
make lint         # gofmt + go vet + golangci-lint
```

Local stack: brought up by `local/docker-compose.yaml` (service `mockplaid`, profile `application`).

## Mock ⇄ real toggle

```bash
cd local
make plaid-mock   # backend → mockplaid + cdn.plaid.com hosts redirect (default)
make plaid-real   # backend → sandbox.plaid.com + remove hosts redirect
```
`plaid-real` also needs real `BACKEND_PLAID_CLIENT_ID/SECRET` in `local/.env`.

## Notes / nuances

- `cdn.plaid.com` is **HSTS-preloaded** in browsers → the self-signed cert must be **trusted**
  (`cd local && make hosts && make certs && make trust`), not click-through bypassed.
- Webhooks: **none** (the Plaid POC has no webhook handling).
- The mock never validates a processor token against real Plaid (a mock trusts it); the real
  Fiant does, which is why local PTI must also point at **mockpti**, not real Fiant.
