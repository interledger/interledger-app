The correct way to run specific tests from `features/*.feature` is to tag the specific test appropriately and then use the -args and -tags flags together.
```
# Run only the @signuponly tests
go test -v -timeout 5m -args -tags @signuponly 
# Run tests tagged with @withdrawal AND @fees
go test -v -timeout 10m -args -tags="@withdrawal&&@fees"
# Run tests without debug info
go test -v -timeout 10m -args -debug=false
```

DO NOT SUPPRESS TEST OUTPUT EVER

## KYC Flow Behavior
- The application does **NOT** log users out after KYC completion. The `personal-details` action:
  1. Removes the flow cookie entry (`exitFlow`)
  2. Calls `setKYCStatusPending` on the backend gRPC
  3. Redirects to `/` (the dashboard) via `redirectWithSnackbar`
- A redirect to `/login` during KYC would only happen if `setKYCStatusPending` returns a gRPC `Unauthenticated` error, which means the Kratos session was **already invalid** before KYC submitted — not caused by KYC itself. This is standard auth error handling in `ConnectError` (see `typescript/protea/app/lib/error.server.ts`).
- The `iWaitForTheKYCCompletion` step correctly detects this edge case and fails with a clear error. Any claim that "the wallet logs you out after KYC" is incorrect.

## Troubleshooting
- Remember during tests users are unique, so database cleanup has very limited value if at all.
- You can investigate temporal jobs by executin temporal cli commands within the temporal container
- We spent a long time chasing Kratos `format: "tel"` validation failures that appeared to reject valid phone numbers. After cleaning the environment with `make reset` in `local`, the issue disappeared. The root cause is still unknown, so keep this in mind if `tel` errors resurface after environment changes.
- Important details about phone number troubleshooting
  + Keep in mind that the tests aim to generate randomised phone numbers so they are not supposed tobe duplicate
  + We've confirmed that the correct format is +49987654321
  + Most issues relate to kratos validation, either the phone number already exist or format is wrong
  + When starting up the environment then use `make all-nowatch`
  + The `/withdraw` page loads a MockGatehub iframe widget similar to deposits
- `iShouldSeeMyAccountBalanceWith` calls `getUserIDFromSignup` for diagnostics. If it logs `User NOT found in signups table: ...converting NULL to string...`, it means the signup row exists but `user_id` is NULL — a race condition where the Kratos user hasn't been linked yet. This is informational only; the actual balance check runs on the page.
- `WaitForLoadState(networkidle)` without a timeout will block indefinitely if the page never settles. Always pass `Timeout: playwright.Float(10000)` when calling it inside retry loops.

## Maintain
It is the job of the agent to add, update or remove relevant information to this file.