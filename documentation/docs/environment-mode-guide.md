# Environment Mode Guide

> **How and when to branch behaviour on `environment.mode`.** The backend runs in one of five modes, and some behaviour legitimately differs between them: relaxing a check that only makes sense to relax off-production, or pointing at a provider's sandbox instead of its live endpoint. This page is the canonical reference for that mechanism, and for the one rule that matters most: whenever you relax something outside production, make it impossible for that relaxation to happen in production.

**Related documents:**

- [Configuration Strategy Guide](configuration-strategy-guide.md) - The wider strategy: ship features inert, safe defaults, and fail-fast validation
- [Backend Configuration Guide](backend-configuration-guide.md) - The `environment.mode` field and every other backend setting
- [KYC & Identity Verification](kyc-guide.md) - One subsystem whose behaviour varies by mode

**Quick Navigation:**

- **What are the modes?** → See [The five modes](#the-five-modes)
- **How do I read the mode in code?** → See [Reading the mode in code](#reading-the-mode-in-code)
- **When should I branch on it?** → See [When it is appropriate to branch on the mode](#when-it-is-appropriate-to-branch-on-the-mode)
- **How do I keep production safe?** → See [Always guard production](#always-guard-production)
- **Mode check or feature flag?** → See [Mode or feature flag?](#mode-or-feature-flag)

---

## The five modes

`environment.mode` is a required backend setting, validated as one of `prod`, `sandbox`, `dev`, `local`, or `test`. It is the single behavioural switch that tells the backend what kind of environment it is running in.

| Mode | Where it runs | Character |
|---|---|---|
| `prod` | Production | Strictest. Real providers, the real key store, and all guardrails enforced. |
| `sandbox` | The public sandbox environment | Deliberately production-like. It selects providers' sandbox endpoints but keeps production-grade security, so a version proven on sandbox is safe to promote to production. |
| `dev` | The shared dev environment (`dev1`) | Integration environment for work in progress. |
| `local` | A developer's machine (Docker Compose) | The most relaxed. It allows `http`, generates keys directly in the database, skips Vault, and uses no-op verifiers so the stack runs with no external dependencies. |
| `test` | Automated tests (`go test`) | The same relaxations as `local`, selected automatically. |

Note that `config.IsTestExecution` returns true whenever the process is running under `go test`, even when the mode is not literally `test`. Use it for shortcuts that should only ever apply during automated testing.

---

## Reading the mode in code

Read the mode through the helper methods on `EnvironmentConfig`, never by comparing the raw string:

```go
if b.Config().Environment.IsModeLocal() {
    // local-only shortcut
}
```

The available helpers are `IsModeProd()`, `IsModeSandbox()`, `IsModeDev()`, `IsModeLocal()`, and `IsModeTest()`.

Comparing `Environment.Mode` against a string literal is a lint error. A ruleguard rule in [`go/gorules/rules.go`](https://github.com/interledger/interledger-app/blob/main/go/gorules/rules.go) rejects `x.Environment.Mode == "prod"` and points you at the helpers, so the set of valid modes stays defined in one place and every call site reads as intent rather than a magic string.

---

## When it is appropriate to branch on the mode

Branching on the mode is appropriate when the behaviour is tied to the *kind* of environment rather than to a feature you want to toggle on its own. The backend already does this in a handful of well-scoped places:

- **Infrastructure that does not exist outside deployed environments.** Custodial keys are generated with the real key store in deployed environments, but in `local` and `test` they are generated and stored directly in the database ([`go/backend/keys/ops/ops.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/keys/ops/ops.go)), and Vault encryption is skipped ([`go/backend/vault/vault.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/vault/vault.go)).
- **Transport security you can only relax locally.** GateHub API URLs must use `https` in every mode except `local`, where `http` is permitted so the mock services work ([`go/backend/providers/gatehub/ops/ops.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/providers/gatehub/ops/ops.go)).
- **Choosing a provider's sandbox rather than its live endpoint.** Chimoney and GateHub point at the provider's sandbox host in every non-production mode and at the live host only in `prod` ([`go/backend/providers/chimoney/ops/ops.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/providers/chimoney/ops/ops.go)).

Keep these branches small and obvious. A reader should be able to see, right at the branch, exactly what is being relaxed and why.

---

## Always guard production

The danger of mode-conditional behaviour is that a relaxation meant for `dev` or `local` silently takes effect in production. Whenever you skip a security check or take a shortcut outside production, make it impossible for that shortcut to run in production. There are two complementary techniques, and anything security-sensitive should use both.

**1. Write the branch so production is excluded by construction.** Gate the shortcut on `!IsModeProd()`, or on `IsModeLocal() || IsModeTest()`, never on the inverse. Written that way, the relaxed path simply does not exist for production, and adding a new mode later cannot accidentally opt it in.

**2. Fail startup if a dangerous setting could be active in production.** When the relaxation is driven by a separate config flag rather than by the mode itself, add a guardrail in `validateStart` ([`go/backend/config/start.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/config/start.go)) that rejects the dangerous combination when the mode is `prod`, and mirror it in the Helm validation so it is caught before deployment. The backend already guards two of these:

- `twilio.enabled` must be `true` in `prod`, because the disabled path runs a no-op verifier that accepts any OTP.
- `persona.sandbox_fake_za_id` must be `false` in `prod`, because it fabricates identity documents.

Both guardrails fail at `helm template` and again at backend startup, so a misconfigured production deployment never becomes ready. See [Failing fast on bad configuration](configuration-strategy-guide.md#failing-fast-on-bad-configuration) for how those two layers fit together.

The rule to remember: a non-production shortcut without a production guard is an incident waiting for the wrong config file.

---

## Mode or feature flag?

Both a mode check and a feature flag let behaviour vary between environments, so it helps to know which to reach for.

- **Use a feature flag** (an `enabled`-style config option that defaults off) when the behaviour is a feature you want to turn on and off independently, roll out one environment at a time, or test in isolation. This is the default approach, described in the [Configuration Strategy Guide](configuration-strategy-guide.md).
- **Use `environment.mode`** when the behaviour is intrinsic to the type of environment and would never be toggled on its own: relaxing a security requirement that only makes sense to relax off-production, or selecting real versus local infrastructure.

If you ever find yourself wishing you could enable a mode-driven behaviour in one specific environment but not in another of the same mode, that is the signal it should have been a feature flag.
