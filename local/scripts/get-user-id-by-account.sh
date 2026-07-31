#!/usr/bin/env bash
# get-user-id-by-account.sh — resolve the Kratos user ID from a linked_account UUID.
#
# Take the account ID straight from the /accounts/<id> URL.
#
# Usage:
#   ./get-user-id-by-account.sh <linked-account-uuid>
#
# Example:
#   ./get-user-id-by-account.sh 3f1a2b4c-dead-beef-0000-123456789abc

set -euo pipefail

ACCOUNT_ID="${1:-}"

if [[ -z "$ACCOUNT_ID" ]]; then
  echo "Usage: $0 <linked-account-uuid>" >&2
  exit 1
fi

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"
PGDATABASE="${PGDATABASE:-backend}"

export PGPASSWORD

psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "
  SELECT uw.user_id
  FROM linked_accounts la
  JOIN user_wallets uw ON uw.wallet_id = la.wallet_id
  WHERE la.id = '$ACCOUNT_ID';
"
