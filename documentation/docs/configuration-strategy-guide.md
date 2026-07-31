# Configuration Strategy Guide

> **How we add configuration safely.** This page describes the strategy every developer should follow when introducing new configuration to the wallet. New features ship *inert* and are switched on through configuration, sensible defaults never break an existing environment, and bad configuration fails fast, either before it can reach a cluster or, at the latest, when the application starts.

**Related documents:**

- [Backend Configuration Guide](backend-configuration-guide.md) - The concrete YAML scheme, secret handling, and a full reference of every backend setting
- [Environment Mode Guide](environment-mode-guide.md) - How and when to branch behaviour on `environment.mode`, and how to keep non-production shortcuts out of production
- [Environment Variables](env-variables.md) - Runtime configuration for the Protea and Botanist frontends
- [Error Handling Guide](error-handling-guide.md) - How the application surfaces failures at runtime

**Quick Navigation:**

- **Why does this matter?** → See [The design goal](#the-design-goal-a-safe-release-flow)
- **What is the core rule?** → See [Ship features inert](#the-core-rule-ship-features-inert)
- **What if my feature is an exception?** → See [Exceptions](#exceptions)
- **I'm adding config to something already live.** → See [Adding configuration to live features](#adding-configuration-to-live-features)
- **What is a "good" default?** → See [Good defaults vs bad defaults](#good-defaults-vs-bad-defaults)
- **Where should bad config be caught?** → See [Failing fast on bad configuration](#failing-fast-on-bad-configuration)

---

## The design goal: a safe release flow

The purpose of this strategy is to maximise the flow of our release process by making it safe.

We want to upgrade production to the exact version that has been running on sandbox without worrying that some new behaviour will suddenly take effect. We also want to spin up test environments that have only a *selected* set of features enabled, in isolation, so we can safely run whichever experiment we choose.

This ties directly to how releases move through the project. Every merge to `main` produces a versioned build and chart. `dev1` is bumped automatically, while `sandbox` and `production` are promoted from a version that has already proven itself elsewhere (see the Release Process in [`.github/copilot-instructions.md`](https://github.com/interledger/interledger-app/blob/main/.github/copilot-instructions.md)). That promotion is only trustworthy if a new version does not change behaviour until we ask it to.

**As a general rule, any version merged into `main` should be considered safe to deploy to any environment.** New features stay dormant until they are explicitly activated through configuration.

---

## The core rule: ship features inert

Write new code so that a new feature must be activated through configuration. Until it is activated the feature stays inert. The new code path is present in the binary but is never exercised, so deploying the new version changes nothing observable.

This is already the dominant pattern in the backend. Integrations and behavioural switches are gated behind an `enabled`-style flag and stay off unless a config file turns them on:

| Flag | What it gates |
|---|---|
| `pti.enabled` | The PTI/Fiant provider integration |
| `plaid.enabled` | Plaid bank linking |
| `twilio.enabled` | Live Twilio Verify (when off, a no-op verifier is used) |
| `email.enabled` | Outgoing SendGrid email |
| `gatehub.enabled` | The GateHub integration |
| `rafiki.node_enabled` | Rafiki full-node event orchestration |
| `delete_account_enabled` | The account-deletion flow |

See the [Backend Configuration Guide](backend-configuration-guide.md#configuration-reference-startconfig) for the full set. When you add a feature, follow the same shape: default it **off**, and let each environment opt in through its own config overlay.

The payoff is threefold:

- **Deploying is decoupled from releasing.** The code can ship to every environment long before anyone decides to turn it on.
- **Rollout is per-environment.** A feature can be exercised on `dev1`, then `sandbox`, then `production`, one config change at a time, with no code change and no redeploy.
- **Experiments are cheap and isolated.** A throwaway environment can enable exactly one feature and nothing else.

---

## Exceptions

Some changes genuinely cannot ship inert. A change to behaviour that is already live everywhere, or a migration that must happen in lockstep with a deploy, are the usual examples.

Exceptions are allowed, but they are **not a solo decision**. When a change cannot follow the ship-inert rule:

- **Involve the whole team** in deciding how to handle the situation.
- **Plan a version-specific deployment strategy** for that release before it is merged. Agree the order of environments, any coordinated migration or config change, and the rollback path.

If you find yourself reaching for an exception, first double-check that the change really cannot be hidden behind a flag. Most can.

---

## Adding configuration to live features

Sometimes you are not adding a new feature, but a new *option* to a feature or behaviour that is already live in most environments. A new field appears in the config struct, and every existing environment's config file predates it.

Before you merge, ask one question:

> **Is there a sensible default for this option that will not break any existing environment?**

If the answer is yes, give the field that default so the already-running environments keep working untouched. If the answer is no, and the option is mission-critical with no safe value, then treat it like required configuration and validate it so that no environment can deploy without setting it (see [Failing fast](#failing-fast-on-bad-configuration)).

---

## Good defaults vs bad defaults

A **good default** is a value that is correct, or at worst harmless, in *every* environment, including production. It lets old config files keep working when a new field is introduced.

A **bad default** is a value that happens to be right for one environment (usually the developer's) but is wrong, or actively dangerous, in another. If the code becomes active in an environment that relied on the default, it will silently do the wrong thing: talk to the wrong provider, connect to a host that does not exist, or bypass a control.

```yaml
# Good defaults: safe everywhere
signature_version: 1
max_retries: 3
log_level: info
port: "8080"

# Bad defaults: right in one place, wrong (or dangerous) elsewhere
provider_base_url: "https://staging.someprovider.com"   # points prod at staging
redis_url: "redis://localhost:3322"                     # only exists on a laptop
```

The danger with the bad examples is exactly what happens when the code becomes active in an environment that did not override the default. A service that defaults its provider URL to staging will, in production, quietly send real traffic to a staging system. One that defaults Redis to `localhost` will fail to connect the moment it runs anywhere but a developer's machine.

The rule of thumb: if you cannot pick a value that is safe in production, there is no good default. Make the field required and validate it rather than shipping a value that only works somewhere.

Some fields are dangerous in the *opposite* direction, where a default that is convenient for local development would be a security hole in production. The backend guards these explicitly rather than trusting a default:

- `twilio.enabled` **must** be `true` in `prod`. The disabled path runs a no-op verifier that accepts any OTP.
- `persona.sandbox_fake_za_id` **must** be `false` in `prod`. It fabricates identity documents.

These guardrails are enforced at both layers described below. They are also an example of a broader pattern: whenever behaviour is relaxed outside production, production must be guarded so the relaxation cannot take effect there. The [Environment Mode Guide](environment-mode-guide.md) covers that pattern in full.

---

## Failing fast on bad configuration

When configuration *is* wrong, we want to find out as early as possible: ideally before it ever reaches a cluster, and at the very latest the instant the application tries to start. A badly configured service should never be allowed to serve traffic.

We catch bad configuration at two layers.

### Before deployment: Helm validation

We use Helm to render our deployment instructions into Kubernetes manifests. Helm can also **validate configuration while it renders**, so a broken configuration is rejected long before it reaches any cluster, because the release simply never templates.

This project already does this in [`helm/interledger-app/templates/validation.backend.yaml`](https://github.com/interledger/interledger-app/blob/main/helm/interledger-app/templates/validation.backend.yaml), which calls Helm's `fail` on invalid or missing values, and is exercised by `helm unittest` in the `helm-tests.yml` workflow. For example, it rejects an unknown `environment.mode`, an enabled integration with missing credentials, and the production guardrails above.

**When should I add Helm-level validation?**

- **When changing or moving an existing configuration option.** Use Helm validation to refuse a deployment that still carries the outdated shape, so an environment cannot deploy against config the new code no longer understands.
- **When adding mission-critical configuration that has no safe default.** Validate in Helm that the value is set, so *every* environment is forced to supply it before a deployment is even attempted.

Validation that lives in Helm is enforced in CI and at render time, so you can rely on it being caught well before any environment is touched.

### At application startup: sanitise and refuse to start

Helm validation only covers what Helm can see. The application must still **sanitise its own inputs**, with very few exceptions, and when the configuration it is given is invalid it must **refuse to start**.

The backend does this in [`go/backend/config/start.go`](https://github.com/interledger/interledger-app/blob/main/go/backend/config/start.go). `configa` validates the parsed config against the struct's `validate` tags, and `validateStart` enforces the conditional rules such as feature-flag dependencies and the production guardrails. A fatal configuration error stops the process rather than letting it run half-configured.

Refusing to start is deliberate. If a service starts anyway and only fails when the misconfigured path is first hit, you invite the classic failure where it crashes in the middle of the night because of a typo an operator made that afternoon. Failing fast at startup means the orchestration system (Kubernetes) never routes traffic to the bad pod. The rollout stalls on a pod that will not become ready, the previous healthy version keeps serving, and the problem surfaces immediately, at deploy time, to the person making the change.

### Keep the two layers in agreement

The Helm checks and the application checks are intentionally mirrors of each other, and the validation template comments even point back to the specific rules in `start.go`. Helm gives fast, pre-cluster feedback. The application is the authoritative backstop that guarantees a running process was configured correctly no matter how it was launched. When you add or change a rule in one layer, add or change it in the other, and cover both in their respective tests (`helm/interledger-app/tests/backend.validation_test.yaml` and the Go config tests).
