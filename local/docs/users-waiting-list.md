# User Waiting List - How It Works

## Overview

The Interledger Wallet implements a **country-based waiting list system** that restricts signups to specific supported countries. When a user attempts to sign up from an unsupported country, they are automatically redirected to join a waiting list instead of completing the registration flow.

This document explains how the waiting list works, how users are approved, and where the configuration is managed.

---

## When Users Are Placed on the Waiting List

### Supported Countries Check

During the signup flow, after a user enters their personal details (name, email, country), the system checks if their country is in the **supported countries list**.

**Code Reference**: [protea/app/routes/signup/route.tsx](../../typescript/protea/app/routes/signup/route.tsx#L164-L178)

```typescript
// In detailsAction() function
if (
  !(
    country == 'CA' ||
    country == 'US' ||
    country == 'ZA' ||
    isEUCountry(country)
  )
) {
  return redirect(
    `/waitlist?country=${country}&email=${email}&fullName=${firstName} ${lastName}`
  )
}
```

### Supported Countries (as of current implementation)

The following countries can proceed directly with signup:
- **CA** - Canada
- **US** - United States
- **ZA** - South Africa
- **EU Countries** - All European Union member states (checked via `isEUCountry()` helper)

### Unsupported Countries → Waitlist

If a user's country is **not** in the supported list, they are:
1. Automatically redirected to `/waitlist` with their details pre-filled
2. Shown a message: *"Leave your details below and we will notify you as soon as enrollment opens."*
3. Required to submit their information to join the waiting list

**User Experience**:
- User fills out signup form with their country (e.g., Mexico, Brazil, India)
- Clicks "Next" or "Continue"
- System detects unsupported country
- User is redirected to waitlist page with pre-filled data
- User sees a friendly message explaining they're on a waiting list

---

## Waitlist Database Schema

### Table: `waitlist_signups`

**Location**: [go/backend/db/schema.hcl](../../go/backend/db/schema.hcl)

```hcl
table "waitlist_signups" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    null = false
    type = text
  }
  column "country_code" {
    null = false
    type = character(2)
  }
  column "full_name" {
    null = false
    type = text
  }
  column "beta_opt_in" {
    null = false
    type = bool
  }
  column "can_signup" {
    null = false
    type = bool
    default = false  # Users start on waitlist (false)
  }
  column "mug_id" {
    null = true
    type = text
  }
  column "user_id" {
    null = true
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamp
    default = sql("now()::TIMESTAMP")
  }
}
```

### Key Fields

- **`id`**: Unique identifier for the waitlist signup
- **`email`**: User's email address
- **`country_code`**: 2-letter ISO country code
- **`full_name`**: User's full name
- **`beta_opt_in`**: Whether user opted into beta testing
- **`can_signup`**: **Critical field** - determines if user is approved to sign up
  - `false` (default): User is on waitlist
  - `true`: User has been approved by admin
- **`mug_id`**: Optional - for special promotional mugs with unique wallet addresses
- **`user_id`**: Set after user completes signup (links to Kratos identity)

---

## How to Approve Users (Admin Portal)

### Admin Portal - Waitlist Management

**Access**: https://admin.interledger.app/waitlist (or local admin portal)

**Admin View**: [botanist/app/routes/waitlist.tsx](../../typescript/botanist/app/routes/waitlist.tsx)

### Steps to Approve a User

1. **Log into Admin Portal** (Botanist)
   - Navigate to `/waitlist` route
   - See a table of all waitlist signups

2. **View Waitlist Table**
   
   The table shows:
   - Full Name
   - Email
   - Country Code
   - Beta Opt-In status (✓ or ✗)
   - Mug ID (if applicable)
   - **Action column**: "Approve" button or "✓ (Approved)" status

3. **Click "Approve" Button**
   
   - Click the "Approve" button next to the user's entry
   - This triggers the `AllowWaitlistSignup` RPC call
   - Backend updates `can_signup = true` in the database
   - The button changes to show a checkmark (✓) indicating approval

4. **Copy Signup Link**
   
   After approval, clicking the checkmark icon copies a special signup URL to clipboard:
   ```
   https://interledger.app/signup?waitlistSignupId=<USER_ID>
   ```

5. **Send Link to User**
   
   - Manually email or communicate the signup link to the approved user
   - User clicks the link and can now complete signup
   - **Note**: There is currently **no automated email notification** - admin must manually notify the user

### Backend Flow

**Admin Action Flow**:

1. Admin clicks "Approve" button
2. Frontend calls `allowWaitlistSignup` RPC with user's `id`
3. Backend updates database:
   ```sql
   UPDATE waitlist_signups 
   SET can_signup = true 
   WHERE id = '<user_id>';
   ```
4. User can now proceed with signup when they visit the special link

**Code References**:
- Admin Frontend: [botanist/app/routes/waitlist.tsx](../../typescript/botanist/app/routes/waitlist.tsx)
- Admin Backend: [go/backend/admin/waitlist.go](../../go/backend/admin/waitlist.go)
- Database Logic: [go/backend/waitlist/ops/ops.go](../../go/backend/waitlist/ops/ops.go)

---

## Special Feature: Mug IDs

### Promotional Mugs with Wallet Addresses

The system supports a special promotional feature where physical mugs have unique wallet addresses associated with them.

**User Flow**:
1. User receives a physical Interledger Wallet mug with a unique code/QR
2. User visits `/waitlist?mug=<MUG_ID>`
3. System checks if mug is available (not already claimed)
4. If available, user can join waitlist and claim the mug's wallet address
5. Admin approves user
6. User completes signup and inherits the mug's wallet address

**Mug Availability Check**:
- Backend checks if `mug_id` already exists in `waitlist_signups`
- If not found, mug is available
- If found, mug is already claimed

**Database Behavior**:
```sql
INSERT INTO waitlist_signups (email, country_code, full_name, beta_opt_in, mug_id) 
VALUES ($1, $2, $3, $4, $5) 
ON CONFLICT (email, country_code) 
DO UPDATE SET mug_id = excluded.mug_id 
WHERE waitlist_signups.mug_id IS NULL;
```
This allows users to update their waitlist entry with a mug ID if they join the waitlist again with a mug code.

---

## Configuration & Code Locations

### Where Supported Countries Are Defined

**File**: `typescript/protea/app/routes/signup/route.tsx`

**Function**: `detailsAction()` - lines ~164-178

```typescript
if (
  !(
    country == 'CA' ||
    country == 'US' ||
    country == 'ZA' ||
    isEUCountry(country)
  )
) {
  return redirect(`/waitlist?...`)
}
```

**To Add a New Supported Country**:
1. Edit the condition in `detailsAction()` function
2. Add new country code to the list (e.g., `country == 'MX'`)
3. Deploy updated frontend

**Example**:
```typescript
if (
  !(
    country == 'CA' ||
    country == 'US' ||
    country == 'ZA' ||
    country == 'MX' ||  // Added Mexico
    isEUCountry(country)
  )
) {
  return redirect(`/waitlist?...`)
}
```

### Backend API Endpoints

**Protea (User-Facing)**:
- `POST /waitlist` - Join waitlist (form submission)
- Backend RPC: `JoinWaitlist` - [go/backend/grpc/waitlist.go](../../go/backend/grpc/waitlist.go)

**Botanist (Admin Portal)**:
- `GET /waitlist` - View all waitlist signups
- `POST /waitlist` - Approve a user (toggle `can_signup`)
- Backend RPC: `ListWaitlistSignups`, `AllowWaitlistSignup` - [go/backend/admin/waitlist.go](../../go/backend/admin/waitlist.go)

### Database Operations

**Location**: `go/backend/waitlist/ops/ops.go`

Key functions:
- `AddSignup()` - Create waitlist entry
- `CanSignup()` - Check if user is approved
- `AllowSignupById()` - Approve user (set `can_signup = true`)
- `SetSignupComplete()` - Mark signup as completed (set `user_id`)
- `ListSignups()` - Fetch all waitlist entries for admin
- `IsMugAvailable()` - Check if mug ID is unclaimed

---

## Common Scenarios

### Scenario 1: User from Unsupported Country Tries to Sign Up

1. User navigates to `/signup`
2. Fills in: Name: "Maria Garcia", Email: "maria@example.com", Country: "Mexico (MX)"
3. Clicks "Next"
4. System detects MX is not supported
5. User is redirected to `/waitlist?country=MX&email=maria@example.com&fullName=Maria Garcia`
6. User sees: *"Leave your details below and we will notify you as soon as enrollment opens."*
7. User submits waitlist form
8. Record created in `waitlist_signups` with `can_signup = false`

### Scenario 2: Admin Approves a Waitlist User

1. Admin logs into Botanist admin portal
2. Navigates to `/waitlist`
3. Sees Maria Garcia in the list with "Approve" button
4. Clicks "Approve"
5. Backend updates: `UPDATE waitlist_signups SET can_signup = true WHERE id = '<maria_id>'`
6. Admin copies signup link: `https://interledger.app/signup?waitlistSignupId=<maria_id>`
7. Admin emails Maria: "Your account has been approved! Click here to complete signup: [link]"

### Scenario 3: Approved User Completes Signup

1. Maria receives approval email with signup link
2. Clicks link → redirected to `/signup?waitlistSignupId=<id>`
3. Backend checks: `CanSignup(id)` → returns `true`
4. Maria proceeds through normal signup flow (phone verification, password)
5. On completion, backend calls: `SetSignupComplete(id, user_id)`
6. Record updated: `UPDATE waitlist_signups SET user_id = '<kratos_id>' WHERE id = '<maria_id>'`
7. Maria now has a full account

### Scenario 4: User Tries to Signup Without Approval

1. Unapproved user somehow obtains or guesses a waitlist signup ID
2. Visits `/signup?waitlistSignupId=<id>`
3. Backend checks: `CanSignup(id)` → returns `false` (still on waitlist)
4. Signup flow blocks or shows error: "Your account is not yet approved."

---

## Manual Database Queries

### Check a user's waitlist status

```sql
SELECT 
  id, 
  full_name, 
  email, 
  country_code, 
  can_signup, 
  user_id, 
  created_at
FROM waitlist_signups
WHERE email = 'user@example.com';
```

### Manually approve a user

```sql
UPDATE waitlist_signups 
SET can_signup = true 
WHERE email = 'user@example.com' 
AND country_code = 'MX';
```

### Find all unapproved users from a specific country

```sql
SELECT 
  id, 
  full_name, 
  email, 
  created_at
FROM waitlist_signups
WHERE country_code = 'MX' 
AND can_signup = false
AND user_id IS NULL
ORDER BY created_at DESC;
```

### Find all approved but not yet signed up users

```sql
SELECT 
  id, 
  full_name, 
  email, 
  country_code, 
  created_at
FROM waitlist_signups
WHERE can_signup = true 
AND user_id IS NULL
ORDER BY created_at DESC;
```

### Check if a mug ID is available

```sql
SELECT id, full_name, email 
FROM waitlist_signups 
WHERE mug_id = 'UNIQUE_MUG_CODE';
-- If no results, mug is available
-- If results, mug is claimed
```

---

## Future Enhancements

Current limitations and potential improvements:

1. **No Automated Email Notifications**
   - Admin must manually email the signup link to approved users
   - Could integrate with email service to auto-notify on approval

2. **Hardcoded Country List**
   - Supported countries are hardcoded in TypeScript
   - Could move to database configuration for dynamic updates

3. **No Bulk Approval**
   - Admin must approve users one-by-one
   - Could add bulk approval feature in admin portal

4. **No Waitlist Position**
   - Users don't know their position in queue
   - Could add waitlist position display

5. **No Auto-Approval Rules**
   - All approvals require manual admin action
   - Could implement auto-approval based on criteria (country opens, quota, etc.)

6. **No Country Expansion Notification**
   - When a country becomes supported, waitlisted users from that country aren't auto-notified
   - Could implement country expansion notification workflow

---

## Related Documentation

- [Email Verification Troubleshooting](email-verification-troubleshooting.md)
- [GateHub KYC Account Activation](gatehub-kyc-account-activation.md)

---

## Troubleshooting

### User claims they were redirected to waitlist but are from supported country

**Check**:
1. Verify country code in database: `SELECT * FROM waitlist_signups WHERE email = 'user@example.com';`
2. Ensure country code matches expected format (2-letter ISO code)
3. Check if EU country is correctly detected by `isEUCountry()` helper
4. Review signup flow logs for any errors

### User can't complete signup after approval

**Check**:
1. Verify `can_signup = true`: `SELECT can_signup FROM waitlist_signups WHERE id = '<id>';`
2. Ensure correct signup link format: `https://interledger.app/signup?waitlistSignupId=<ID>`
3. Check backend logs for `CanSignup()` RPC call results
4. Verify user hasn't already completed signup (`user_id` should be NULL)

### Admin portal doesn't show waitlist signups

**Check**:
1. Ensure admin is logged in with correct credentials
2. Check backend logs for `ListWaitlistSignups` RPC errors
3. Verify database connectivity
4. Run manual query: `SELECT COUNT(*) FROM waitlist_signups;`

### Mug ID shows as unavailable but shouldn't be

**Check**:
1. Query database: `SELECT * FROM waitlist_signups WHERE mug_id = 'CODE';`
2. Verify mug ID exists in allowed mugs list (code-defined)
3. Check for typos in mug ID entry
4. Ensure mug ID hasn't been claimed by another user

---

## Summary

The waiting list system is a **country-based gating mechanism** that:
- Automatically redirects users from unsupported countries to a waitlist
- Requires admin approval before users can complete signup
- Integrates with the main signup flow via special query parameters
- Supports promotional features like unique mug wallet addresses

**Admin Workflow**: View waitlist → Approve user → Copy signup link → Manually notify user

**Configuration**: Country support is defined in `protea/app/routes/signup/route.tsx` `detailsAction()` function
