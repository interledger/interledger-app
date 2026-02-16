The correct way to run specific tests from `features/*.feature` is to tag the specific test appropriately and then use the -args and -tags flags together.
`go test -v -timeout 5m -args -tags @signuponly`

You can also target multiple tags using a && parameter to act as a better filter.
`go test -v -args -tags "@kyc && @xago"`

DO NOT SUPPRESS TEST OUTPUT EVER

## Troubleshooting
- Remember during tests users are unique, so database cleanup has very limited value if at all.
- Keep in mind that from the e2e test perspecive we should use the public url mode to access iframes. For example `mockxago` is available at `mockxago.interledger.test`
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
- **MockXago Persona KYC Implementation** (added Feb 16, 2026)
  + MockXago now serves a Persona-like KYC iframe at `/kyc/iframe` endpoint
  + This enables South African (Xago) users to complete KYC verification in tests
  + The iframe accepts form submission at `/kyc/submit` and sends webhook notification to backend
  + Backend receives `id.verification.accepted` webhook and triggers Xago sub-account creation
  + Test step: `I fill and submit the mockxago KYC iframe` handles form filling and submission
  + **How it works**:
    1. Frontend redirects South Africa users to Persona KYC (backend routes via `GetKYCProviderWidget`)
    2. Persona widget points to MockXago's `/kyc/iframe` endpoint  
    3. User fills verification form (name, address, DOB, etc.)
    4. Form submission sends POST to `/kyc/submit` with multipart form data
    5. MockXago saves sub-account details and sends webhook to backend
    6. Backend receives webhook and starts Temporal workflow for Xago onboarding
    7. Eventually triggers creation of Xago balance account and linked account setup


## Maintain
It is the job of the agent to add, update or remove relevant information to this file.