# Migration Email Job

Sends a one-off announcement email — typically a migration notice — to all users, to a
region, or to a handful of addresses for testing. The job is triggered by hand; nothing
schedules it.

Code: `go/backend/jobs/2026_08_send_migration_email.go`.

---

## Triggering it

Start the workflow from the Temporal admin portal
(locally `https://temporal.mgnt.interledger.test`):

| Field | Value |
|-------|-------|
| Workflow Type | `SendMigrationEmailJob` |
| Task Queue | `backend` |
| Input | the JSON below |

```json
{
  "subject": "We are moving your account to a new provider",
  "paragraphs": [
    { "heading": "What is changing" },
    { "paragraph": "From 1 September your wallet moves to a new payment provider." },
    { "paragraph": "Your balance and wallet address stay the same." }
  ],
  "region": "EU"
}
```

**Always do a test send first** — same input, but with `email` set to your own address
instead of `region`. There is no other way to preview the email.

Check that the environment has `email.enabled` set. With it false the backend uses the
noop email client: the job logs `NOT SENDING: migration email` and completes with no
failures, which looks exactly like a successful run.

---

## Targeting

Set **either** `region` **or** `email`. If `email` is set, `region` is ignored.

| Field | Value | Sends to |
|-------|-------|----------|
| `region` | `"ALL"` | every user |
| `region` | `"EU"` | all EU countries |
| `region` | ISO country code, e.g. `"US"`, `"ZA"` | that country |
| `email` | `"me@example.com"` | one user, for testing |
| `email` | `"a@x.com, b@y.com"` | those users — used to retry a failed run |

Region comes from the user's `countryCode` in Kratos, so it is where the user signed up,
not where they are now. Every address in `email` must match a user: if one does not, the
job fails without sending anything, so a typo cannot silently skip someone.

---

## What the user receives

- **Subject** — your `subject`, used as the email subject and passed to the template.
- **Greeting** — `Hello <first name>,` is prepended automatically. Do not write your own.
- **Body** — your `paragraphs`, in order. Supported blocks: `{"paragraph": "..."}`,
  `{"heading": "..."}`, `{"code": "..."}`, and `{"table": [{"label": "...", "text": "..."}]}`.
- **Button** — always **View account**, linking to the wallet login page. Not configurable.

---

## Failures and re-runs

The workflow returns the list of addresses whose send failed, for example:

```json
["alice@example.com", "bruno@example.com"]
```

Sends are **not retried**: SendGrid has no idempotency key, so retrying an ambiguous
timeout risks emailing someone twice. Nothing else records who was emailed — if the
result lists failures, re-run the job with those addresses in `email` and the same
subject and paragraphs.

An empty list means every recipient was sent to. A failure listing recipients (rather
than sending) fails the whole job before any email goes out.
