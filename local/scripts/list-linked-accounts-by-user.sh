#!/usr/bin/env bash
# list-linked-accounts-by-user.sh — list all linked_accounts for a Kratos user ID.
#
# Use get-user-id-by-account.sh first if you only have the account ID from the URL.
#
# Usage:
#   ./list-linked-accounts-by-user.sh <kratos-user-uuid>
#
# Example:
#   ./list-linked-accounts-by-user.sh d96254b2-9bef-41f5-90b6-535b9c94f5f1

set -euo pipefail

USER_ID="${1:-}"

if [[ -z "$USER_ID" ]]; then
  echo "Usage: $0 <kratos-user-uuid>" >&2
  exit 1
fi

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"
PGDATABASE="${PGDATABASE:-backend}"

export PGPASSWORD

psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "
  SELECT
    la.id,
    la.name,
    la.mask,
    la.provider,
    la.provider_id,
    la.type,
    la.state,
    la.plaid_account_id,
    la.deleted_at
  FROM linked_accounts la
  JOIN user_wallets uw ON uw.wallet_id = la.wallet_id
  WHERE uw.user_id = '$USER_ID'
  ORDER BY la.deleted_at NULLS FIRST, la.name;
"
