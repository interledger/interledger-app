# Wallets vs Accounts vs Addresses (Support Deep Dive)

> **Architecture deep dive.** Understand the three-layer model: wallet, linked accounts, and addresses.

**Related documents:**

- [Concepts](concepts.md) — Core terminology (this document extends Concepts)
- [Payments Guide](payments-explainer.md) — How addresses are used in payment flows
- [Provider Payments Guide](provider-payments-guide.md) — Provider-specific account structures
- [Ledger System Architecture](ledger-system-explainer.md) — How account choices affect reconciliation behavior
- [Transaction Types Reference](transaction-types-explainer.md) — How transactions reference linked accounts
- [Payment Troubleshooting](payment-troubleshooting-guide.md) — Practical routing and account-debug workflows
- [KYC Explainer](kyc-explainer.md) — Wallet activation and KYC impact on accounts

**Quick Navigation:**

- **What's the difference between wallet and account?** → See [System Model](#system-model-one-wallet-many-linked-accounts)
- **Why do providers differ?** → See [Provider Shape Differences](#provider-shape-why-linked-account-is-not-identical-across-integrations)
- **How are addresses used?** → See [Address Model](#address-model-wallet-urls-vs-ilpopen-payments-addresses)
- **Payment routing issues?** → See [Incident Triage Model](#incident-triage-model-for-support)

This document extends [concepts.md](./concepts.md) and explains how wallets, linked accounts, and addresses fit together in the Interledger App. It is written for senior technical support who need a reliable operational mental model when diagnosing payment-routing, currency, and provider-integration issues.

**Context:** The Interledger App Wallet is a reference implementation demonstrating how Account Servicing Entities (ASEs) can integrate with the Interledger network to enable cross-border payments, high-speed transactions, and content creator monetization.

At a high level, the system has three different layers that are often conflated:

- **Wallet layer (product identity)**: one wallet per user, user-facing identity anchor, Interledger network participant.
- **Linked-account layer (money movement rails)**: provider-specific accounts that actually send/receive value.
- **Address layer (discovery/Open Payments identity)**: wallet URLs/payment pointers used to target a wallet for Interledger payments.

Most support confusion comes from mixing these layers. A payment can be addressed to a wallet URL (Open Payments), but settlement still occurs through linked accounts.

---

## System model: one wallet, many linked accounts

In Interledger App, a wallet is intentionally a container, not a single balance bucket. A single wallet can hold multiple linked accounts across providers and currencies. That is how one user can have EUR and GBP availability (or ZAR and USD, etc.) while still having one wallet identity.

A linked account contains the routing details that matter for execution:

- provider (`gatehub`, `pti`, `xago`, `chimoney`, ...)
- account type (`balance`, `bank_account`, `card`, `interac`, ...)
- provider-specific identifier (`provider_id`)
- directional capability (`can_send`, `can_receive`)
- send/receive currency metadata (`send_currency`, `receive_currency`)
- state/default flags used by payment selection logic (`state`, `default_send`, `default_receive`)

Operationally, this means:

- one user wallet can legitimately expose many rails
- each rail can have different restrictions
- "wallet has funds" always means "at least one linked account has usable balance/capability"

So if a customer reports a currency issue, support should inspect linked-account records before wallet metadata.

---

## Provider shape: why "linked account" is not identical across integrations

The linked-account abstraction is shared, but provider semantics differ.

### Xago

Xago uses both `balance` and `bank_account` linked accounts.

- `balance` accounts are the in-app transacting rails.
- `bank_account` accounts are beneficiary payout rails (used primarily as withdrawal destinations).
- In practice, Xago P2P flows are balance-to-balance.
- Xago balance-account creation is constrained by currently supported account currencies in this integration path.

So yes: in Xago, linked bank accounts are absolutely linked accounts in Interledger App terms, but they are not general-purpose balance rails.

### PTI

PTI uses `balance`, `card`, and `bank_account` linked accounts. PTI bank accounts are modeled as ACH-style rails with broad send/receive metadata, and PTI-specific payment validations decide whether a given flow is valid.

### GateHub

GateHub is primarily modeled around `balance` (plus card-related integration domains), not around a generic `bank_account` linked-account rail in the same way Xago/PTI are.

### Chimoney

Chimoney uses `balance` plus `interac` as the payout-destination concept.

The support implication is simple: "bank account" is not a universal concept across providers even though everything is represented as linked accounts internally.

---

## Address model: wallet URLs vs ILP/Open Payments addresses

Wallet addresses in Interledger App and wallet addresses in Rafiki/Open Payments are tightly related but not the same abstraction.

### Interledger App wallet URL

- Stored as wallet address records and used for lookup/discovery in app flows.
- A wallet can hold multiple address records.

### Rafiki/Open Payments wallet address

- Provisioned in Rafiki and mapped to the wallet via payment-pointer ID.
- Used for Open Payments operations and webhook linkage.
- Created with a specific asset association.

This is why users/support may see "wallet address" in UI and "wallet address/payment pointer" in payment infrastructure logs: same user-facing target, different operational layer.

---

## How currency targeting works during payment creation

When sender and receiver each have multiple linked accounts, targeting a specific currency is mostly an account-selection problem.

Current payment behavior combines explicit selection and automatic resolution:

1. If neither side specifies linked-account IDs, backend attempts to auto-pair compatible balance accounts by currency.
2. If account IDs are provided, backend validates ownership/capability/type constraints.
3. If receiver account is absent or unusable, backend can fall back to a default receive account for the relevant sender currency.

In a practical example (receiver has EUR and GBP accounts), selecting/constructing an EUR payment usually drives account choice toward EUR-compatible linked accounts. But this is not magic wallet-level routing; it is linked-account resolution plus validation.

This also explains why users may occasionally perceive "wrong account selected": fallback/default logic can choose a different valid rail if explicit targeting was incomplete.

Cross-currency behavior is intentionally constrained in the current payment engine for standard P2P paths, so multi-currency availability does not imply unrestricted FX routing.

**Note:** Account selection also depends on [KYC status](kyc-explainer.md#4-compliance-reality-kyc-is-tied-to-transactions). See [Payment Story](payments-explainer.md#the-payment-story) for a walkthrough of the full payment flow.

---

## Open Payments single-asset exposure and the multi-currency wallet gap

A key architectural tension today is:

- internal wallets can hold multiple linked-account currencies
- public Open Payments address provisioning is asset-scoped per Rafiki wallet address

In other words, one address advertises one asset context, while the wallet behind it may support more currencies via linked accounts.

For support, this means incoming/outgoing Open Payments currency behavior must be interpreted through asset-code-to-linked-account matching, not by assuming an address is a universal pointer to all wallet currencies.

For product evolution, if the goal is explicit per-currency and/or per-persona targeting, the likely direction is:

- multiple active Open Payments addresses per wallet/persona
- explicit address-to-linked-account policy mapping
- clearer sender-visible intent controls to reduce fallback ambiguity

---

## Incident triage model for support

When a report says "wrong currency", "wrong account", or "payment went to unexpected destination", diagnose in this order:

1. **Identity layer**: sender/receiver wallet IDs and address identifiers used.
2. **Execution layer**: persisted `sender_account` and `receiver_account` on the payment.
3. **Currency layer**: sender/receiver amount currency plus linked-account send/receive currencies.
4. **Selection path**: explicit account choice vs auto-pairing vs default-receive fallback.
5. **Provider constraints**: account type/provider compatibility rules for that payment type.
6. **Open Payments context**: whether asset-scoped address behavior influenced account matching.

Using this sequence keeps analysis grounded in how the system actually routes funds and prevents terminology-based misdiagnosis.

---

## See Also

- [Concepts: Interledger App vs Service Providers](concepts.md) — Core terminology and provider translation table
- [Payments & Transactions](payments-explainer.md) — Payment lifecycle and the two-ledger system
- [KYC Explainer](kyc-explainer.md) — How identity verification gates payment authorization
- [GateHub Cards](gatehub-cards-explainer.md) — Card-specific linked accounts (EUR card type)
