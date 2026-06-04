# Argo CD Script Guide (One Pager)

This folder contains two Bash scripts used to refresh, sync, and verify Argo CD applications selected by label.

## Scripts

- `argocd_refresh_sync.sh`
	- End-to-end workflow:
		- fetch Argo applications
		- filter by label selector
		- refresh each app
		- sync each app
		- wait until each app is Healthy and Synced
- `argocd_select_applications.sh`
	- Selector utility used by the main script.
	- Reads Argo applications JSON and outputs matching app names.

## Prerequisites

- `bash`
- `curl`
- `jq`
- Cloudflare Access service token credentials (client ID + client secret)
- Argo CD API token for the target Argo instance

## Environment Variables

Required by `argocd_refresh_sync.sh`:

- `ARGOCD_ENDPOINT`
- `ARGOCD_AUTH_TOKEN`
- `CF_ACCESS_CLIENT_ID`
- `CF_ACCESS_CLIENT_SECRET`
- `APPLICATION_SELECTOR`

Optional:

- `TIMEOUT_SECONDS` (default: `1800`)
- `POLL_INTERVAL_SECONDS` (default: `20`)
- `ARGOCD_HTTP_CLIENT` (default: `curl`, used mainly for tests)

## Local Usage

Use inline env vars:

```bash
cd ci/scripts
APPLICATION_SELECTOR='environment=wallet-dev1' \
TIMEOUT_SECONDS=300 \
POLL_INTERVAL_SECONDS=10 \
ARGOCD_ENDPOINT='https://sandbox-argo.interledger.tech' \
ARGOCD_AUTH_TOKEN='...' \
CF_ACCESS_CLIENT_ID='...' \
CF_ACCESS_CLIENT_SECRET='...' \
./argocd_refresh_sync.sh
```

Use an env file:

```bash
cd ci/scripts
APPLICATION_SELECTOR='environment=wallet-dev1' \
TIMEOUT_SECONDS=300 \
POLL_INTERVAL_SECONDS=10 \
./argocd_refresh_sync.sh --env-file .env
```

Notes:

- Inline env vars override values loaded from `--env-file`.
- The script masks `ARGOCD_AUTH_TOKEN` and `CF_ACCESS_CLIENT_SECRET` in output.

## Selector Syntax

Supported selector parts:

- equality: `key=value`
- inequality: `key!=value`

Multiple parts are comma-separated and use AND semantics.

Examples:

- `environment=wallet-dev1`
- `environment=wallet-dev1,team=wallet`
- `environment!=wallet-sandbox`

## Quick curl Checks (Before Full Script)

If the script fails early, verify access first:

```bash
curl -sS -L \
	-H "CF-Access-Client-Id: ${CF_ACCESS_CLIENT_ID}" \
	-H "CF-Access-Client-Secret: ${CF_ACCESS_CLIENT_SECRET}" \
	-H "Authorization: Bearer ${ARGOCD_AUTH_TOKEN}" \
	"${ARGOCD_ENDPOINT%/}/api/v1/applications" \
	-w "\nstatus=%{http_code} type=%{content_type} url=%{url_effective}\n"
```

Expected:

- JSON response from Argo CD.

Common failure patterns:

- HTML page with "Sign in ・ Cloudflare Access": Cloudflare policy/token issue.
- `401` JSON with Argo error (for example invalid signature): wrong Argo token for that instance.

## Testing

Run tests from repo root:

```bash
ci/tests/argocd_select_applications_test.sh
ci/tests/argocd_refresh_sync_test.sh
```

What they cover:

- selector parsing and matching behavior
- full refresh/sync flow via mocked HTTP client
- no-match failure behavior
- env-file loading behavior

## Troubleshooting Checklist

1. Confirm endpoint matches environment (`sandbox-argo` vs `development-argo`).
2. Confirm Cloudflare service token is authorized for that hostname and route.
3. Confirm Argo token is valid for that Argo instance.
4. Verify selector matches existing labels.
5. Retry with a smaller selector first (for example one environment label).

## Security Notes

- Do not commit real tokens/secrets.
- Treat any secret shown in terminal/chat as exposed and rotate it.

