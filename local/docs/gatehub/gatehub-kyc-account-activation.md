# Gatehub-linked Account Activation

## Trigger and preconditions
- The wallet UI "Activate" entry point calls the backend gRPC `GetGatehubOnboardingWidget`, which requires an authenticated user, a wallet in context, and an EU country; non-EU wallets receive a failed precondition response so activation cannot start ([go/backend/grpc/gatehub.go](go/backend/grpc/gatehub.go#L10-L117)).
- If a Gatehub user does not exist for the wallet, the backend **starts a Temporal workflow** to create one on-demand before returning the onboarding widget URL.
  - The workflow ID is deterministic: `gatehub_create_user_{walletID}`. Each wallet has one logical workflow, reused across multiple Activate clicks.
  - The workflow runs activities with exponential backoff retries; if activities fail, the workflow automatically retries before surfacing errors to the caller.
- GateHub docs note that **connecting a managed user to a gateway should only happen after onboarding/KYC completes** (Users → “Connecting to Gateway”). This matches the backend flow, which waits for KYC verification webhooks before activating the wallet.

## Gatehub user creation workflow
- `CreateGatehubUserWorkflow` runs the following steps ([go/backend/providers/gatehub/ops/workflows.go](go/backend/providers/gatehub/ops/workflows.go#L22-L62)):
  - Lookup an existing external user mapping; if missing, call the Gatehub API to create a user using the wallet holder’s email ([go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L58-L94)).
  - Persist the `gatehub_users` mapping and ensure a Gatehub EUR linked account exists (reusing it if already present) ([go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L96-L157)).
  - Configure the ledger account for that linked account so balance movements can be posted ([go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L159-L183)).
- Workflow ID pattern: `gatehub_create_user_{walletID}` is fixed, so the same workflow ID is reused for a given wallet.
- Workflow ID policy: `TERMINATE_IF_RUNNING` means a failed or stalled creation can be retried by re-requesting the widget. If a workflow is already running, it completes before a new attempt launches. To force a fresh attempt after debugging, manually terminate the workflow (see Temporal workflow management below).

## What the user sees
- After the workflow completes, the backend returns the Gatehub onboarding widget URL. The frontend embeds it so the user can submit KYC directly to Gatehub.
- Deposit/withdraw widgets follow the same prerequisite lookup/creation, but use the on/off-ramp widget API instead ([go/backend/grpc/gatehub.go](go/backend/grpc/gatehub.go#L37-L91)).
- The GateHub onboarding iframe sends postMessage events. The public docs define `OnboardingCompleted` with a **string** payload of `'submitted' | 'resubmitted'` (Appendix → Message events). This differs from some internal mocks that wrap the payload as JSON.

## Activation signal (post-KYC)
- Gatehub sends `id.verification.*` webhooks. `HandleUserVerificationWebhook` maps the webhook `user_uuid` back to the wallet; if it cannot, the handler returns 500 and no activation occurs ([go/backend/providers/gatehub/ops/webhooks.go](go/backend/providers/gatehub/ops/webhooks.go#L259-L295)).
- The backend expects the verification payload to include `data.gateway` (containing “paywiser”) and `data.verified.short` with `accepted`/`rejected`. This structure is enforced in `HandleUserVerificationWebhook` and is **not** currently documented in the public GateHub docs.
- On `accepted`, the backend starts `BackfillAccountAndSetKYC` ([go/backend/providers/gatehub/ops/webhooks.go](go/backend/providers/gatehub/ops/webhooks.go#L281-L294)), which triggers `BackfillAccountWorkflow` to:
  - Optionally move funds from a configured Gatehub sender into the user’s EUR balance (only when `GATEHUB_SENDING_USER_ID` is set) and mark the backfill ([go/backend/providers/gatehub/ops/workflows.go](go/backend/providers/gatehub/ops/workflows.go#L335-L376)).
  - Set the wallet’s KYC status to `Level1`, which is the application’s indicator that the Gatehub-backed account is activated ([go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L611-L618)).
- On `rejected`, the handler sets the wallet KYC status to `Denied`; the user remains inactive ([go/backend/providers/gatehub/ops/webhooks.go](go/backend/providers/gatehub/ops/webhooks.go#L282-L286)).

## Failure modes and whether activation proceeds
- Gatehub user creation failure (API errors, missing wallet users, etc.) surfaces as an error from `GetGatehubOnboardingWidget`, so the widget URL is not returned and activation cannot start. No `gatehub_users` row is written, so later webhooks will also fail to map and will 500.
- If the webhook arrives but mapping to a wallet fails (no `gatehub_users` entry or mismatched UUID), the handler logs and returns 500 before setting KYC, so activation does not occur ([go/backend/providers/gatehub/ops/webhooks.go](go/backend/providers/gatehub/ops/webhooks.go#L268-L279)).
- If the GateHub “get wallets for user” response does not include a wallet marked `primary`, `CreateGatehubWalletLinkedAccount` fails with “Could not find a primary wallet for gatehub user”, and the create-user workflow retries indefinitely ([go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L110-L136)).
- If the onboarding iframe posts an unexpected event payload shape (e.g., not the `'submitted' | 'resubmitted'` string documented by GateHub), the frontend may not call `/personal-details` to advance the flow, leaving KYC in `Pending`.
- Temporal workflows are idempotent by workflow ID; a subsequent activation attempt reuses the same IDs and resumes or retries the steps rather than duplicating accounts.

## Temporal workflow management and debugging

When debugging activation issues, use the Temporal CLI to inspect and manage workflows:

**List all workflows:**
```bash
docker compose exec -T temporal temporal workflow list --namespace default
```

**Inspect a specific workflow (by wallet ID):**
```bash
# Workflow ID pattern: gatehub_create_user_{walletID}
docker compose exec -T temporal temporal workflow describe --namespace default --workflow-id gatehub_create_user_{walletID}
```

**Show workflow execution history (activity logs, failures, retries):**
```bash
docker compose exec -T temporal temporal workflow show --namespace default --workflow-id gatehub_create_user_{walletID}
```

**Terminate a stuck workflow (e.g., after code fixes or debugging):**
```bash
docker compose exec -T temporal temporal workflow terminate --namespace default --workflow-id gatehub_create_user_{walletID}
```
After termination, re-clicking Activate will start a fresh workflow with the same ID.

**Common scenarios:**
- Workflow stuck in retry loop → Check activity error messages in `workflow show` output. Common causes: API call failures (e.g., 401 Unauthorized), timeout issues. Fix the underlying issue, terminate the workflow, then retry activation.
- Multiple activation attempts → Check `workflow list` for workflows in RUNNING state. If a workflow is already running from a previous attempt, the next Activate click will wait for it to complete (due to `TERMINATE_IF_RUNNING` policy).
- Stale `gatehub_users` mapping → If a workflow successfully created a user but later fails (e.g., webhook never arrives), the mapping persists in the database. Re-clicking Activate will skip the create-user step and return the existing widget URL, but the mapping is now stale if the Gatehub user changed.
1. Frontend requests `GetGatehubOnboardingWidget` → backend checks EU precondition and runs the create-user workflow if needed ([go/backend/grpc/gatehub.go](go/backend/grpc/gatehub.go#L10-L35), [go/backend/providers/gatehub/ops/workflows.go](go/backend/providers/gatehub/ops/workflows.go#L22-L62)).
2. Backend returns widget URL → frontend renders Gatehub onboarding iframe where the user completes KYC.
3. Gatehub posts `id.verification.accepted` webhook → backend runs `BackfillAccountAndSetKYC`/`BackfillAccountWorkflow`, optionally backfills funds, and sets KYC status to Level1 ([go/backend/providers/gatehub/ops/webhooks.go](go/backend/providers/gatehub/ops/webhooks.go#L281-L294), [go/backend/providers/gatehub/ops/workflows.go](go/backend/providers/gatehub/ops/workflows.go#L335-L376), [go/backend/providers/gatehub/ops/activity.go](go/backend/providers/gatehub/ops/activity.go#L611-L618)).
4. The KYC status update is the flag other subsystems read to treat the wallet as activated; until that status is updated, features that require activation remain unavailable.

## Debugging activation failures

When activation fails (e.g., `GetGatehubOnboardingWidget` returns an error), follow this checklist:

**Step 1: Check the Temporal workflow**
```bash
docker compose exec -T temporal temporal workflow show --namespace default --workflow-id gatehub_create_user_{walletID}
```
Look at the activity history. If an activity shows `FAILED` with a specific error (e.g., `401 Unauthorized`), that's your root cause.

**Step 2: Check backend logs for API details**
Backend logs the exact URL, method, and payload for Gatehub API calls:
```bash
docker compose logs backend --since 5m | grep -E "Gatehub signing|Gatehub CreateUser"
```
This reveals which endpoint is being called and with what data. Common issues:
- URL pointing to production instead of mockgatehub (env override not applied)
- Payload missing required fields
- Timestamp format mismatches (backend sends milliseconds; Gatehub may expect seconds)

**Step 3: Check mockgatehub request logs**
If debugging against mockgatehub, check whether requests reach the service:
```bash
docker compose logs mockgatehub --since 5m | grep -v health
```
If no POST requests appear, the backend isn't calling mockgatehub (likely pointing to production or a misconfigured URL).

**Step 3b: Verify mockgatehub response shape**
If you are using mockgatehub locally, ensure:
- `GET /core/v1/users/:userUuid` returns a wallet with `primary: true` (required by `CreateGatehubWalletLinkedAccount`).
- `id.verification.accepted` webhook payload includes `data.gateway` and `data.verified.short` so the backend can map and accept the verification.

**Step 4: Verify the `gatehub_users` database table**
After a workflow completes (successfully or after termination), check whether the mapping was created:
```bash
docker compose exec -T backend psql -h postgres -U wallet -d wallet_backend -c "SELECT wallet_id, gatehub_user_id FROM gatehub_users WHERE wallet_id = '{walletID}';"
```
If the row exists but later webhooks fail, the mapping is stale and the workflow needs re-running.

**Step 5: Force a retry after fixes**
Once you've identified and fixed the issue:
```bash
# Terminate the stuck workflow
docker compose exec -T temporal temporal workflow terminate --namespace default --workflow-id gatehub_create_user_{walletID}

# Optionally clean the database mapping (if needed to force fresh creation)
docker compose exec -T backend psql -h postgres -U wallet -d wallet_backend -c "DELETE FROM gatehub_users WHERE wallet_id = '{walletID}';"

# Re-click Activate in the UI to start a fresh workflow
```
