# entityconf

A generic, typed configuration mechanism that attaches named **confs** to
arbitrary entities (wallets today; other entity types later) without
requiring a schema migration to add or remove one.

> This package is self-contained: nothing else in the application
> constructs or calls it yet.

## Why a conf, not a flag or a column

Confs are declared in code, not in the database schema — adding or
removing one is a plain Go change (add or remove a struct field), never a
migration or a schema diff. Storage is sparse: no row is ever written for
an entity unless its value actually differs from the current default.
Reading a conf's value is ordinary, compile-checked Go — a real struct
field, not a stringly-typed lookup.

## Quick start

Declare one struct per entity type, tagging each field:

```go
package mypackage

import "github.com/interledger/interledger-app/go/backend/entityconf"

const EntityWallet entityconf.EntityType = "wallet"

type WalletConfs struct {
	SendEnabled    bool   `conf:"wallet.send_enabled" default:"true" display:"Send Payments" desc:"Allows a wallet to initiate outbound payments"`
	ReceiveEnabled bool   `conf:"wallet.receive_enabled" default:"true" display:"Receive Payments" desc:"Allows a wallet to receive incoming payments"`
	MaxDailyLimit  int    `conf:"wallet.max_daily_limit" default:"1000" display:"Max Daily Limit" desc:"Maximum amount a wallet can send per day"`
	Nickname       string `conf:"wallet.nickname" default:"" display:"Nickname" desc:"Optional display nickname for the wallet"`
}

func init() {
	entityconf.MustRegister(EntityWallet, WalletConfs{})
}
```

Four tags per field:

| Tag | Meaning |
|---|---|
| `conf` | The conf's key. Required on every exported field — use `conf:"-"` to explicitly opt a field out. Must be prefixed with `"<entityType>."` (e.g. `wallet.`). |
| `default` | The factory default, as a string, parsed into the field's actual Go type (`bool`, `int`, or `string`) when the struct is registered. |
| `display` | Short admin-portal label (`display_name`). |
| `desc` | Longer explanation (`description`). |

Register from an `init()` function so a malformed tag panics the moment the
binary starts, not at some later, harder-to-trace point.

Read an entity's confs with `Load`, resolving into a plain struct:

```go
var confs WalletConfs
if err := entityconf.Load(ctx, store, walletID, &confs); err != nil {
	return err
}
confs.SendEnabled // a real, typed bool — the wallet's own override if one
                  // exists, else the conf's current effective default
```

`Load` infers the entity type from the destination struct's registration —
you never pass it by hand.

## Storage

Persistence is abstracted behind the `Store` interface (`store.go`), with
two implementations sharing the exact same behavioral contract (see
`store_contract_test.go`, run against both):

- `NewInMemoryStore()` — fully in-memory, concurrency-safe, no database.
  Useful for tests and for anything not yet wired to a real database.
- `NewPostgresStore(db *sqlx.DB)` — backed by the `entity_confs` /
  `entity_conf_values` tables in `go/backend/db/schema.hcl`. Values and
  defaults are stored as `jsonb`, decoded according to each conf's declared
  `Type`. The caller owns the `*sqlx.DB`'s lifecycle.

Before `Load` can resolve anything, the registry's definitions must be
synced into the store once (mirrors what would happen at application boot):

```go
store := entityconf.NewPostgresStore(db) // or entityconf.NewInMemoryStore()
if err := store.SyncDefinitions(ctx, entityconf.Definitions()); err != nil {
	return err
}
```

`SyncDefinitions` is idempotent and safe to call repeatedly (e.g. every
process start): it refreshes `code_default`/type/labels for every known
key, seeds `effective_default` from `code_default` only the *first* time a
key is seen (so an admin-edited default is never clobbered by a later
sync), and marks any previously-synced key no longer present as
deprecated (un-deprecating it if it reappears).

Admin-style edits go through the `Store` directly:

```go
store.SetValue(ctx, EntityWallet, walletID, "wallet.send_enabled", false) // per-wallet override
store.ClearValue(ctx, EntityWallet, walletID, "wallet.send_enabled")     // revert to effective default
store.SetEffectiveDefault(ctx, "wallet.send_enabled", false)             // change the default for everyone without an override
```

## Sparsity

No row is ever written for an entity unless its value actually diverges
from the effective default. A direct, intentional consequence: changing the
effective default changes the resolved value for *every* entity that has
no explicit override — existing and new alike, not just future signups.

## Limitations, on purpose

- **`default` tags are parsed at registration time, not compile time.** A
  typo like `default:"tru"` on a `bool` field is a loud panic when
  `MustRegister` runs (e.g. the first time the package initializes — which
  in tests is the first time that `init()` fires), not a `go build` error.
  What *does* stay compile-time-checked: once loaded, a struct field like
  `confs.SendEnabled` is a real Go `bool` — using it wrong, renaming it, or
  removing it is an ordinary compile error.
- **Entity IDs are plain strings.** Nothing stops passing a wrong entity's
  ID into `Load` for a struct registered against a different entity type —
  you'd get a logically wrong result, not a compile error. `Load` does
  guarantee you can't target the wrong *entity_type*, since that's baked
  into the registration, not passed by the caller.
- **Every exported field must be tagged.** A field with no `conf` tag at
  all fails registration outright, rather than silently never becoming a
  conf. Use `conf:"-"` for fields that genuinely aren't confs.

## Package layout

| File | Contents |
|---|---|
| `entityconf.go` | `EntityType`, `ValueType` |
| `definition.go` | `Definition`, `StoredDefinition` |
| `registry.go` | `Registry`: `Register`/`MustRegister`, `Load`, `Definitions`/`DefinitionsFor` |
| `default.go` | Package-level convenience wrappers over a default `Registry` |
| `store.go` | The `Store` interface |
| `memory_store.go` | `NewInMemoryStore` |
| `postgres_store.go` | `NewPostgresStore`, backed by `entity_confs`/`entity_conf_values` |
| `errors.go` | Sentinel errors, all `errors.Is`-compatible |
| `store_contract_test.go` | Shared behavioral contract run against every `Store` implementation |

Running the Postgres-backed tests locally requires a reachable Postgres
with the schema applied and the generated test migrations available — see
`go/backend/db`'s own test setup for how that's provided in this repo.
