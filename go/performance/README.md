# performance

Transaction throughput and latency testing against a running wallet backend.

The harness speaks gRPC directly to the backend, so the same binary works against
local Compose and against a deployed environment reached through a port-forward.
It drives real payments through the real API: create, confirm, and then poll each
payment until it settles.

## Why this and not k6

k6 is a good load tool, and its gRPC support is real, but it measures the wrong
thing here. `CreatePayment` does not move money — it persists a payment and starts
a Temporal workflow:

```go
// go/backend/payments/ops/ops.go
_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, PaymentWorkflow, dbp.ID)
```

Payments then move `Created → Confirmed → Processing → Completed/Failed`
asynchronously. A load generator that reports RPC latency will happily claim
thousands of transactions per second while the Temporal task queue backs up and
real settlement time climbs into minutes.

Measuring the number that matters means tracking every payment's lifecycle to a
terminal state, holding state for hundreds of in-flight payments, and applying
backpressure when the backend stops keeping up. That is awkward in k6's VU model
and natural in Go — where the generated `BackendServiceClient` in
[go/proto/backend/v1/](../proto/backend/v1/) is already available and typed.

k6 would still be the right choice for load-testing the protea HTTP frontend or a
stateless read endpoint. It is not the right choice for this.

## Quick start (local)

```bash
cd local && make all          # backend, Kratos, Postgres, Temporal, provider mocks
cd ../go/performance

make provision COUNT=20 SENDERS=10    # creates wallets, writes wallets.local.yaml
make smoke                            # 2 payments per sender — checks the harness
make run                              # full drain run, many-to-many
```

