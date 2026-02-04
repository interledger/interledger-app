# Kratos SDK Migration Task

## Objective
Migrate from direct HTTP `fetch()` calls to the official Ory Kratos TypeScript SDK (`@ory/kratos-client`) for all identity-related operations in the Protea application.

## Task Checklist

### Phase 1: Research & Documentation (Complete)
- [x] Identify all current Kratos HTTP call usages (frontend)
- [x] Research latest Ory Kratos SDK documentation
- [x] Document frontend usages in `old-usages.md`
- [x] Create SDK replacement mappings in `how-to-replace.md`
- [x] Document backend usages in `backend-usages.md`
- [/] Get user approval on migration plan

> **Note:** The Go backend already uses the Kratos Go SDK properly. Only the TypeScript frontend needs migration.

### Phase 2: Infrastructure Setup
- [x] Update `@ory/kratos-client` to latest version
- [x] Create SDK client wrapper in `app/lib/kratos-client.server.ts`
- [x] Update type imports to use latest SDK types

### Phase 3: Route Migrations
- [x] Migrate `login.tsx` (login flow)
- [ ] Migrate `signup/route.tsx` (registration flow)
- [ ] Migrate `logout.tsx` (logout flow)
- [ ] Migrate `recovery.tsx` (recovery flow)
- [ ] Migrate `recovery_.password.tsx` (settings during recovery)
- [ ] Migrate `verify.tsx` (verification flow)
- [ ] Migrate `settings_.password.tsx` (password settings)
- [ ] Migrate `settings_.phone.tsx` (phone settings)
- [ ] Migrate `login_.challenge.tsx` (session refresh)
- [ ] Migrate `totp_.challenge.tsx` (TOTP AAL2 login)
- [x] Migrate `totp_.two-factor-authentication.tsx` (TOTP setup)
- [ ] Migrate `otp_.challenge.tsx` (OTP challenge + settings init)
- [ ] Migrate API routes (`api.totp-challenge-init.tsx`, `api.totp-challenge-verify.tsx`, `api.check-totp-enabled.tsx`)
- [ ] Migrate `deposit/route.tsx` (session check)

### Phase 4: Core Library Updates
- [ ] Refactor `kratos.server.ts` helper functions
- [ ] Refactor `totp.server.ts` helper functions
- [ ] Update all imports across the application

### Phase 5: Verification
- [ ] Test all authentication flows
- [ ] Test all recovery flows
- [ ] Test all settings flows
- [ ] Test TOTP/2FA flows
- [ ] Verify error handling works correctly
