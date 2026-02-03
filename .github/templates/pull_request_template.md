# Pull Request Checklist

## Functional Completion

- [ ] Acceptance criteria are fully implemented and verified
- [ ] Feature works as intended across supported regions and currencies
- [ ] Edge cases and error states are handled (e.g. network failures, retries)
- [ ] No known critical or high-severity defects remain open

## Security & Compliance

- [ ] No new security vulnerabilities introduced (dependency + code scanning)
- [ ] All data handling complies with KYC and privacy requirements
- [ ] Sensitive data is:
  - [ ] Not logged
  - [ ] Properly encrypted at rest and in transit
- [ ] Compliance-relevant flows (KYC, account locking, payments) follow approved rules
- [ ] Any compliance assumptions are reviewed or explicitly accepted

## Testing & Verification

- [ ] Automated tests added or updated:
  - [ ] Unit tests for business logic
  - [ ] Integration tests for wallet, payments, and KYC boundaries
- [ ] All tests pass in CI
- [ ] Manual verification completed where automation is insufficient (documented)

## Observability & Operability

- [ ] Relevant logs, metrics, and alerts are in place
- [ ] Failures are detectable and actionable
- [ ] User-visible errors have clear, non-leaking messages
- [ ] Rollback or mitigation strategy is defined (for high risk level PRs)

## Documentation

- [ ] Technical documentation updated (architecture, flows, assumptions)
- [ ] User-facing or support-relevant changes documented
- [ ] RunBooks or escalation notes added for:
  - [ ] KYC issues
  - [ ] Payment failures
  - [ ] Account restrictions

## Deployment & Release

- [ ] Change is deployable via the standard pipeline
- [ ] Rollout strategy defined (if needed)
- [ ] No manual production steps required (or explicitly documented)
- [ ] Release notes updated if user-visible behavior changes

## Ownership & Handoff

- [ ] Clear owner for post-release monitoring
- [ ] Known limitations or follow-ups are tracked
- [ ] Support and compliance teams / providers are informed of relevant changes

## Regulatory Readiness (when applicable)

- [ ] Actions are auditable
- [ ] Decisions are traceable (who / what / when)
- [ ] No shortcuts taken that would block an audit or incident review
