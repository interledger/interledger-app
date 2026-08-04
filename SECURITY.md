# Security Policy

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
pull requests, or discussions.**

Report them privately through GitHub's
[private vulnerability reporting](https://github.com/interledger/interledger-app/security/advisories/new)
on this repository. That opens a draft security advisory visible only to you
and the maintainers, and it is the fastest way to reach us.

Please include as much of the following as you can:

- The type of issue (for example: authentication bypass, injection, privilege
  escalation, exposure of funds or account data).
- The affected component — `go/backend`, `go/pacioli`, one of the
  `typescript/` apps, the Helm charts, or the CI/CD pipeline.
- Full paths of the source files involved, and the commit or release tag you
  tested against.
- Step-by-step reproduction instructions, and a proof of concept if you have
  one.
- The impact you believe an attacker could achieve.

If you are reporting an issue in a third-party dependency, please tell us
whether you have already notified that project.

## What to expect

- We aim to acknowledge a report within **3 working days**.
- We will confirm whether we can reproduce the issue and give you an initial
  assessment of severity.
- We will keep you updated as we work on a fix, and we will let you know when
  it ships.
- We will credit you in the advisory unless you ask us not to.

Please give us a reasonable opportunity to release a fix before disclosing the
issue publicly.

## Scope

In scope:

- This repository — the wallet backend, the Protea / Botanist / Hortus
  frontends, the Helm charts, and the CI/CD workflows.
- Deployed Interledger wallet environments, **only** where you are testing
  against your own account and data.

Out of scope:

- Denial of service, volumetric, or resource-exhaustion testing against any
  deployed environment.
- Social engineering of Interledger Foundation staff or contributors.
- Reports generated solely by automated scanners with no demonstrated impact.
- Findings in the `local/` development stack or the `go/mock/*` services.
  These exist only for local development and testing and deliberately ship
  with well-known, non-production credentials.
- Missing hardening headers or configuration issues with no demonstrated
  exploit path.

Do not access, modify, or exfiltrate data belonging to other users. If you
encounter third-party data during testing, stop and tell us immediately.

## Supported versions

Security fixes land on `main` and are released from there. Where a maintenance
branch (`<major>.<minor>.x`) is active, backports are made at the maintainers'
discretion. Older releases are not supported — please upgrade to the latest
release before reporting an issue.

## Reporting non-security bugs

For anything that is not a security issue, please open a regular
[GitHub issue](https://github.com/interledger/interledger-app/issues). For
conduct concerns, see [code_of_conduct.md](.github/code_of_conduct.md).
