# Interledger App Wallet

The Interledger App Wallet is a multi-provider payment platform built on the Interledger network. It operates as an **Account Servicing Entity (ASE)**, connecting to payment providers in different countries to offer users a unified wallet experience across an open payment network.

It enables:

- **Cross-border payments** that are quick and easy
- **High-speed transactions** across currencies and providers
- **Content creator monetization** via Open Payments and Web Monetization

Users get a single wallet tied to one payment provider (GateHub, PTI, Xago, or Chimoney) based on their country, with linked accounts that can hold balances across different currencies.

---

## Documentation Guide

### Getting Started

| Document | Description |
|----------|-------------|
| [Core Concepts](terminology.md) | Terminology reference — maps provider vocabularies to Interledger App concepts |
| [Signup Flow](signup-guide.md) | The complete registration process: profile, phone verification, TOTP, wallet address |
| [KYC & Identity Verification](kyc-guide.md) | How identity verification gates wallet activation, per provider |
| [KYC Resubmission Flow (Phase 1)](kyc-resubmission-flow.md) | How users can resubmit KYC documents in-app (Phase 1) |

### Architecture

| Document | Description |
|----------|-------------|
| [Wallets, Accounts & Addresses](wallets-accounts-addresses-guide.md) | The three-layer model: wallet containers, linked accounts, and payment addresses |
| [Dual-Ledger System](ledger-system-guide.md) | Why we maintain Pacioli (internal) and provider (external) ledgers, and how they reconcile |

### Payments

| Document | Description |
|----------|-------------|
| [Payments & Transactions](payments-guide.md) | End-to-end guide: how money moves, reservation, settlement, and the payment lifecycle |
| [Transaction Types Reference](transaction-types-reference.md) | Field specifications for send, receive, deposit, withdrawal, and card transactions |
| [Provider Comparison](provider-payments-reference.md) | GateHub vs PTI vs Xago vs Chimoney — speeds, fees, capabilities, and special cases |

### Integrations

| Document | Description |
|----------|-------------|
| [GateHub Cards](gatehub-cards-guide.md) | EUR debit card issuing: customer onboarding, card lifecycle, 3DS, webhooks |
| [Plaid Bank Linking](plaid-integration-guide.md) | US bank linking via Plaid → PTI/Fiant: connect flow, soft-delete, concurrency, and dev issues |

### Operations

| Document | Description |
|----------|-------------|
| [Payment Troubleshooting](payment-troubleshooting-guide.md) | Systematic debugging framework for stuck payments, balance mismatches, and missing receipts |

### Developers

| Document | Description |
|----------|-------------|
| [Logging Reference](logging-reference.md) | Log levels, formatting standards, and structured logging guidelines |
| [Configuration Strategy Guide](configuration-strategy-guide.md) | How to add configuration safely: ship features inert, choose safe defaults, and fail fast on bad config in Helm and at startup |
| [Environment Mode Guide](environment-mode-guide.md) | How and when to branch behaviour on `environment.mode`, and how to keep non-production shortcuts out of production |
| [Backend Configuration Guide](backend-configuration-guide.md) | The Go backend's YAML config scheme, secret handling, and a full reference of every available setting |
| [Environment Variables](env-variables.md) | Environment variables for the Protea and Botanist frontends — secret classification, per-environment values, and local defaults |

---

## Quick Reference

```mermaid
graph TD
    User["User"]
    Signup["Signup & Auth<br/>(Ory Kratos)"]
    KYC["KYC Verification<br/>(Provider-specific)"]
    Wallet["Interledger Wallet"]
    LA["Linked Accounts<br/>(Balance, Bank, Card)"]
    Pacioli["Pacioli Ledger<br/>(Internal accounting)"]
    Providers["Payment Providers<br/>(GateHub, PTI, Xago, Chimoney)"]
    Rafiki["Rafiki<br/>(ILP Connector)"]
    OpenPay["Open Payments<br/>(Cross-provider)"]

    User --> Signup --> KYC --> Wallet
    Wallet --> LA --> Providers
    Wallet --> Rafiki --> OpenPay
    LA --> Pacioli

    style User fill:#1a3a5c,stroke:#0f2440,color:#fff
    style Wallet fill:#264d73,stroke:#1a3a5c,color:#fff
    style LA fill:#33608a,stroke:#264d73,color:#fff
    style Pacioli fill:#4074a1,stroke:#33608a,color:#fff
    style Providers fill:#4074a1,stroke:#33608a,color:#fff
    style Rafiki fill:#33608a,stroke:#264d73,color:#fff
    style OpenPay fill:#4074a1,stroke:#33608a,color:#fff
    style Signup fill:#33608a,stroke:#264d73,color:#fff
    style KYC fill:#33608a,stroke:#264d73,color:#fff
```