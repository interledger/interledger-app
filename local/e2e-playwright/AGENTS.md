The correct way to run specific tests from `features/*.feature` is to tag the specific test appropriately and then use the -args and -tags flags together.
```
go test -v -timeout 5m -args -tags @signuponly 
```

DO NOT SUPPRESS TEST OUTPUT EVER

## Troubleshooting
- Remember during tests users are unique, so database cleanup has very limited value if at all.
- We spent a long time chasing Kratos `format: "tel"` validation failures that appeared to reject valid phone numbers. After cleaning the environment with `make reset` in `local`, the issue disappeared. The root cause is still unknown, so keep this in mind if `tel` errors resurface after environment changes.
- Important details about phone number troubleshooting
  + Keep in mind that the tests aim to generate randomised phone numbers so they are not supposed tobe duplicate
  + We've confirmed that the correct format is +49987654321
  + Most issues relate to kratos validation, either the phone number already exist or format is wrong
- When starting up the environment then use `make all-nowatch`
- **Wallet address form submission flakiness** (observed Feb 3, 2026)
  + The `@kyc` test sometimes fails with "wallet does not appear to be in 'Reserved' state, still on wallet-address"
  + This means the wallet-address form submission failed and the page never redirected away from `/wallet-address`
  + **Debugging steps when this happens**:
    1. Check backend logs: `docker compose logs backend | grep -i "wallet\|error"`
    2. Check if wallet was created in DB despite UI not redirecting: `docker compose exec -T postgres psql -U postgres -d backend -c "SELECT id, name, wallet_address, created_at FROM wallets ORDER BY created_at DESC LIMIT 5;"`
    3. Look for validation errors (Rafiki asset lookup, duplicate address, etc.)
    4. Check middleware logs for timing of wallet creation vs form submission
  + **Root cause speculation**: Middleware may be creating a default wallet between form render and submit, causing a duplicate wallet conflict that blocks form submission
  + **Mitigation**: DB poll helper added to wait for stable wallet count, but form submission failure is a separate issue that needs frontend/backend error surfacing
- **P2P payment deposit balance verification** (observed Feb 4, 2026)
  + Deposits via mockgatehub iframe complete successfully (webhook delivered, Temporal workflow processes)
  + BUT: Balance does not appear on frontend dashboard even after multiple page reloads
  + User context is correct (verified via logs: `Current user context: sender (email: XXXXXX-sender-p2p@example.com)`)
  + **Debugging steps**:
    1. Check mockgatehub logs to verify deposit webhook was sent and received
    2. Check backend logs for Temporal workflow completion
    3. Verify database has the deposit transaction recorded
    4. Check if balance is being returned by backend API but not rendered by frontend
    5. Look at screenshot `after-deposit.png` to see actual page state
  + **Possible root causes**:
    - Frontend not re-fetching balance data after deposit
    - Balance API endpoint returning wrong user's data
    - Temporal workflow completing but not updating correct user record
    - User UUID mismatch between mockgatehub deposit and backend user


## Maintain
It is the job of the agent to add, update or remove relevant information to this file.