`make provision` only makes sense against local development — see
[Provisioning](#provisioning) below. Against any other environment, list existing
wallets in an overlay file by hand.

## Scenarios

A scenario is one or more YAML files, deep-merged in order through
[configa](../configa/README.md), so later files override earlier ones. The
committed files in [scenarios/](scenarios/) describe the shape of a run; wallet
credentials go in a local overlay that is git-ignored:

```bash
./bin/perf run -config scenarios/many-to-many.yaml -config wallets.local.yaml
```

| Scenario | Shape | What it tells you |
|---|---|---|
| [smoke.yaml](scenarios/smoke.yaml) | 2 payments per sender, 1/s | The harness works |
| [many-to-many.yaml](scenarios/many-to-many.yaml) | N senders → N receivers, paired | Aggregate system capacity |
| [fan-in.yaml](scenarios/fan-in.yaml) | N senders → 1 receiver | Per-wallet contention |

Running both of the latter two is the point: if throughput collapses under fan-in
but not many-to-many, the limit is per-wallet serialisation rather than overall
capacity.

### Pairing modes

| `run.pairing` | Behaviour |
|---|---|
| `index` | `senders[i]` → `receivers[i]`. Requires equal-length lists. |
| `fan_in` | Every sender → `receivers[0]`. |
| `round_robin` | Each sender rotates through the full receiver list, staggered by index. |
| `random` | A receiver chosen uniformly at random per payment. |

### Stop conditions

| `run.stop` | Behaviour |
|---|---|
| `drain` | Send until the source wallet is empty. Capped at `balance / amount` payments, and stopped early by an insufficient-funds error. |
| `count` | `run.count_per_sender` payments each. |
| `duration` | Send for `run.duration`. |

`run.duration` also acts as a safety ceiling for `drain` and `count`.

### Load profile

- `run.arrival_rate` — target payments/sec across **all** senders combined. Zero
  means unthrottled, which finds the ceiling. Set a fixed rate when comparing
  latency between two builds, since latency at an unthrottled rate is not
  comparable across runs.
- `run.max_in_flight` — confirmed-but-not-settled cap. This is the backpressure
  that keeps the harness honest.
- `settlement.track` — leave this on. With it off, the reported latency is RPC
  round-trip time only and the report says so.

## Reports

Console output splits latency by stage, so an RPC regression is not mistaken for a
settlement regression:

```
── throughput ──
                       count  per second
confirmed              4821   80.4
settled (completed)    4790   79.9
never settled          31     0.5

── latency ──
stage                        n      min     mean    p50     p95     p99     max
create                       4821   8.1ms   24.3ms  19.2ms  61.4ms  118ms   402ms
confirm                      4821   11.4ms  38.7ms  31.0ms  94.2ms  201ms   755ms
settle (confirm → terminal)  4790   201ms   1.42s   1.11s   3.87s   6.20s   14.8s
```

Percentiles are exact — every sample is kept, not bucketed.

### Grafana

Set `metrics.listen` and the run exposes `/metrics` for the local Prometheus, so
perf numbers land on the same dashboards as backend metrics and Tempo traces:

```bash
cd local && make monitoring      # Prometheus, Tempo, Grafana on :3005
```

The harness runs on the host, not in Compose, so add a scrape target to
[local/config/prometheus/prometheus.yml](../../local/config/prometheus/prometheus.yml):

```yaml
  - job_name: perf
    metrics_path: /metrics
    static_configs:
      - targets: ['host.docker.internal:9464']
```

Exported series, all labelled with `job_name` from `metrics.job_label`:

| Metric | Meaning |
|---|---|
| `perf_rpc_duration_seconds{stage}` | Per-RPC latency (native histogram) |
| `perf_settlement_duration_seconds` | Confirm → terminal state |
| `perf_payments_total{outcome}` | `completed`, `failed`, `timed_out`, `rejected` |
| `perf_errors_total{stage,class,code}` | Failures by backend `AppError` code |
| `perf_payments_in_flight` | Confirmed but not settled |
| `perf_senders_active` | Senders still issuing payments |

## Deployed environments

Port-forward the backend's gRPC port and Kratos, then point the scenario at the
local ends of the tunnels. Nothing else changes:

```bash
kubectl port-forward -n <namespace> svc/backend 8443:8443 &
kubectl port-forward -n <namespace> svc/kratos  4433:4433 &

./bin/perf run -config scenarios/many-to-many.yaml -config wallets.staging.yaml
```

Two things to get right:

- **Use `connections: 8` or more.** One HTTP/2 connection through a port-forward
  is a hard bottleneck, and you will measure the tunnel instead of the backend.
- **Never run `perf provision` against a shared environment.** It creates real
  identities, wallets and Rafiki payment pointers.

## Things that will bite you

**KYC limits end a drain run before the wallet empties.** `CreatePayment`,
`UpdatePayment` and `ConfirmPayment` all call `ExceedsKYCLimits`, covering
per-transaction, daily, monthly, 6-monthly and yearly caps. A wallet with a large
balance will hit a daily limit long before it drains. The harness classifies this
as `exhausted` rather than as an error, and the per-sender table shows it as the
stop reason — but if you want a genuine drain, raise the limits for the perf
wallets first.

**Both wallets must be KYC approved.** `CreatePayment` rejects an unapproved
receiver outright. Setup calls `GetPaymentAddress` for every receiver first and
fails fast with a clear message, rather than letting the run fill with identical
rejections.

**Perf wallets must not have TOTP enrolled.** Kratos is configured with
`session.whoami.required_aal: highest_available`, so a password login yields
`aal1` and every RPC for that wallet then fails with an AAL2 error. Setup catches
this via the same `ToSession` call the backend makes.

**Fees mean `balance / amount` is an upper bound.** The authoritative stop is the
backend's insufficient-funds error, which the harness treats as a clean finish.

## Provisioning

`perf provision` creates local development wallets and writes a ready-to-use
overlay. It runs the signup flow the way protea does — `SetSignupUserData`, a
native Kratos registration, `CompleteSignup`, `CreateUserDefaultWallet`,
`CreateWalletAddress` — and then does two things the product deliberately does not
expose:

- **KYC approval**, written directly into `wallet_kyc_status`. There is no RPC for
  this because approval belongs to the KYC provider, and locally that means
  driving a provider mock through the UI. The same direct-write shortcut already
  exists for email verification in [local/scripts](../../local/scripts). Skipped
  when `-dsn` is empty, which leaves the wallets unable to transact.
- **Funding**, via `AddXagoBalanceAccount` plus the `DepositTestXago` RPC — which
  the backend itself refuses to serve when the environment mode is prod. ZA
  wallets only.

```bash
./bin/perf provision -count 200 -senders 100 -fund -deposits 5 -out wallets.local.yaml
```

Wallets with a balance are listed as senders first, since an unfunded sender has
nothing to drain. Anything that could not be completed is reported per wallet as a
note rather than silently skipped.

## Layout

| Path | Contents |
|---|---|
| [cmd/perf/](cmd/perf/) | CLI: `run`, `validate`, `provision` |
| [config/](config/) | Scenario models, loading and validation |
| [auth/](auth/) | Kratos native-login session tokens |
| [client/](client/) | gRPC connection pool, per-wallet clients, `AppError` classification |
| [runner/](runner/) | Setup, sender loops, settlement watcher |
| [metrics/](metrics/) | Prometheus collectors, exact-percentile recorder, console report |
| [provision/](provision/) | Local-only wallet bootstrap |
| [scenarios/](scenarios/) | Committed scenario files |
